package p2p

import (
	"context"
	"log"
	"sync"

	"github.com/pauldin91/wsgo/client"
	"github.com/pauldin91/wsgo/internal"
	"github.com/pauldin91/wsgo/server"
)

type P2PServer struct {
	server server.Server
	peers  map[string]client.Client
	ctx    context.Context
	wg     *sync.WaitGroup
}

func NewP2PServer(ctx context.Context, address string, pr internal.Protocol) P2PServer {
	return P2PServer{
		server: server.NewServer(ctx, address, pr),
		peers:  make(map[string]client.Client),
		ctx:    ctx,
		wg:     &sync.WaitGroup{},
	}
}

func (p2p *P2PServer) Start(peers ...string) {
	p2p.wg.Add(1)
	go func() {
		defer p2p.wg.Done()
		p2p.server.Start()
	}()

	p2p.wait()
}

func (p2p *P2PServer) Connect(peers ...string) {
	for _, p := range peers {
		client := client.NewWsClient(p2p.ctx, p)
		client.Connect()
		p2p.peers[client.GetConnId()] = client
	}
}

func (p2p *P2PServer) StartTls(certFile, certKey string) {
	p2p.wg.Add(1)
	go func() {
		defer p2p.wg.Done()
		p2p.server.StartTls()
	}()

	p2p.wait()
}

func (p2p *P2PServer) wait() {
	go func() {
		for {
			select {
			case <-p2p.ctx.Done():
				log.Println("[p2p]:Interrupt received. Exiting ...")
				return
			}
		}
	}()
}

func (p2p *P2PServer) Shutdown() {
	for _, p := range p2p.peers {
		p.Close()
	}
	p2p.wg.Wait()
	p2p.server.Shutdown()
}

func (p2p *P2PServer) BroadcastMessage(message string) {
	// for _, c := range p2p.GetConnectedClients() {
	// 	c.WriteJSON(message)
	// }
}
