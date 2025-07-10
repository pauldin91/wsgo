package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pauldin91/wsgo/client"
)

func main() {
	// Signal handling setup
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	host := flag.String("host", "ws://localhost:4443/ws", "Server host")

	flag.Parse()

	client := client.NewWsClient(ctx, *host)
	client.Connect()

	client.OnParseMsgHandler(os.Stdin)

	<-ctx.Done()

	log.Println("[main] shutdown signal received")

	client.Close()
}
