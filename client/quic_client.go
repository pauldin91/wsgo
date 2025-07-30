package client

import (
	"context"
	"net"

	"github.com/quic-go/quic-go"
)

type QuicClient struct {
	ctx  context.Context
	conn *quic.Conn
}

func NewQuicClient() *QuicClient {
	return &QuicClient{}
}

func (qs *QuicClient) Connect() error {
	// var wg *sync.WaitGroup = &sync.WaitGroup{}
	// addr := "localhost:3000"
	go func() {
		/*
			rsp, err := hclient.Get(addr)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Got response for %s: %#v", addr, rsp)

			body := &bytes.Buffer{}
			_, err = io.Copy(body, rsp.Body)
			if err != nil {
				log.Fatal(err)
			}

			log.Printf("Response Body (%d bytes):\n%s", body.Len(), body.Bytes())

			wg.Done()*/
	}()
	return nil
}

func (qs *QuicClient) Close() {

}
func (qs *QuicClient) GetConnId() string {
	return ""
}
func (qs *QuicClient) OnMessageReceivedHandler(func([]byte)) {

}
func (qs *QuicClient) OnMessageParseHandler(func(net.Conn)) {

}
func (qs *QuicClient) SendError(err error) {

}
