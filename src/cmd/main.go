package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pauldin91/wsgo/src/core"
)

func main() {
	// Signal handling setup
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	host := "wss://localhost:6443/ws" //flag.String("host", "wss://localhost:6443/ws", "Server host")

	flag.Parse()

	client := core.NewWsClient(ctx, host)

	<-ctx.Done()
	log.Println("[main] shutdown signal received")

	client.Close()

}
