package afk

import (
	"context"
	"fmt"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/impactj90/lofi-tracker/cmd/internal/tracker"
)

var _ Watcher = (*AfkWatcher)(nil)

type AfkWatcher struct {
	Tracker       tracker.Tracker
	IdleThreshold time.Duration
	IsAfkActive   bool
}

func (a *AfkWatcher) Start(ctx context.Context) error {
	tr, _, err := tracker.Init()
	if err != nil {
		fmt.Printf("❌ Failed to initialize tracker: %v\n", err)
		return err
	}

	defer tr.Close()

	ticker := time.NewTicker(a.IdleThreshold)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := a.detectAfk(ctx); err != nil {
				fmt.Printf("%v", err)
				continue
			}
		}
	}
}

func (a *AfkWatcher) detectAfk(ctx context.Context) error {
	sessionStatus, err := a.Tracker.Status()
	if err != nil {
		return fmt.Errorf("x Failed to get Status: %v\n", err)
	}

	idleTime, err := GetIdleTime()
	if err != nil {
		return fmt.Errorf("Error getting idle time: %v", err)
	}

	if idleTime >= a.IdleThreshold && !sessionStatus.IsAfk && !a.IsAfkActive {
		err = a.Tracker.Pause(true)
		if err != nil {
			return fmt.Errorf("❌ Failed to pause tracking: %v\n", err)
		}

		beeep.Notify("Lofi Tracker", "You've been paused due to inactivity. Working Session is Paused", "")
		a.IsAfkActive = true
		return a.watchForResume(ctx)
	}
	return nil
}

func (a *AfkWatcher) watchForResume(ctx context.Context) error {
	resumeThreshold := time.Second * 2
	ticker := time.NewTicker(resumeThreshold)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			idleTime, err := GetIdleTime()
			if err != nil {
				return fmt.Errorf("Error getting idle time: %v", err)
			}

			if idleTime < resumeThreshold {
				err := a.Tracker.Resume()
				if err != nil {
					return fmt.Errorf("❌ Failed to resume session: %v\n", err)
				}

				beeep.Notify("Lofi Tracker", "Welcome Back! Tracking resumed", "")
				a.IsAfkActive = false

				return nil
			}
		}
	}
}
