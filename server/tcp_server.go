package server

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"io"
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
	listener    net.Listener
	msgHandler  func([]byte)
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
		msgHandler:  func([]byte) {},
	}

	return server
}

func (server *TcpServer) Start() error {
	var err error
	if server.config == nil {
		server.listener, err = net.Listen("tcp", server.address)
	} else {
		server.listener, err = tls.Listen("tcp", server.address, server.config)

	}
	if err != nil {
		return err
	}
	server.serve()
	server.waitForShutdown()
	return nil

}

func (server *TcpServer) OnMessageReceived(handler func(msg []byte)) {
	server.msgHandler = handler
}

func (server *TcpServer) Shutdown() {
	for _, c := range server.connections {
		c.Close()
	}
	close(server.errChan)
	server.wg.Wait()
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
	go func() {
		for {
			conn, err := server.listener.Accept()
			if err != nil {
				server.errChan <- err
				return
			}
			clientID := conn.RemoteAddr().String()
			server.mutex.Lock()
			server.connections[clientID] = conn
			server.mutex.Unlock()
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
			server.errChan <- err
			break
		}
		server.msgHandler(buffer)
	}
}

func (server *TcpServer) waitForShutdown() {
	go func() {
		for {
			select {
			case <-server.ctx.Done():
				return
			case <-server.errChan:
				return
			}
		}
	}()
}
