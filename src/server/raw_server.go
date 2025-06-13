package server

import (
	"bufio"
	"context"
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
	connections map[string]*net.Conn
	ctx         context.Context
	cancel      context.CancelFunc
	wg          *sync.WaitGroup
	errChan     chan error
	listener    *net.Listener
}

func NewTcpServer(ctx context.Context, serveAddress string) TcpServer {
	cotx, cancel := context.WithCancel(ctx)
	return TcpServer{
		address:     serveAddress,
		ctx:         cotx,
		cancel:      cancel,
		connections: make(map[string]*net.Conn),
		errChan:     make(chan error),
		mutex:       sync.Mutex{},
		wg:          &sync.WaitGroup{},
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
		(*c).Close()
	}
	close(srv.errChan)
	srv.wg.Wait()
}

func (ws *TcpServer) GetConnections() map[string]*net.Conn {
	return ws.connections
}

func (ws *TcpServer) broadcastMessage(message string) {
	ws.mutex.Lock()
	defer ws.mutex.Unlock()

	for clientID, conn := range ws.connections {

		if _, err := (*conn).Write([]byte(message)); err != nil {
			log.Println("Error writing to client", clientID, ":", err)

			delete(ws.connections, clientID)
		}
	}
}

func (ws *TcpServer) closeConnection(clientID string) {
	ws.mutex.Lock()
	(*ws.connections[clientID]).Close()
	delete(ws.connections, clientID)
	ws.mutex.Unlock()
	fmt.Println("Client disconnected:", clientID)
}

func (ws *TcpServer) listenForConnections() {

	for {
		conn, err := (*ws.listener).Accept()
		if err != nil {
			fmt.Println("could not estblish connection")
		}
		clientID := fmt.Sprintf("%p", conn)

		ws.mutex.Lock()
		ws.connections[clientID] = &conn
		ws.mutex.Unlock()

		initialMsg := fmt.Sprintf("New client connected with id : %s\n", clientID)
		log.Print(initialMsg)
		ws.wg.Add(1)
		go ws.handleConnection(clientID)
	}
}

func (server *TcpServer) handleConnection(clientID string) {
	defer server.wg.Done()
	defer server.closeConnection(clientID)

	for {
		reader := bufio.NewReader(*server.connections[clientID])
		buffer, _, err := reader.ReadLine()

		if err != nil {
			if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
				break
			}
			log.Printf("Error on clients [%s] connection : %v", clientID, err)
			break
		}
		msg := fmt.Sprintf("[client] %s says: %s\n", clientID, string(buffer))
		server.broadcastMessage(msg)
		log.Printf("[client] %s says: %s\n", clientID, string(buffer))
	}
}
