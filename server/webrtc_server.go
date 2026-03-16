package server

import (
	"context"
	"fmt"
)

type WebRTCServer struct {
	onMessageReceivedHandler func([]byte)
}

func NewWebRTCServer() WebRTCServer {
	return WebRTCServer{
		onMessageReceivedHandler: func(msg []byte) {
			fmt.Printf("Received : %s", string(msg))
		},
	}
}

func (s *WebRTCServer) Start(ctx context.Context) {

	<-ctx.Done()
}
func (s *WebRTCServer) OnMessageReceived(handler func([]byte)) {
	if handler != nil {
		s.onMessageReceivedHandler = handler
	}
}

func (s *WebRTCServer) Shutdown() {}
