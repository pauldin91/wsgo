package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pauldin91/wsgo/internal"
	"github.com/pauldin91/wsgo/p2p"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	host := flag.String("host", ":4443", "Server host")
	peers := flag.String("peers", "", "peers")
	flag.Parse()
	fmt.Println(*peers)
	fmt.Println(*host)

	p2pServer := p2p.NewP2PServer(ctx, *host, internal.TCP)
	p2pServer.Start()
	p2pServer.Connect(*peers)
	p2pServer.SetMsgReceivedHandler(func(b []byte) {
		fmt.Printf("%s", []byte("Echo: "+string(b)+"\n"))
	})
	p2pServer.MsgReceivedHandler(func(b []byte) {
		fmt.Printf("%s", []byte("Echo: "+string(b)+"\n"))
	})
	p2pServer.OnParseMsgHandler(os.Stdin)
	<-ctx.Done()
	p2pServer.Shutdown()
}
