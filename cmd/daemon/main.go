package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/impactj90/lofi-tracker/cmd/internal/afk"
	"github.com/impactj90/lofi-tracker/cmd/internal/tracker"
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

	tr, _, err := tracker.Init()
	if err != nil {
		fmt.Printf("Error initializing Tracker: %v", err)
		os.Exit(1)
	}

	daemon := afk.Daemon{
		Afk: &afk.AfkWatcher{
			Tracker:       tr,
			IdleThreshold: time.Second * 15,
			IsAfkActive:   false,
		},
	}

	daemon.Run(ctx)

	<-ctx.Done()
	fmt.Println("lofi-tracker daemon exited")
}
