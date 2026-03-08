package client

import (
	"context"
	"fmt"
	"net"
)

type QuicClient struct {
	onMsgReceivedHandler func([]byte)
	msgParseHandler      func(net.Conn)
}

func NewQuicClient() *QuicClient {
	return &QuicClient{}
}

func (qs *QuicClient) Connect(ctx context.Context) error {
	return fmt.Errorf("QUIC client not implemented")
}

func (qs *QuicClient) Close() {

}

func (qs *QuicClient) GetConnId() string {
	return ""
}

func (qs *QuicClient) OnMessageReceived(handler func([]byte)) {
	qs.onMsgReceivedHandler = handler
}

func (qs *QuicClient) OnMessageParse(handler func(net.Conn)) {
	qs.msgParseHandler = handler

}

func (qs *QuicClient) SendError(err error) {

}

func (qs *QuicClient) Send(msg []byte) error {
	return fmt.Errorf("QUIC client not implemented")
}
