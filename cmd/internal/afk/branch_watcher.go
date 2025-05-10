package afk

import (
	"context"
	"fmt"
	"time"

	"github.com/impactj90/lofi-tracker/cmd/internal/git"
	"github.com/impactj90/lofi-tracker/cmd/internal/tracker"
)

var _ Watcher = (*BranchWatcher)(nil)

type BranchWatcher struct {
	Tracker       tracker.Tracker
	CheckInterval time.Duration
}

func (b *BranchWatcher) Start(ctx context.Context) error {
	tr, _, err := tracker.Init()
	if err != nil {
		fmt.Printf("❌ Failed to initialize tracker: %v\n", err)
		return err
	}

	defer tr.Close()

	ticker := time.NewTicker(b.CheckInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			status, err := tr.Status()
			if err != nil {
				return err
			}

			detectedBranch, err := git.GetCurrentBranchName()
			if err != nil {
				return err
			}

			if status.Branch != detectedBranch {
				if !status.IsPaused {
					if err := tr.Pause(false); err != nil {
						return err
					}
				}

				if err := tr.Start(detectedBranch); err != nil {
					return err
				}

				fmt.Printf("Switched branches. We paused the Session for: %s and started it for: %s", status.Branch, detectedBranch)
			}

		}
	}
}
