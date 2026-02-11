package client

import (
	"context"
	"fmt"
	"net"
)

type QuicClient struct {
	ctx                context.Context
	cancel             context.CancelFunc
	msgReceivedHandler func([]byte)
	msgParseHandler    func(net.Conn)
}

func NewQuicClient() *QuicClient {
	return &QuicClient{}
}

func (qs *QuicClient) Connect(ctx context.Context) error {
	return fmt.Errorf("QUIC client not implemented")
}

func (qs *QuicClient) Close() {
	if qs.cancel != nil {
		qs.cancel()
	}
}

func (qs *QuicClient) GetConnId() string {
	return ""
}

func (qs *QuicClient) OnMessageReceivedHandler(handler func([]byte)) {
	qs.msgReceivedHandler = handler
}

func (qs *QuicClient) OnMessageParseHandler(handler func(net.Conn)) {
	qs.msgParseHandler = handler

}

func (qs *QuicClient) SendError(err error) {

}

func (qs *QuicClient) Send(msg []byte) error {
	return fmt.Errorf("QUIC client not implemented")
}
