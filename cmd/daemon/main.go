package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/impactj90/lofi-tracker/cmd/internal/afk"
	"github.com/impactj90/lofi-tracker/cmd/internal/db"
	"github.com/impactj90/lofi-tracker/cmd/internal/git"
	"github.com/impactj90/lofi-tracker/cmd/internal/tracker"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		fmt.Printf("Received shutdown signal %s\n", sig.String())
		cancel()
	}()

	tr, _, err := tracker.Init()
	if err != nil {
		fmt.Printf("Error initializing Tracker: %v\n", err)
		os.Exit(1)
	}
	defer tr.Close()

	afkWatcher := &afk.AfkWatcher{
		Tracker:       tr,
		IdleThreshold: 15 * time.Minute,
		IsAfkActive:   false,
	}

	branchWatcher := git.NewBranchWatcher(0)
	branchEvents := branchWatcher.Events()

	afkErrCh := make(chan error, 1)
	branchErrCh := make(chan error, 1)

	go func() {
		afkErrCh <- afkWatcher.Start(ctx)
	}()

	go func() {
		branchErrCh <- branchWatcher.Start(ctx)
	}()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("lofi-tracker daemon exited")
			return
		case err := <-afkErrCh:
			if err != nil && !errors.Is(err, context.Canceled) {
				fmt.Printf("AFK watcher error: %v\n", err)
			}
			afkErrCh = nil
		case err := <-branchErrCh:
			if err != nil {
				fmt.Printf("Branch watcher error: %v\n", err)
				if !errors.Is(err, context.Canceled) {
					beeep.Notify("Lofi Tracker", "Branch automation stopped. Check repository status.", "")
				}
			}
			branchErrCh = nil
		case event, ok := <-branchEvents:
			if !ok {
				branchEvents = nil
				continue
			}

			result, err := tr.ForceStart(event.To)
			if err != nil {
				fmt.Printf("Failed to sync branch session: %v\n", err)
				if pauseErr := tr.Pause(false); pauseErr != nil && !errors.Is(pauseErr, db.ErrNoActiveSession) {
					fmt.Printf("Failed to pause session after error: %v\n", pauseErr)
				}
				beeep.Notify("Lofi Tracker", "Tracking paused due to branch sync error.", "")
				continue
			}

			if result.Previous != nil && event.From != "" {
				duration := tracker.FormatSummaryDuration(result.Previous.TotalDuration)
				message := fmt.Sprintf("Completed %s after %s", event.From, duration)
				beeep.Notify("Lofi Tracker", message, "")
			}

			fmt.Printf("Tracking branch %s\n", result.Current.Branch)
		}
	}
}
