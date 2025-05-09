package afk

import (
	"context"

	"github.com/impactj90/lofi-tracker/cmd/internal/tracker"
)

var _ Watcher = (*BranchWatcher)(nil)

type BranchWatcher struct {
}

func (b *BranchWatcher) Start(ctx context.Context) error {
	tr, branchname, err := tracker.Init()
	//TODO: make a loop to check if branch has changed, for an intervall with every second?
	return nil
}

