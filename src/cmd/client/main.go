package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pauldin91/wsgo/src/client"
)

func main() {
	// Signal handling setup
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	host := "ws://localhost:6443/ws" //flag.String("host", "wss://localhost:6443/ws", "Server host")

	flag.Parse()

	client := client.NewWsClient(ctx, host)
	client.Connect()
	client.ListenForInput(bufio.NewReader(os.Stdin))

	<-ctx.Done()

	log.Println("[main] shutdown signal received")

	client.Close()
}
