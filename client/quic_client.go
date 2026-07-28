package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/quic-go/quic-go"
)

type QuicClient struct {
	address              string
	errorChan            chan error
	conn                 *quic.Conn
	stream               *quic.Stream
	connMutex            sync.RWMutex
	wg                   *sync.WaitGroup
	onMsgReceivedHandler func([]byte)
	msgParseHandler      func(net.Conn)
}

func NewQuicClient(address string) *QuicClient {
	return &QuicClient{
		address:              address,
		wg:                   &sync.WaitGroup{},
		errorChan:            make(chan error, 1),
		onMsgReceivedHandler: func(b []byte) {},
		msgParseHandler:      func(net.Conn) {},
	}
}

func (qs *QuicClient) OnMessageReceivedHandler(handler func([]byte)) {
	qs.onMsgReceivedHandler = handler
}

func (qs *QuicClient) OnMessageParseHandler(handler func(net.Conn)) {
	qs.msgParseHandler = handler
}

func (qs *QuicClient) Connect(ctx context.Context) error {
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-echo-example"},
	}
	conn, err := quic.DialAddr(ctx, qs.address, tlsConf, nil)
	if err != nil {
		return err
	}

	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		conn.CloseWithError(0, "stream open failed")
		return err
	}

	qs.connMutex.Lock()
	qs.conn = conn
	qs.stream = stream
	qs.connMutex.Unlock()

	log.Printf("connected to QUIC server %s", qs.address)

	qs.wg.Add(2)
	go func() {
		defer qs.wg.Done()
		qs.readMessages()
	}()
	go func() {
		defer qs.wg.Done()
		qs.handleShutdown(ctx)
	}()
	return nil
}

func (qs *QuicClient) Send(msg []byte) error {
	qs.connMutex.RLock()
	stream := qs.stream
	qs.connMutex.RUnlock()
	if stream == nil {
		return fmt.Errorf("connection not established")
	}
	_, err := stream.Write(append(msg, '\n'))
	return err
}

func (qs *QuicClient) readMessages() {
	qs.connMutex.RLock()
	stream := qs.stream
	qs.connMutex.RUnlock()

	buf := make([]byte, 4096)
	for {
		n, err := stream.Read(buf)
		if err != nil {
			select {
			case qs.errorChan <- err:
			default:
			}
			return
		}
		qs.onMsgReceivedHandler(buf[:n])
	}
}

func (qs *QuicClient) SendError(err error) {
	select {
	case qs.errorChan <- err:
	default:
	}
}

func (qs *QuicClient) GetConnId() string {
	qs.connMutex.RLock()
	defer qs.connMutex.RUnlock()
	return fmt.Sprintf("%p", qs.conn)
}

func (qs *QuicClient) Close() {
	qs.connMutex.Lock()
	if qs.stream != nil {
		qs.stream.Close()
	}
	if qs.conn != nil {
		qs.conn.CloseWithError(0, "client closed")
	}
	qs.connMutex.Unlock()
	qs.wg.Wait()
	close(qs.errorChan)
}

func (qs *QuicClient) handleShutdown(ctx context.Context) {
	select {
	case <-ctx.Done():
	case <-qs.errorChan:
	}
	qs.connMutex.Lock()
	if qs.stream != nil {
		qs.stream.Close()
	}
	if qs.conn != nil {
		qs.conn.CloseWithError(0, "shutdown")
	}
	qs.connMutex.Unlock()
}
