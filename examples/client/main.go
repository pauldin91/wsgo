package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/pauldin91/wsgo/client"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	host := flag.String("host", ":4443", "Server address")
	proto := flag.String("protocol", "tcp", "Protocol to use: tcp, websocket, quic, webrtc")
	flag.Parse()

	c, err := client.NewClient(*host, *proto)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	c.OnMessageReceivedHandler(func(msg []byte) {
		log.Printf("Received: %s", msg)
	})

	var wg sync.WaitGroup

	wg.Add(2)
	go func() {

		defer wg.Done()
		if err := c.Connect(ctx); err != nil {
			log.Fatalf("failed to connect: %v", err)
		}
	}()

	log.Printf("connected via %s to %s", *proto, *host)
	go func() {
		defer wg.Done()
		reader := bufio.NewReader(os.Stdin)
		for {
			input, _, err := reader.ReadLine()
			if err != nil {
				stop()
				return
			}
			if string(input) == "exit" {
				stop()
				return
			}
			if err := c.Send(input); err != nil {
				log.Printf("send error: %v", err)
			}
		}
	}()

	wg.Wait()

	log.Println("shutdown signal received")
}
