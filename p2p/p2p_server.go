package p2p

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/pauldin91/wsgo/client"
	proto "github.com/pauldin91/wsgo/protocol"
	"github.com/pauldin91/wsgo/server"
)

type P2PServer struct {
	address                  string
	server                   server.Server
	wg                       *sync.WaitGroup
	protocolType             string
	errorChan                chan error
	peersMutex               sync.RWMutex
	peers                    map[string]client.Client
	incomingMsgQueues        map[string]chan proto.Message
	incomingMsgQueuesMux     sync.RWMutex
	onMessageSentHandler     func([]byte)
	onMessageReceivedHandler func([]byte)
}

func NewP2PServer(ctx context.Context, address string, protocol string) (*P2PServer, error) {
	srv, err := server.NewServer(address, protocol)
	if err != nil {
		return nil, err
	}
	return &P2PServer{
		address:                  address,
		server:                   srv,
		peers:                    make(map[string]client.Client),
		wg:                       &sync.WaitGroup{},
		protocolType:             protocol,
		errorChan:                make(chan error, 1),
		incomingMsgQueues:        make(map[string]chan proto.Message),
		onMessageSentHandler:     func(msg []byte) {},
		onMessageReceivedHandler: func(msg []byte) { log.Printf("Echo: %v\n", msg) },
	}, nil
}

func (p *P2PServer) GetConnections() map[string]string {
	p.peersMutex.RLock()
	defer p.peersMutex.RUnlock()
	conns := p.server.GetConnections()
	clients := make(map[string]string, len(conns))
	for k, v := range conns {
		clients[k] = v.RemoteAddr().String()
	}
	return clients
}

func (p *P2PServer) Start(ctx context.Context) {
	p.wg.Add(2)
	go func() {
		defer p.wg.Done()
		p.server.Start(ctx)
	}()
	go func() {
		defer p.wg.Done()
		p.waitForShutdown(ctx)
	}()
}

func (p *P2PServer) SetMsgReceivedHandler(handle func([]byte)) {
	p.server.OnMessageReceived(handle)
}

func (p *P2PServer) MsgReceivedHandler(handle func([]byte)) {
	p.onMessageReceivedHandler = handle
}

func (p *P2PServer) SetSendMsg(clientId string, msg []byte) {
	conns := p.server.GetConnections()
	if conn, exists := conns[clientId]; exists {
		conn.Write(msg)
	}
}

func (p *P2PServer) Connect(peers ...string) error {
	for _, peerAddr := range peers {
		cl, err := client.NewClient(context.Background(), peerAddr, p.protocolType)
		if err != nil {
			return fmt.Errorf("failed to create client for %s: %w", peerAddr, err)
		}
		if err := cl.Connect(context.Background()); err != nil {
			return fmt.Errorf("failed to connect to %s: %w", peerAddr, err)
		}
		connID := cl.GetConnId()
		cl.OnMessageReceivedHandler(func(msg []byte) {
			p.incomingMsgQueuesMux.Lock()
			if _, exists := p.incomingMsgQueues[connID]; !exists {
				p.incomingMsgQueues[connID] = make(chan proto.Message, 10)
			}
			p.incomingMsgQueuesMux.Unlock()

			if m, err := proto.NewMessage(msg, peerAddr, p.address); err == nil {
				select {
				case p.incomingMsgQueues[connID] <- m:
				default:
				}
			}
			js, _ := json.Marshal(msg)
			p.onMessageReceivedHandler(js)
		})
		p.peersMutex.Lock()
		p.peers[connID] = cl
		p.peersMutex.Unlock()
	}
	fmt.Printf("connected to peers: %s\n", peers)
	return nil
}

func (p *P2PServer) OnParseMsgHandler(src *os.File) {
	reader := bufio.NewReader(src)
	var s string = "Available connections:-> "
	c := 0

	p.peersMutex.RLock()
	var m map[int]client.Client = make(map[int]client.Client, len(p.peers))
	for i, cl := range p.peers {
		s += fmt.Sprintf("%d. %s, ", c, i)
		m[c] = cl
		c++
	}
	p.peersMutex.RUnlock()

	s = strings.Trim(s, ",")
	fmt.Println(s)

	input, _ := reader.ReadString('\n')
	choice, _ := strconv.Atoi(strings.TrimSpace(input))
	fmt.Printf("selected choice is : %d\n", choice)

	parser := func() {
		for {
			input, _, err := reader.ReadLine()
			if err != nil {
				return
			}
			text := strings.TrimSpace(string(input))
			if text == "exit" {
				return
			}
			if selectedClient, exists := m[choice]; exists {
				selectedClient.Send([]byte(text + "\n"))
			}
		}
	}
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		parser()
	}()
}

func (p *P2PServer) waitForShutdown(ctx context.Context) {
	select {
	case <-p.errorChan:
	case <-ctx.Done():
	}
}

func (p *P2PServer) Shutdown() {
	p.peersMutex.Lock()
	for _, peer := range p.peers {
		peer.Close()
	}
	p.peersMutex.Unlock()
	p.wg.Wait()
	p.server.Shutdown()
	close(p.errorChan)
}
