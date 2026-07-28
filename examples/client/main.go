package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
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

	if err := c.Connect(ctx); err != nil {
		log.Fatalf("failed to connect: %v", err)
	}

	log.Printf("connected via %s to %s", *proto, *host)

	reader := bufio.NewReader(os.Stdin)
	go func() {
		for {
			input, _, err := reader.ReadLine()
			if err != nil {
				c.SendError(err)
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

	<-ctx.Done()
	log.Println("shutdown signal received")
	c.Close()
}
