package server

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/pauldin91/wsgo/internal/crypto"
	"github.com/quic-go/quic-go"
)

type QuicServer struct {
	address                  string
	listener                 *quic.Listener
	connectionsMutex         sync.RWMutex
	connections              map[string]*quic.Conn
	wg                       *sync.WaitGroup
	onMessageReceivedHandler func([]byte)
}

func NewQuicServer(ctx context.Context, address string) *QuicServer {
	return &QuicServer{
		address:     address,
		connections: map[string]*quic.Conn{},
		wg:          &sync.WaitGroup{},
	}
}

func (s *QuicServer) Start(ctx context.Context) {
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

func (s *QuicServer) acceptConnections(ctx context.Context) {
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

func (s *QuicServer) handleConnection(ctx context.Context, conn *quic.Conn) {
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

func (s *QuicServer) OnMessageReceived(handler func([]byte)) {
	s.onMessageReceivedHandler = handler
}

func (s *QuicServer) GetConnections() map[string]net.Conn {
	return make(map[string]net.Conn)
}

func (s *QuicServer) Shutdown() {
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
