package p2p

import (
	"context"
	"log"

	"github.com/pauldin91/wsgo/client"
	"github.com/pauldin91/wsgo/server"
)

type P2PServer struct {
	this   client.Client
	server server.Server
}

func NewP2PServer(hostAddr, peerAddr, protocol string) (*P2PServer, error) {

	server, err := server.NewServer(hostAddr, protocol)
	if err != nil {
		return nil, err
	}
	client, err := client.NewClient(peerAddr, protocol)
	if err != nil {
		return nil, err
	}

	return &P2PServer{
		server: server,
		this:   client,
	}, nil
}

func (p *P2PServer) Start(ctx context.Context) {
	p.server.Start(ctx)
}

func (p *P2PServer) Shutdown() {
	p.server.Shutdown()
	p.this.Disconnect()
}

func (p *P2PServer) Connect(ctx context.Context) error {
	var err error
	if err = p.this.Connect(ctx); err != nil {
		log.Fatalf("unable to create client: %v", err.Error())
		return err
	}
	return nil
}

func (p *P2PServer) OnMessageReceived(serverHandler, clientHandler func([]byte)) {
	if serverHandler != nil {
		p.server.OnMessageReceived(serverHandler)
	}
	if clientHandler != nil {
		p.this.OnMessageReceived(clientHandler)
	}
}

func (p *P2PServer) Broadcast(msg []byte) {
	p.server.Broadcast(msg)
}

func (p *P2PServer) Send(msg []byte) {
	p.this.Send(msg)
}
