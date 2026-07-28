package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/pion/webrtc/v4"
)

type WebRTCClient struct {
	signalingURL         string
	errorChan            chan error
	pc                   *webrtc.PeerConnection
	dc                   *webrtc.DataChannel
	connMutex            sync.RWMutex
	wg                   *sync.WaitGroup
	onMsgReceivedHandler func([]byte)
	msgParseHandler      func(net.Conn)
	openCh               chan struct{}
}

func NewWebRTCClient(signalingURL string) *WebRTCClient {
	return &WebRTCClient{
		signalingURL:         signalingURL,
		wg:                   &sync.WaitGroup{},
		errorChan:            make(chan error, 1),
		onMsgReceivedHandler: func(b []byte) {},
		msgParseHandler:      func(net.Conn) {},
		openCh:               make(chan struct{}),
	}
}

func (c *WebRTCClient) OnMessageReceivedHandler(handler func([]byte)) {
	c.onMsgReceivedHandler = handler
}

func (c *WebRTCClient) OnMessageParseHandler(handler func(net.Conn)) {
	c.msgParseHandler = handler
}

func (c *WebRTCClient) Connect(ctx context.Context) error {
	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return err
	}

	dc, err := pc.CreateDataChannel("data", nil)
	if err != nil {
		pc.Close()
		return err
	}

	dc.OnOpen(func() {
		log.Printf("connected to WebRTC server %s", c.signalingURL)
		close(c.openCh)
	})
	dc.OnMessage(func(msg webrtc.DataChannelMessage) {
		c.onMsgReceivedHandler(msg.Data)
	})

	offer, err := pc.CreateOffer(nil)
	if err != nil {
		pc.Close()
		return err
	}

	gatherDone := webrtc.GatheringCompletePromise(pc)
	if err = pc.SetLocalDescription(offer); err != nil {
		pc.Close()
		return err
	}
	<-gatherDone

	body, _ := json.Marshal(pc.LocalDescription())
	resp, err := http.Post(c.signalingURL, "application/json", bytes.NewReader(body))
	if err != nil {
		pc.Close()
		return err
	}
	defer resp.Body.Close()

	var answer webrtc.SessionDescription
	if err = json.NewDecoder(resp.Body).Decode(&answer); err != nil {
		pc.Close()
		return err
	}
	if err = pc.SetRemoteDescription(answer); err != nil {
		pc.Close()
		return err
	}

	c.connMutex.Lock()
	c.pc = pc
	c.dc = dc
	c.connMutex.Unlock()

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.handleShutdown(ctx)
	}()

	select {
	case <-c.openCh:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func (c *WebRTCClient) Send(msg []byte) error {
	c.connMutex.RLock()
	dc := c.dc
	c.connMutex.RUnlock()
	if dc == nil {
		return fmt.Errorf("connection not established")
	}
	return dc.Send(msg)
}

func (c *WebRTCClient) SendError(err error) {
	select {
	case c.errorChan <- err:
	default:
	}
}

func (c *WebRTCClient) GetConnId() string {
	c.connMutex.RLock()
	defer c.connMutex.RUnlock()
	return fmt.Sprintf("%p", c.pc)
}

func (c *WebRTCClient) Close() {
	c.connMutex.Lock()
	if c.dc != nil {
		c.dc.Close()
	}
	if c.pc != nil {
		c.pc.Close()
	}
	c.connMutex.Unlock()
	c.wg.Wait()
	close(c.errorChan)
}

func (c *WebRTCClient) handleShutdown(ctx context.Context) {
	select {
	case <-ctx.Done():
	case <-c.errorChan:
	}
	c.connMutex.Lock()
	if c.dc != nil {
		c.dc.Close()
	}
	if c.pc != nil {
		c.pc.Close()
	}
	c.connMutex.Unlock()
}
