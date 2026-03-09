package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/pauldin91/wsgo/p2p"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	peer := p2p.NewPeer(":8081", "tcp")
	peer.OnMessageReceivedByClient(func(b []byte) {
		peer.Broadcast(b)
	})
	peer.Connect(ctx, ":8080")
	peer.Start(ctx)
	defer peer.Shutdown()
	<-ctx.Done()
}
