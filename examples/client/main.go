package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
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

	host := flag.String("host", ":8081", "Server host")
	protocol := flag.String("protocol", "tcp", "Protocol")
	flag.Parse()

	client, err := client.NewClient(*host, *protocol)
	if err != nil {
		log.Fatalf("invalid protocol %v", err)
	}
	client.OnMessageReceived(func(msg []byte) {
		log.Printf("Received: %s", msg)
	})

	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}

	fmt.Printf("connection: %s\n", client.GetConnId())
	reader := bufio.NewReader(os.Stdin)
	go func() {
		for {
			input, _, err := reader.ReadLine()
			if err != nil {
				client.SendError(err)
				return
			}
			if string(input) == "exit" {
				stop()
				return
			}
			if err := client.Send(input); err != nil {
				log.Printf("Send error: %v", err)
			}
		}
	}()

	<-ctx.Done()

	log.Println("[main] shutdown signal received")

	client.Close()
}
