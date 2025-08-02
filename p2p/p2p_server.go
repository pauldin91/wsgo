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

	"github.com/pauldin91/wsgo/client"
	"github.com/pauldin91/wsgo/internal"
	"github.com/pauldin91/wsgo/server"
)

type P2PServer struct {
	address          string
	server           server.Server
	ctx              context.Context
	wg               *sync.WaitGroup
	protocol         internal.Protocol
	errChan          chan error
	peers            map[string]client.Client
	msgQueueIncoming map[string]chan internal.Message
	sendMsgHandler   func([]byte)
	rcvMsgHandler    func([]byte)
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
		rcvMsgHandler:    func([]byte) {},
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

func (p2p *P2PServer) SetMsgReceivedHandler(handle func([]byte)) {

	p2p.server.OnMessageReceived(handle)
}

func (p2p *P2PServer) MsgReceivedHandler(handle func([]byte)) {
	p2p.rcvMsgHandler = handle
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
			js, _ := json.Marshal(msg)
			p2p.rcvMsgHandler(js)
		})
		p2p.peers[client.GetConnId()] = client
	}
	fmt.Printf("connected to peers: %s\n", peers)
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
