package server

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type TcpServer struct {
	address     string
	mutex       sync.Mutex
	connections map[string]net.Conn
	ctx         context.Context
	wg          *sync.WaitGroup
	errChan     chan error
	listener    *net.Listener
	msgHandler  func(net.Conn, []byte)
	config      *tls.Config
}

func NewTcpServer(ctx context.Context, serveAddress string) *TcpServer {
	server := &TcpServer{
		address:     serveAddress,
		ctx:         ctx,
		connections: make(map[string]net.Conn),
		errChan:     make(chan error),
		mutex:       sync.Mutex{},
		wg:          &sync.WaitGroup{},
		msgHandler:  func(net.Conn, []byte) {},
	}

	return server
}

func (server *TcpServer) Start() {

	log.Printf("INFO: WS server started on %s\n", server.address)
	listener, err := net.Listen("tcp", server.address)
	server.listener = &listener
	go func() {
		server.listenForConnections()
	}()
	if err != nil {
		log.Fatal("Could not start Tcp server:", err)
	}
	server.waitForSignal()

}

func (server *TcpServer) StartTls() {
	// cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// config := &tls.Config{Certificates: []tls.Certificate{cert}}
	log.Printf("INFO: WS server started on %s\n", server.address)
	listener, err := tls.Listen("tcp", server.address, server.config)
	server.listener = &listener
	server.listenForConnections()
	if err != nil {
		log.Fatal("Could not start Tcp server:", err)
	}
	server.waitForSignal()

}

func (server *TcpServer) OnMessageReceived(handler func(conn net.Conn, msg []byte)) {
	server.msgHandler = handler
}

func (server *TcpServer) waitForSignal() {
	go func() {
		for {
			select {
			case <-server.ctx.Done():
				log.Println("[server] caught interrupt signal")
				return
			case err := <-server.errChan:
				log.Printf("error : %s\n", err)
				return
			}
		}
	}()
}

func (server *TcpServer) Shutdown() {
	for _, c := range server.connections {
		c.Close()
	}
	close(server.errChan)
	server.wg.Wait()
}

func (server *TcpServer) closeConnection(clientID string) {
	server.mutex.Lock()
	server.connections[clientID].Close()
	delete(server.connections, clientID)
	server.mutex.Unlock()
	fmt.Println("Client disconnected:", clientID)
}

func (server *TcpServer) listenForConnections() {
	go func() {

		for {
			conn, err := (*server.listener).Accept()
			if err != nil {
				fmt.Println("could not estblish connection")
			}
			clientID := fmt.Sprintf("%p", conn)

			server.mutex.Lock()
			server.connections[clientID] = conn
			server.mutex.Unlock()

			initialMsg := fmt.Sprintf("New client connected with id : %s\n", clientID)
			log.Print(initialMsg)
			server.wg.Add(1)
			go server.handleConnection(clientID)
		}
	}()
}

func (server *TcpServer) handleConnection(clientID string) {
	defer server.wg.Done()
	defer server.closeConnection(clientID)

	for {
		reader := bufio.NewReader(server.connections[clientID])
		buffer, _, err := reader.ReadLine()

		if err != nil {
			if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
				break
			}
			log.Printf("Error on clients [%s] connection : %v", clientID, err)
			break
		}
		server.msgHandler(server.connections[clientID], buffer)
	}
}
