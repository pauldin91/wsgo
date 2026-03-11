package client

import (
	"context"
	"fmt"
	"net"
)

type Client interface {
	Close()
	Connect(context.Context) error
	Disconnect() error
	GetConnId() string
	OnMessageReceived(func([]byte))
	OnMessageParse(func(net.Conn))
	SendError(err error)
	Send([]byte) error
}

func NewClient(addr string, protocol string) (Client, error) {
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
