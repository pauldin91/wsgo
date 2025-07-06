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
	cancel      context.CancelFunc
	wg          *sync.WaitGroup
	errChan     chan error
	listener    *net.Listener
	msgHandler  func(msg string)
	config      *tls.Config
}

func NewTcpServer(ctx context.Context, serveAddress string) *TcpServer {
	cotx, cancel := context.WithCancel(ctx)
	return &TcpServer{
		address:     serveAddress,
		ctx:         cotx,
		cancel:      cancel,
		connections: make(map[string]net.Conn),
		errChan:     make(chan error),
		mutex:       sync.Mutex{},
		wg:          &sync.WaitGroup{},
		msgHandler:  func(string) {},
	}
}

func (ws *TcpServer) Start() {

	log.Printf("INFO: WS server started on %s\n", ws.address)
	listener, err := net.Listen("tcp", ws.address)
	ws.listener = &listener
	go func() {
		ws.listenForConnections()
	}()
	if err != nil {
		log.Fatal("Could not start Tcp server:", err)
	}
	ws.waitForSignal()

}

func (ws *TcpServer) StartTls() {
	// cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// config := &tls.Config{Certificates: []tls.Certificate{cert}}
	log.Printf("INFO: WS server started on %s\n", ws.address)
	listener, err := tls.Listen("tcp", ws.address, ws.config)
	ws.listener = &listener
	ws.listenForConnections()
	if err != nil {
		log.Fatal("Could not start Tcp server:", err)
	}
	ws.waitForSignal()

}

func (server *TcpServer) OnMessageReceived(handler func(msg string)) {
	server.msgHandler = handler
}

func (ws *TcpServer) waitForSignal() {
	go func() {
		for {
			select {
			case <-ws.ctx.Done():
				log.Println("[server] caught interrupt signal")
				ws.cancel()
				return
			case err := <-ws.errChan:
				log.Printf("error : %s\n", err)
				return
			}
		}
	}()
}

func (srv *TcpServer) Shutdown() {
	for _, c := range srv.connections {
		c.Close()
	}
	close(srv.errChan)
	srv.wg.Wait()
}

func (ws *TcpServer) closeConnection(clientID string) {
	ws.mutex.Lock()
	ws.connections[clientID].Close()
	delete(ws.connections, clientID)
	ws.mutex.Unlock()
	fmt.Println("Client disconnected:", clientID)
}

func (ws *TcpServer) listenForConnections() {
	go func() {

		for {
			conn, err := (*ws.listener).Accept()
			if err != nil {
				fmt.Println("could not estblish connection")
			}
			clientID := fmt.Sprintf("%p", conn)

			ws.mutex.Lock()
			ws.connections[clientID] = conn
			ws.mutex.Unlock()

			initialMsg := fmt.Sprintf("New client connected with id : %s\n", clientID)
			log.Print(initialMsg)
			ws.wg.Add(1)
			go ws.handleConnection(clientID)
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
		server.msgHandler(string(buffer))
	}
}
