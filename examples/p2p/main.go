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
	var err error
	peer, err := p2p.NewPeer(":8081", ":8080", "tcp")
	peer.Start(ctx)
	if err = peer.Connect(ctx); err == nil {
		peer.OnMessageReceived(
			func(msg []byte) {
				fmt.Printf("[client] received msg: %s\n", msg)
			},
			func(msg []byte) {
				fmt.Printf("[broadcast] msg: %s\n", msg)
				peer.Broadcast(msg)
			})
	}
	defer peer.Shutdown()
	<-ctx.Done()
}
