package p2p

import (
	"context"
	"fmt"
	"log"

	"github.com/pauldin91/wsgo/client"
	"github.com/pauldin91/wsgo/server"
)

type Peer struct {
	peers              []string
	this               client.Client
	server             server.Server
	protocol           string
	msgReceivedHandler func([]byte)
}

func NewPeer(addr, protocol string) *Peer {
	server, err := server.NewServer(addr, protocol)
	if err != nil {
		log.Fatalf("unable to create server: %v", err.Error())
	}
	return &Peer{
		server:             server,
		protocol:           protocol,
		msgReceivedHandler: func(m []byte) { fmt.Printf("generic handler %s\n", m) },
	}
}

func (p *Peer) Start(ctx context.Context) {
	p.server.Start(ctx)
}

func (p *Peer) Shutdown() {
	p.server.Shutdown()
	p.this.Disconnect()
}

func (p *Peer) Connect(ctx context.Context, addr string) error {
	var err error
	p.this, err = client.NewClient(ctx, addr, p.protocol)
	if err != nil {
		log.Fatalf("unable to create client: %v", err.Error())
		return err
	}
	if err = p.this.Connect(ctx); err != nil {
		log.Fatalf("unable to create client: %v", err.Error())
		return err
	}
	return nil
}

func (p *Peer) OnMessageReceived(handler func([]byte)) {
	p.msgReceivedHandler = handler
}

func (p *Peer) OnMessageReceivedByClient(handler func([]byte)) {
	p.this.OnMessageReceived(handler)
}

func (p *Peer) OnMessageReceivedByServer(handler func([]byte)) {
	p.server.OnMessageReceived(handler)
}
func (p *Peer) Broadcast(msg []byte) {
	p.server.Broadcast(msg)
}

func (p *Peer) Send(msg []byte) {
	p.this.Send(msg)
}
