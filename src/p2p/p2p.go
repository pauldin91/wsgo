package p2p

import (
	"bufio"
	"context"
	"log"
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pauldin91/wsgo/src/client"
	"github.com/pauldin91/wsgo/src/server"
)

type p2p interface {
	Start([]string)
	ExposeFirstPeerForInput()
	GetConnections() map[string]string
	Shutdown()
}

type P2PServer struct {
	server server.WsServer
	peers  map[string]*client.WsClient
	ctx    context.Context
	wg     *sync.WaitGroup
}

func NewP2PServer(ctx context.Context, address string) P2PServer {
	return P2PServer{
		server: server.NewWsServer(ctx, address),
		peers:  make(map[string]*client.WsClient),
		ctx:    ctx,
		wg:     &sync.WaitGroup{},
	}
}

func (p2p *P2PServer) connectToPeers(peers []string) {
	for _, p := range peers {
		client := client.NewWsClient(p2p.ctx, p)
		client.Connect()
		p2p.peers[client.GetConnId()] = client
	}
}

func (p2p *P2PServer) ExposeFirstPeerForInput() {
	for _, c := range p2p.peers {
		c.ListenForInput(bufio.NewReader(os.Stdin))
		break
	}
}

func (p2p *P2PServer) Start(peers []string) {
	p2p.wg.Add(1)
	go func() {
		defer p2p.wg.Done()
		p2p.server.Start()
		p2p.connectToPeers(peers)
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

func (p2p *P2PServer) GetConnections() map[string]string {
	var result map[string]string = make(map[string]string)
	for i, c := range p2p.server.GetConnections() {
		result[i] = c.RemoteAddr().String()
	}
	return result
}
func (p2p *P2PServer) GetConnectedClients() map[string]*websocket.Conn {
	return p2p.server.GetConnections()
}

func (p2p *P2PServer) Shutdown() {
	for _, p := range p2p.peers {
		p.Close()
	}
	p2p.wg.Wait()
	p2p.server.Shutdown()
}
