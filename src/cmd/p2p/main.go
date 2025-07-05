package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/pauldin91/wsgo/src/p2p"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	p2pServer := p2p.NewP2PServer(ctx, "localhost:6446")
	p2pServer.Start()
	p2pServer.ExposeFirstPeerForInput()
	<-ctx.Done()
	p2pServer.Shutdown()
}
