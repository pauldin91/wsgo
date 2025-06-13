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

	host := flag.String("host", ":6443", "Server host")

	flag.Parse()

	client := client.NewWsClient(ctx, *host)
	client.Connect()
	reader := bufio.NewReader(os.Stdin)
	client.ListenForInput(reader)

	<-ctx.Done()

	log.Println("[main] shutdown signal received")

	client.Close()
}
