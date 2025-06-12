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

	client := client.NewTcpClient(ctx, *host)
	client.Connect()
	client.ListenForInput(bufio.NewReader(os.Stdin))

	<-ctx.Done()

	log.Println("[main] shutdown signal received")

	client.Close()
}
