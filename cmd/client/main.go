package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/pauldin91/wsgo/client"
)

func main() {
	// Signal handling setup
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	host := flag.String("host", ":4443", "Server host")

	flag.Parse()

	client := client.NewTcpClient(ctx, *host)
	client.Connect()
	client.HandleInputFrom(os.Stdin, func(src *os.File) {

		go func() {
			reader := bufio.NewReader(src)
			for {
				input, _, err := reader.ReadLine()
				if err != nil {
					client.SendError(err)
					return
				}
				text := strings.TrimSpace(string(input))
				if text == "exit" {
					return
				}
				client.Send(text)
			}
		}()
	})

	<-ctx.Done()

	log.Println("[main] shutdown signal received")

	client.Close()
}
