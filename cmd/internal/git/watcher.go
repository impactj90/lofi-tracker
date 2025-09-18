package git

import (
	"context"
	"fmt"
	"time"
)

const defaultBranchPollInterval = 5 * time.Second

// BranchChanged represents a branch transition event detected by the watcher.
type BranchChanged struct {
	From string
	To   string
	At   time.Time
}

type branchResolver func(context.Context) (string, error)

// BranchWatcher polls the current git branch and emits events when it changes.
type BranchWatcher struct {
	Interval time.Duration
	Out      chan BranchChanged

	resolve branchResolver
}

// NewBranchWatcher creates a new watcher with the provided interval. When interval
// is zero it falls back to a sensible default.
func NewBranchWatcher(interval time.Duration) *BranchWatcher {
	if interval <= 0 {
		interval = defaultBranchPollInterval
	}

	return &BranchWatcher{
		Interval: interval,
		Out:      make(chan BranchChanged, 1),
		resolve: func(ctx context.Context) (string, error) {
			return resolveBranchLabel(ctx)
		},
	}
}

// Start begins polling until the context is cancelled. The output channel is
// closed on exit.
func (w *BranchWatcher) Start(ctx context.Context) error {
	defer close(w.Out)

	current := ""

	// Emit an immediate event to ensure automation picks up the active branch
	// without waiting for the first tick.
	now := time.Now()
	label, err := w.resolve(ctx)
	if err != nil {
		// Without a label we cannot proceed; surface the error so the caller can
		// decide how to handle the failure.
		return fmt.Errorf("resolve branch label: %w", err)
	}
	select {
	case w.Out <- BranchChanged{From: current, To: label, At: now}:
		current = label
	case <-ctx.Done():
		return ctx.Err()
	}

	ticker := time.NewTicker(w.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			branch, err := w.resolve(ctx)
			if err != nil {
				// Continue polling; transient git errors should not stop the watcher.
				fmt.Printf("branch watcher error: %v\n", err)
				continue
			}

			if branch == current {
				continue
			}

			change := BranchChanged{From: current, To: branch, At: time.Now()}
			select {
			case w.Out <- change:
				current = branch
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

// Events exposes the channel of branch change notifications.
func (w *BranchWatcher) Events() <-chan BranchChanged {
	return w.Out
}
