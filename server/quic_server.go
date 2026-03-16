package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/pauldin91/wsgo/internal/crypto"
	"github.com/pauldin91/wsgo/protocol"
	"github.com/quic-go/quic-go"
)

type QUICServer struct {
	address                  string
	listener                 *quic.Listener
	connectionsMutex         sync.RWMutex
	connections              map[string]*quic.Conn
	wg                       *sync.WaitGroup
	onMessageReceivedHandler func([]byte)
}

func NewQuicServer(address string) *QUICServer {
	return &QUICServer{
		address:     address,
		connections: map[string]*quic.Conn{},
		wg:          &sync.WaitGroup{},
	}
}

func (s *QUICServer) Start(ctx context.Context) {
	var err error
	s.listener, err = quic.ListenAddr(s.address, crypto.GenerateTLSConfig(), nil)
	if err != nil {
		return
	}
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.acceptConnections(ctx)
	}()
}

func (s *QUICServer) acceptConnections(ctx context.Context) {
	for {
		conn, err := s.listener.Accept(ctx)
		if err != nil {
			return
		}
		s.connectionsMutex.Lock()
		s.connections[conn.RemoteAddr().String()] = conn
		s.connectionsMutex.Unlock()
		s.wg.Add(1)
		go func(c *quic.Conn) {
			defer s.wg.Done()
			s.handleConnection(ctx, c)
		}(conn)
	}
}

func (s *QUICServer) handleConnection(ctx context.Context, conn *quic.Conn) {
	stream, err := conn.AcceptStream(ctx)
	if err != nil {
		return
	}
	defer stream.Close()
	var buffer []byte = make([]byte, 1024)
	n, err := stream.Read(buffer)
	if err != nil {
		fmt.Printf("error read : %v\n", err.Error())
		return
	}
	s.onMessageReceivedHandler(buffer[:n])
}

func (s *QUICServer) OnMessageReceived(handler func([]byte)) {
	if handler != nil {
		s.onMessageReceivedHandler = handler
	}
}

func (s *QUICServer) Shutdown() {
	if s.listener != nil {
		s.listener.Close()
	}
	s.connectionsMutex.Lock()
	for _, c := range s.connections {
		c.CloseWithError(quic.ApplicationErrorCode(quic.NoError), "server closed")
	}
	s.connectionsMutex.Unlock()
	s.wg.Wait()
}

func (s *QUICServer) Broadcast(msg []byte) error {
	s.connectionsMutex.Lock()
	defer s.connectionsMutex.Unlock()
	for _, c := range s.connections {
		s, err := c.OpenStreamSync(context.Background())
		if err != nil {
			return err
		}
		s.Write(msg)
		s.Close()
	}
	return nil
}

func (s *QUICServer) SendTo(msg protocol.Message) error {
	s.connectionsMutex.RLock()
	defer s.connectionsMutex.RUnlock()
	if conn, ok := s.connections[msg.Receiver]; ok {
		message := protocol.Message{Sender: msg.Sender, Content: msg.Content}
		deliverable, err := json.Marshal(message)
		if err != nil {
			return err
		}
		s, err := conn.OpenStreamSync(context.Background())
		if err != nil {
			return err
		}
		s.Write(deliverable)
		s.Close()
	}
	return errors.New("address not found")

}
