package server

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/pauldin91/wsgo/internal"
	"github.com/quic-go/quic-go"
)

type loggingWriter struct{ io.Writer }

func (w loggingWriter) Write(b []byte) (int, error) {
	fmt.Printf("Server: Got '%s'\n", string(b))
	return w.Writer.Write(b)
}

type QuicServer struct {
	ctx           context.Context
	address       string
	listener      *quic.Listener
	connections   map[string]*quic.Conn
	msgRcvHandler func([]byte)
}

func NewQuicServer(ctx context.Context, address string) *QuicServer {
	return &QuicServer{
		ctx:           ctx,
		address:       address,
		connections:   map[string]*quic.Conn{},
		msgRcvHandler: func(b []byte) {},
	}
}

func (qs *QuicServer) Start() error {
	var err error
	qs.listener, err = quic.ListenAddr(qs.address, nil, nil)
	if err != nil {
		return err
	}
	qs.handleConnections()
	return nil
}

func (qs *QuicServer) StartTls() error {
	var err error
	qs.listener, err = quic.ListenAddr(qs.address, internal.GenerateTLSConfig(), nil)
	if err != nil {
		return err
	}
	qs.handleConnections()
	return nil

}

func (qs *QuicServer) handleConnections() {
	go func() {
		for {

			conn, err := qs.listener.Accept(context.Background())
			if err != nil {
				continue
			}
			qs.connections[conn.RemoteAddr().String()] = conn
			qs.handle(conn)
		}
	}()

}

func (qs *QuicServer) handle(conn *quic.Conn) {
	go func() {
		stream, err := conn.AcceptStream(qs.ctx)
		if err != nil {
			return
		}
		var buffer []byte = make([]byte, 0)
		_, err = stream.Read(buffer)
		if err != nil {
			fmt.Printf("error read : %v\n", err.Error())
		}
	}()
}

func (qs *QuicServer) OnMessageReceived(handler func([]byte)) {
	qs.msgRcvHandler = handler
}

func (qs *QuicServer) GetConnections() net.Conn {
	panic("unimplemented")
}

func (qs *QuicServer) Shutdown() {
	for _, c := range qs.connections {
		c.CloseWithError(quic.ApplicationErrorCode(quic.NoError), "server closed")
	}

}
