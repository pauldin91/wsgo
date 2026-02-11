package p2p

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pauldin91/wsgo/client"
	proto "github.com/pauldin91/wsgo/protocol"
	"github.com/pauldin91/wsgo/server"
)

type P2PServer struct {
	address          string
	server           server.Server
	ctx              context.Context
	cancel           context.CancelFunc
	wg               *sync.WaitGroup
	protocolType     string
	errChan          chan error
	peers            map[string]client.Client
	peersMutex       sync.Mutex
	wspeers          map[string]*websocket.Conn
	msgQueueIncoming map[string]chan proto.Message
	sendMsgHandler   func([]byte)
	rcvMsgHandler    func([]byte)
}

func NewP2PServer(ctx context.Context, address string, protocol string) (*P2PServer, error) {
	srv, err := server.NewServer(address, protocol)
	if err != nil {
		return nil, err
	}
	return &P2PServer{
		address:          address,
		server:           srv,
		ctx:              ctx,
		peers:            make(map[string]client.Client),
		peersMutex:       sync.Mutex{},
		wspeers:          make(map[string]*websocket.Conn),
		wg:               &sync.WaitGroup{},
		protocolType:     protocol,
		errChan:          make(chan error, 1),
		msgQueueIncoming: make(map[string]chan proto.Message),
	}, nil
}

func (p2p *P2PServer) GetConnections() map[string]string {
	clients := make(map[string]string)
	for k, v := range p2p.server.GetConnections() {
		clients[k] = v.RemoteAddr().String()

	}
	return clients
}

func (p2p *P2PServer) Start(ctx context.Context) {
	p2p.ctx, p2p.cancel = context.WithCancel(ctx)
	p2p.wg.Add(2)
	go func() {
		defer p2p.wg.Done()
		p2p.server.Start(p2p.ctx)
	}()
	go func() {
		defer p2p.wg.Done()
		p2p.wait(p2p.ctx)
	}()
}

func (p2p *P2PServer) SetMsgReceivedHandler(handle func([]byte)) {

	p2p.server.OnMessageReceived(handle)
}

func (p2p *P2PServer) MsgReceivedHandler(handle func([]byte)) {
	p2p.rcvMsgHandler = handle
}

func (p2p *P2PServer) SetSendMsg(clientId string, msg []byte) {
	p2p.server.GetConnections()[clientId].Write(msg)
}

func (p2p *P2PServer) Connect(peers ...string) error {
	for _, p := range peers {
		cl, err := client.NewClient(p2p.ctx, p, p2p.protocolType)
		if err != nil {
			return fmt.Errorf("failed to create client for %s: %w", p, err)
		}
		if err := cl.Connect(p2p.ctx); err != nil {
			return fmt.Errorf("failed to connect to %s: %w", p, err)
		}
		cl.OnMessageReceivedHandler(func(msg []byte) {
			p2p.msgQueueIncoming[cl.GetConnId()] = make(chan proto.Message, 10)
			if m, err := proto.NewMessage(msg, p, p2p.address); err == nil {
				select {
				case p2p.msgQueueIncoming[cl.GetConnId()] <- m:
				default:
				}
			}
			js, _ := json.Marshal(msg)
			p2p.rcvMsgHandler(js)
		})
		p2p.peersMutex.Lock()
		p2p.peers[cl.GetConnId()] = cl
		p2p.peersMutex.Unlock()
	}
	fmt.Printf("connected to peers: %s\n", peers)
	return nil
}

func (ws *P2PServer) OnParseMsgHandler(src *os.File) {
	reader := bufio.NewReader(src)
	var s string = "Available connections:-> "
	c := 0
	var m map[int]client.Client = make(map[int]client.Client)
	for i, cl := range ws.peers {
		s += fmt.Sprintf("%d. %s, ", c, i)
		m[c] = cl
		c++
	}
	s = strings.Trim(s, ",")

	fmt.Println(s)

	input, _ := reader.ReadString('\n')
	choice, _ := strconv.Atoi(input)
	fmt.Printf("selected choice is : %d\n", choice)

	parser := func() {

		for {
			input, _, err := reader.ReadLine()
			fmt.Println(input)
			if err != nil {

				return
			}
			text := strings.TrimSpace(string(input))
			if text == "exit" {
				return
			}
			m[choice].Send([]byte(string(input) + "\n"))
		}

	}
	ws.wg.Add(1)
	go func() {
		defer ws.wg.Done()
		go parser()
	}()
}

func (p2p *P2PServer) wait(ctx context.Context) {
	for {
		select {
		case <-p2p.errChan:
			return
		case <-ctx.Done():
			return
		}
	}
}

func (p2p *P2PServer) Shutdown() {
	if p2p.cancel != nil {
		p2p.cancel()
	}
	p2p.peersMutex.Lock()
	for _, p := range p2p.peers {
		p.Close()
	}
	p2p.peersMutex.Unlock()
	p2p.wg.Wait()
	p2p.server.Shutdown()
	if p2p.errChan != nil {
		close(p2p.errChan)
	}
}
