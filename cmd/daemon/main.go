package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/impactj90/lofi-tracker/cmd/internal/afk"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		fmt.Printf("Received shutdown signal %s", sig.String())
		cancel()
	}()

	afk.Run(ctx)

	<-ctx.Done()
	fmt.Println("lofi-tracker daemon exited")
}

