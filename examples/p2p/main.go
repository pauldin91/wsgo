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

	peer.Connect(ctx, ":8080")
	peer.OnMessageReceivedByClient(func(msg []byte) {

		fmt.Printf("Received msg %s as client", msg)
	})
	peer.OnMessageReceived(func(b []byte) {
		peer.Broadcast(b)
	})
	peer.Start(ctx)
	defer peer.Shutdown()
	<-ctx.Done()
}
