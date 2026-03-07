package p2p

import (
	"context"
	"log"

	"github.com/pauldin91/wsgo/client"
	"github.com/pauldin91/wsgo/server"
)

type Peer struct {
	peers    []string
	this     client.Client
	server   server.Server
	protocol string
}

func NewPeer(addr, protocol string) *Peer {
	server, err := server.NewServer(addr, protocol)
	if err != nil {
		log.Fatalf("unable to create server: %v", err.Error())
	}
	return &Peer{
		server:   server,
		protocol: protocol,
	}
}

func (p *Peer) Start(ctx context.Context) {
	p.server.Start(ctx)
}

func (p *Peer) Shutdown() {
	p.server.Shutdown()
}

func (p *Peer) Connect(ctx context.Context, addr string) {
	var err error
	p.this, err = client.NewClient(ctx, addr, p.protocol)
	if err != nil {
		log.Fatalf("unable to create client: %v", err.Error())
	}
	if err = p.this.Connect(ctx); err != nil {
		log.Fatalf("unable to create client: %v", err.Error())
	}
}
