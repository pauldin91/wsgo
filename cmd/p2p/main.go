package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/pauldin91/wsgo/p2p"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	host := flag.String("host", ":4443", "Server host address")
	peers := flag.String("peers", "", "Comma-separated list of peer addresses to connect to")
	flag.Parse()

	p2pServer, err := p2p.NewP2PServer(ctx, *host, "tcp")
	if err != nil {
		fmt.Printf("Failed to create P2P server: %v\n", err)
		os.Exit(1)
	}
	p2pServer.Start(ctx)
	if *peers != "" {
		if err := p2pServer.Connect(strings.Split((*peers), ",")); err != nil {
			fmt.Printf("Failed to connect to peers: %v\n", err)
		}
	}
	p2pServer.SetMsgReceivedHandler(func(b []byte) {
		fmt.Printf("Echo: %s\n", b)
	})
	p2pServer.MsgReceivedHandler(func(b []byte) {
		fmt.Printf("Echo: %s\n", b)
	})
	p2pServer.OnParseMsgHandler(os.Stdin)
	<-ctx.Done()
	p2pServer.Shutdown()
}
