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

	host := flag.String("host", ":4443", "Server host")

	flag.Parse()

	client := client.NewTcpClient(ctx, *host)
	client.Connect()
	reader := bufio.NewReader(os.Stdin)
	client.ListenForInput(reader)
	// client.Input(reader, func(r *bufio.Reader) {
	// 	func() {
	// 		for {
	// 			input, _, err := reader.ReadLine()
	// 			if err != nil {
	// 				fmt.Println("[console] Read error:", err)
	// 				return
	// 			}

	// 			text := strings.TrimSpace(string(input))
	// 			if text == "exit" {
	// 				fmt.Println("[console] Exiting by user command.")
	// 				return
	// 			}
	// 			client.Send(text)
	// 		}
	// 	}()
	// })

	<-ctx.Done()

	log.Println("[main] shutdown signal received")

	client.Close()
}
