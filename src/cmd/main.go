package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/pauldin91/wsgo/src/core"
)

func main() {
	// Signal handling setup
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	host := flag.String("host", "wss://localhost:6443/ws", "Server host")

	flag.Parse()

	client := core.NewWsClient(ctx, *host)

	go func() {
		sig := <-sigChan
		fmt.Printf("\nReceived signal: %s. Exiting...\n", sig)
		cancel()
	}()

	go client.Close()

	readFromConsole(ctx, client)

}

func readFromConsole(ctx context.Context, client *core.WsClient) {
	// Signal handling setup
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Input reader setup
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Type something (Ctrl+C to exit):")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Exiting...")
			return

		default:
			input, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("Read error:", err)
				return
			}

			text := strings.TrimSpace(input)
			if text == "exit" {
				fmt.Println("Exiting by user command.")
				return
			}

			client.Send(text)
		}
	}
}
