package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pauldin91/wsgo/p2p"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	peer := p2p.NewPeer(":8081", "tcp")
	var err error
	peer.Start(ctx)
	if err = peer.Connect(ctx, ":8080"); err == nil {
		peer.OnMessageReceivedByClient(func(msg []byte) {
			fmt.Printf("[client] received msg: %s\n", msg)
			peer.Broadcast(msg)
		})
	}
	defer peer.Shutdown()
	<-ctx.Done()
}
