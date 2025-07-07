package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/pauldin91/wsgo/internal"
	"github.com/pauldin91/wsgo/p2p"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	p2pServer := p2p.NewP2PServer(ctx, "localhost:6446", internal.TCP)
	p2pServer.Start()
	p2pServer.SetOnMsgReceivedHandler(func(c net.Conn, b []byte) {
		c.Write([]byte("Echo: " + string(b) + "\n"))
	})
	<-ctx.Done()
	p2pServer.Shutdown()
}
