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
	ctx           context.Context
	cancel        context.CancelFunc
	address       string
	listener      *quic.Listener
	connections   map[string]*quic.Conn
	mutex         sync.Mutex
	wg            *sync.WaitGroup
	msgRcvHandler func([]byte)
}

func NewQuicServer(ctx context.Context, address string) *QuicServer {
	return &QuicServer{
		ctx:         ctx,
		address:     address,
		connections: map[string]*quic.Conn{},
		mutex:       sync.Mutex{},
		wg:          &sync.WaitGroup{},
	}
}

func (qs *QuicServer) Start(ctx context.Context) {
	var err error
	qs.listener, err = quic.ListenAddr(qs.address, crypto.GenerateTLSConfig(), nil)
	if err != nil {
		return
	}
	qs.ctx, qs.cancel = context.WithCancel(ctx)
	qs.wg.Add(1)
	go func() {
		defer qs.wg.Done()
		qs.handleConnections()
	}()
}

func (qs *QuicServer) handleConnections() {
	for {
		conn, err := qs.listener.Accept(qs.ctx)
		if err != nil {
			return
		}
		qs.mutex.Lock()
		qs.connections[conn.RemoteAddr().String()] = conn
		qs.mutex.Unlock()
		qs.wg.Add(1)
		go func(c *quic.Conn) {
			defer qs.wg.Done()
			qs.handle(c)
		}(conn)
	}
}

func (qs *QuicServer) handle(conn *quic.Conn) {
	stream, err := conn.AcceptStream(qs.ctx)
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
	qs.msgRcvHandler(buffer[:n])
}

func (qs *QuicServer) OnMessageReceived(handler func([]byte)) {
	qs.msgRcvHandler = handler
}

func (qs *QuicServer) GetConnections() map[string]net.Conn {
	return make(map[string]net.Conn)
}

func (qs *QuicServer) Shutdown() {
	if qs.cancel != nil {
		qs.cancel()
	}
	if qs.listener != nil {
		qs.listener.Close()
	}
	qs.mutex.Lock()
	for _, c := range qs.connections {
		c.CloseWithError(quic.ApplicationErrorCode(quic.NoError), "server closed")
	}
	qs.mutex.Unlock()
	qs.wg.Wait()
}
