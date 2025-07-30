package p2p

import (
	"context"
	"sync"

	"github.com/pauldin91/wsgo/client"
	"github.com/pauldin91/wsgo/internal"
	"github.com/pauldin91/wsgo/server"
)

type P2PServer struct {
	address          string
	server           server.Server
	peers            map[string]client.Client
	ctx              context.Context
	wg               *sync.WaitGroup
	protocol         internal.Protocol
	errChan          chan error
	msgQueueIncoming map[string]chan internal.Message
	sendMsgHandler   func([]byte)
}

func NewP2PServer(ctx context.Context, address string, pr internal.Protocol) *P2PServer {
	return &P2PServer{
		address:          address,
		server:           server.NewServer(ctx, address, pr),
		peers:            make(map[string]client.Client),
		ctx:              ctx,
		wg:               &sync.WaitGroup{},
		protocol:         pr,
		errChan:          make(chan error),
		msgQueueIncoming: make(map[string]chan internal.Message),
		sendMsgHandler:   func([]byte) {},
	}
}

func (p2p *P2PServer) Start() {
	p2p.wg.Add(1)
	go func() {
		defer p2p.wg.Done()
		err := p2p.server.Start()
		if err != nil {
			p2p.errChan <- err
			return
		}
	}()

	p2p.wait()
}

func (p2p *P2PServer) StartTls() {
	p2p.wg.Add(1)
	go func() {
		defer p2p.wg.Done()
		err := p2p.server.StartTls()
		if err != nil {
			p2p.errChan <- err
			return
		}
	}()

	p2p.wait()
}

func (p2p *P2PServer) SetMsgReceivedHandler(handle func([]byte)) {

	p2p.server.OnMessageReceived(handle)
}

func (p2p *P2PServer) SetSendMsg(clientId string, msg []byte) {
	p2p.server.GetConnections()[clientId].Write(msg)
}

func (p2p *P2PServer) Connect(peers ...string) {
	for _, p := range peers {
		client := client.NewClient(p2p.ctx, p, p2p.protocol)
		client.Connect()
		client.OnMessageReceivedHandler(func(msg []byte) {
			p2p.msgQueueIncoming[client.GetConnId()] = make(chan internal.Message)
			p2p.msgQueueIncoming[client.GetConnId()] <- internal.NewMessage(msg, p, p2p.address)
		})
		p2p.peers[client.GetConnId()] = client

	}
}
func (p2p *P2PServer) ReadMsg(clientId string) []internal.Message {
	msgs := make([]internal.Message, 0)

	for {
		select {
		case msg := <-p2p.msgQueueIncoming[clientId]:
			msgs = append(msgs, msg)
		default:
			return msgs
		}
	}
}

func (p2p *P2PServer) wait() {
	go func() {
		for {
			select {
			case <-p2p.errChan:
				return
			case <-p2p.ctx.Done():
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
