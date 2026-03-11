package p2p

import (
	"context"
	"log"

	"github.com/pauldin91/wsgo/client"
	"github.com/pauldin91/wsgo/server"
)

type Peer struct {
	this   client.Client
	server server.Server
}

func NewPeer(hostAddr, peerAddr, protocol string) (*Peer, error) {

	server, err := server.NewServer(hostAddr, protocol)
	if err != nil {
		return nil, err
	}
	client, err := client.NewClient(peerAddr, protocol)
	if err != nil {
		return nil, err
	}

	return &Peer{
		server: server,
		this:   client,
	}, nil
}

func (p *Peer) Start(ctx context.Context) {
	p.server.Start(ctx)
}

func (p *Peer) Shutdown() {
	p.server.Shutdown()
	p.this.Disconnect()
}

func (p *Peer) Connect(ctx context.Context) error {
	var err error
	if err = p.this.Connect(ctx); err != nil {
		log.Fatalf("unable to create client: %v", err.Error())
		return err
	}
	return nil
}

func (p *Peer) OnMessageReceived(serverHandler, clientHandler func([]byte)) {
	p.server.OnMessageReceived(serverHandler)
	p.this.OnMessageReceived(clientHandler)
}

func (p *Peer) Broadcast(msg []byte) {
	p.server.Broadcast(msg)
}

func (p *Peer) Send(msg []byte) {
	p.this.Send(msg)
}
