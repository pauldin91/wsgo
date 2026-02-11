package server

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net"
	"sync"
)

type TcpServer struct {
	address     string
	mutex       sync.Mutex
	connections map[string]net.Conn
	wg          *sync.WaitGroup
	errChan     chan error
	listener    net.Listener
	msgHandler  func([]byte)
	config      *tls.Config
}

func NewTcpServer(serveAddress string) *TcpServer {
	server := &TcpServer{
		address:     serveAddress,
		connections: make(map[string]net.Conn),
		errChan:     make(chan error),
		mutex:       sync.Mutex{},
		wg:          &sync.WaitGroup{},
	}

	return server
}

func (server *TcpServer) Start(ctx context.Context) {
	var err error
	if server.config == nil {
		server.listener, err = net.Listen("tcp", server.address)
	} else {
		server.listener, err = tls.Listen("tcp", server.address, server.config)
	}
	if err != nil {
		select {
		case server.errChan <- err:
		default:
		}
		return
	}
	server.wg.Add(2)
	go func() {
		defer server.wg.Done()
		server.serve()
	}()
	go func() {
		defer server.wg.Done()
		server.waitForShutdown(ctx)
	}()
}

func (server *TcpServer) OnMessageReceived(handler func(msg []byte)) {
	server.msgHandler = handler
}

func (server *TcpServer) Shutdown() {
	if server.listener != nil {
		server.listener.Close()
	}
	for _, c := range server.connections {
		c.Close()
	}
	server.wg.Wait()
	if server.errChan != nil {
		close(server.errChan)
	}
}

func (server *TcpServer) GetConnections() map[string]net.Conn {
	return server.connections
}

func (server *TcpServer) closeConnection(clientID string) {
	server.mutex.Lock()
	server.connections[clientID].Close()
	delete(server.connections, clientID)
	server.mutex.Unlock()
}

func (server *TcpServer) serve() {
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			select {
			case server.errChan <- err:
			default:
			}
			return
		}
		clientID := conn.RemoteAddr().String()
		server.mutex.Lock()
		server.connections[clientID] = conn
		server.mutex.Unlock()
		server.wg.Add(1)
		go server.handleConnection(clientID)
	}
}

func (server *TcpServer) handleConnection(clientID string) {
	defer server.wg.Done()
	defer server.closeConnection(clientID)

	reader := bufio.NewReader(server.connections[clientID])
	for {
		buffer, _, err := reader.ReadLine()
		if err != nil {
			if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
				log.Printf("Client %s disconnected\n", clientID)
				break
			}
			select {
			case server.errChan <- err:
			default:
			}
			break
		}
		server.msgHandler(buffer)
	}
}

func (server *TcpServer) waitForShutdown(ctx context.Context) {

	for {
		select {
		case rcv := <-ctx.Done():
			log.Printf("shutdown signal received %v\n", rcv)
			return
		case err := <-server.errChan:
			log.Printf("error %v\n", err)
			return
		}
	}

}
