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

type TCPServer struct {
	connectionsMutex         sync.RWMutex
	connections              map[string]net.Conn
	wg                       *sync.WaitGroup
	listener                 net.Listener
	onMessageReceivedHandler func([]byte)
}

func NewTCPServer(serveAddress string, tlsConfig *tls.Config) *TCPServer {
	var err error
	var ln net.Listener
	if tlsConfig != nil {
		ln, err = tls.Listen("tcp", serveAddress, tlsConfig)
	} else {
		ln, err = net.Listen("tcp", serveAddress)

	}
	if err != nil {
		return nil
	}
	return &TCPServer{
		connections:              make(map[string]net.Conn),
		wg:                       &sync.WaitGroup{},
		listener:                 ln,
		onMessageReceivedHandler: func(bytes []byte) {},
	}
}

func (s *TCPServer) Start(ctx context.Context) {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case rcv := <-ctx.Done():
				log.Printf("shutdown signal received %v\n", rcv)
				s.Shutdown()
				break
			default:
			}
			return
		}
		clientID := conn.RemoteAddr().String()
		s.connectionsMutex.Lock()
		s.connections[clientID] = conn
		s.connectionsMutex.Unlock()
		s.wg.Add(1)
		go s.handleConnection(ctx, clientID)
	}
}

func (s *TCPServer) OnMessageReceived(handler func([]byte)) {
	if handler != nil {
		s.onMessageReceivedHandler = handler
	}
}

func (s *TCPServer) Shutdown() {
	if s.listener != nil {
		s.listener.Close()
	}
	s.connectionsMutex.Lock()
	for _, c := range s.connections {
		c.Close()
	}
	s.connectionsMutex.Unlock()

}

func (s *TCPServer) SendTo(msg protocol.Message) error {
	s.connectionsMutex.RLock()
	defer s.connectionsMutex.RUnlock()
	if conn, ok := s.connections[msg.Receiver]; ok {
		message := protocol.Message{Sender: msg.Sender, Content: msg.Content}
		deliverable, err := json.Marshal(message)
		if err != nil {
			return err
		}
		conn.Write(deliverable)
		return nil
	}
	return errors.New("receiver not found")

}

func (s *TCPServer) closeConnection(clientID string) {
	s.connectionsMutex.Lock()
	if conn, exists := s.connections[clientID]; exists {
		conn.Close()
		delete(s.connections, clientID)
	}
	s.connectionsMutex.Unlock()
}

func (s *TCPServer) GetConnections() map[string]string {
	result := make(map[string]string)
	s.connectionsMutex.Lock()
	defer s.connectionsMutex.Unlock()
	for _, c := range s.connections {
		result[c.RemoteAddr().String()] = c.RemoteAddr().String()
	}
	return result

}

func (s *TCPServer) handleConnection(ctx context.Context, clientID string) {
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
				case rcv := <-ctx.Done():
					log.Printf("shutdown signal received %v\n", rcv)
					conn.Close()
					break
				default:
				}
			}
			break
		}
		s.onMessageReceivedHandler(buffer)
	}
}

func (s *TCPServer) Broadcast(msg []byte) error {
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
