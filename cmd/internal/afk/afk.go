package afk

import (
	"context"
	"fmt"
	"time"

	"github.com/impactj90/lofi-tracker/cmd/internal/tracker"
)

func Run(ctx context.Context) {
	const IDLE_THRESHOLD = time.Minute * 15
	go func() {
		ticker := time.NewTicker(IDLE_THRESHOLD)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				idleTime, err := GetIdleTime()
				if err != nil {
					fmt.Printf("Error getting idle time: %v", err)
					continue
				}

				if idleTime >= IDLE_THRESHOLD {
					tr, branchname, err := tracker.Init()
					if err != nil {
						fmt.Printf("❌ Failed to initialize tracker: %v\n", err)
						return
					}

					defer tr.Close()

					sessionStatus, err := tr.Status()
					if err != nil {
						fmt.Printf("x Failed to get Status: %v\n", err)
					}

					if !sessionStatus.IsPaused {
						err = tr.Pause(true)
						if err != nil {
							fmt.Printf("❌ Failed to pause tracking: %v\n", err)
							continue
						}

						fmt.Printf("You've been idle for %d minutes. Session paused on branch %s", IDLE_THRESHOLD, branchname)
					} else {
						sessionStatus, err = tr.Status()
						if sessionStatus.IsPaused && sessionStatus.IsAfk {
							err := tr.Resume()
							if err != nil {
								fmt.Printf("❌ Failed to resume session: %v\n", err)
								continue
							}

							fmt.Printf("Welcome Back. Resumed tracking on Branch %s", branchname)
						}
					}

				}

			}
		}

	}()
}
