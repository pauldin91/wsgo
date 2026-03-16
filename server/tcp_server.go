package server

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"sync"

	"github.com/pauldin91/wsgo/protocol"
)

type TcpServer struct {
	address                  string
	connectionsMutex         sync.RWMutex
	connections              map[string]net.Conn
	wg                       *sync.WaitGroup
	errorChan                chan error
	listener                 net.Listener
	onMessageReceivedHandler func([]byte)
	tlsConfig                *tls.Config
}

func NewTcpServer(serveAddress string) *TcpServer {
	return &TcpServer{
		address:                  serveAddress,
		connections:              make(map[string]net.Conn),
		errorChan:                make(chan error, 1),
		wg:                       &sync.WaitGroup{},
		onMessageReceivedHandler: func(bytes []byte) {},
	}
}

func (s *TcpServer) Start(ctx context.Context) {
	var err error
	if s.tlsConfig == nil {
		s.listener, err = net.Listen("tcp", s.address)
	} else {
		s.listener, err = tls.Listen("tcp", s.address, s.tlsConfig)
	}
	if err != nil {
		select {
		case s.errorChan <- err:
		default:
		}
		return
	}
	s.wg.Add(2)
	go func() {
		defer s.wg.Done()
		s.acceptConnections()
	}()
	go func() {
		defer s.wg.Done()
		s.waitForShutdown(ctx)
	}()
}

func (s *TcpServer) OnMessageReceived(handler func([]byte)) {
	if handler != nil {
		s.onMessageReceivedHandler = handler
	}
}

func (s *TcpServer) Shutdown() {
	if s.listener != nil {
		s.listener.Close()
	}
	s.connectionsMutex.Lock()
	for _, c := range s.connections {
		c.Close()
	}
	s.connectionsMutex.Unlock()
	s.wg.Wait()
	close(s.errorChan)
}

func (s *TcpServer) SendTo(msg protocol.Message) {
	s.connectionsMutex.RLock()
	defer s.connectionsMutex.RUnlock()
	if conn, ok := s.connections[msg.Receiver]; ok {
		message := protocol.Message{Sender: msg.Sender, Content: msg.Content}
		deliverable, err := json.Marshal(message)
		if err != nil {
			log.Printf("error: %v,could not deliver msg %s", err, message)
		}
		conn.Write(deliverable)
	} else {
		log.Printf("receiver %s not found", msg.Receiver)
	}

}

func (s *TcpServer) closeConnection(clientID string) {
	s.connectionsMutex.Lock()
	if conn, exists := s.connections[clientID]; exists {
		conn.Close()
		delete(s.connections, clientID)
	}
	s.connectionsMutex.Unlock()
}

func (s *TcpServer) acceptConnections() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case s.errorChan <- err:
			default:
			}
			return
		}
		clientID := conn.RemoteAddr().String()
		s.connectionsMutex.Lock()
		s.connections[clientID] = conn
		s.connectionsMutex.Unlock()
		s.wg.Add(1)
		go s.handleConnection(clientID)
	}
}

func (s *TcpServer) handleConnection(clientID string) {
	defer s.wg.Done()
	defer s.closeConnection(clientID)

	s.connectionsMutex.RLock()
	conn := s.connections[clientID]
	s.connectionsMutex.RUnlock()

	reader := bufio.NewReader(conn)
	for {
		buffer, _, err := reader.ReadLine()
		if err != nil {
			if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
				log.Printf("Client %s disconnected\n", clientID)
			} else {
				select {
				case s.errorChan <- err:
				default:
				}
			}
			break
		}
		s.onMessageReceivedHandler(buffer)
	}
}

func (s *TcpServer) waitForShutdown(ctx context.Context) {
	select {
	case rcv := <-ctx.Done():
		log.Printf("shutdown signal received %v\n", rcv)
	case err := <-s.errorChan:
		log.Printf("error %v\n", err)
	}
}

func (s *TcpServer) Broadcast(msg []byte) error {
	s.connectionsMutex.Lock()
	defer s.connectionsMutex.Unlock()
	for _, c := range s.connections {
		_, err := c.Write([]byte(string(msg) + "\n"))
		if err != nil {
			return err
		}
	}
	return nil
}
