package server

import (
	"context"
	"crypto/tls"
	"fmt"
)

type Server interface {
	Start(context.Context)
	//what will the server do with the message
	OnMessageReceived(handler func([]byte))
	GetConnections() map[string]string
	Broadcast([]byte) error
	Shutdown()
}

func NewServer(addr, protocol string, tlsConfig *tls.Config) (Server, error) {
	switch protocol {
	case "tcp":
		return NewTCPServer(addr, tlsConfig), nil
	case "websocket", "ws":
		return NewWSServerWithCerts(addr, nil), nil
	case "quic":
		return NewQuicServer(addr), nil
	case "webrtc":
		return NewWebRTCServer(addr), nil
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", protocol)
	}
}
