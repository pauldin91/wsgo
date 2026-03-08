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
	peer := p2p.NewPeer(":8080", "tcp")
	peer.Start(ctx)
	defer peer.Shutdown()
	<-ctx.Done()
}
