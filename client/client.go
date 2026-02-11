package client

import (
	"context"
	"fmt"
	"net"
)

type Client interface {
	Close()
	Connect(context.Context) error
	GetConnId() string
	OnMessageReceivedHandler(func([]byte))
	OnMessageParseHandler(func(net.Conn))
	SendError(err error)
	Send([]byte) error
}

func NewClient(ctx context.Context, addr string, protocol string) (Client, error) {
	switch protocol {
	case "tcp":
		return NewTcpClient(addr), nil
	case "websocket", "ws":
		return NewWsClient(addr), nil
	case "quic":
		return nil, fmt.Errorf("QUIC protocol not yet implemented")
	case "webrtc":
		return nil, fmt.Errorf("WebRTC protocol not yet implemented")
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}
