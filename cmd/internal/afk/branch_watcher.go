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

func (branchWatcher *BranchWatcher) Start(ctx context.Context) error {
	ticker := time.NewTicker(branchWatcher.CheckInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			detectedBranch, err := git.GetCurrentBranchName()
			if err != nil {
				fmt.Println("%w", err)
				continue
			}

			//This is the SessionStatus of the tracker
			//So this will get the active session only
			status, err := branchWatcher.Tracker.Status()
			if err != nil {
				fmt.Println("%w", err)
				continue
			}

			if status.Branch == detectedBranch {
				//no detected changes
				continue
			}

			if !status.IsPaused {
				if err := branchWatcher.Tracker.Pause(false); err != nil {
					fmt.Println("%w", err)
					continue
				}
			}

			fmt.Printf("%s, %s", status.Branch, detectedBranch)
			_, err = branchWatcher.Tracker.ResumeOrCreateSession(status.Branch)
			if err != nil {
				fmt.Println("error: %w", err)
				continue
			}

			fmt.Printf("Switched branches. We paused the Session for: %s and started it for: %s", status.Branch, detectedBranch)
		}

	}
}
