package afk

import (
	"context"
	"fmt"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/impactj90/lofi-tracker/cmd/internal/tracker"
)

func Run(ctx context.Context) {
	const IDLE_THRESHOLD = time.Minute * 14
	isAfkActive := false

	tr, _, err := tracker.Init()
	if err != nil {
		fmt.Printf("❌ Failed to initialize tracker: %v\n", err)
		return
	}

	defer tr.Close()

	ticker := time.NewTicker(IDLE_THRESHOLD)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sessionStatus, err := tr.Status()
			if err != nil {
				fmt.Printf("x Failed to get Status: %v\n", err)
				continue
			}

			idleTime, err := GetIdleTime()
			if err != nil {
				fmt.Printf("Error getting idle time: %v", err)
				continue
			}

			if idleTime >= IDLE_THRESHOLD {
				if !sessionStatus.IsPaused && !isAfkActive {
					err = tr.Pause(true)
					if err != nil {
						fmt.Printf("❌ Failed to pause tracking: %v\n", err)
						continue
					}

					beeep.Notify("Lofi Tracker", "You've been paused due to inactivity. Working Session is Paused", "")
					isAfkActive = true
				}
			} else {
				fmt.Printf("should resume... isPaused: %b, isAfk: %b, isAfkActive: %b", sessionStatus.IsPaused, sessionStatus.IsAfk, isAfkActive)
				if sessionStatus.IsPaused && sessionStatus.IsAfk && isAfkActive {
					err := tr.Resume()
					if err != nil {
						fmt.Printf("❌ Failed to resume session: %v\n", err)
						continue
					}

					beeep.Notify("Lofi Tracker", "Welcome Back! Tracking resumed", "")
					isAfkActive = false
				}
			}

		}
	}

}
