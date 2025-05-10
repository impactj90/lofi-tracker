package afk

import (
	"context"
)

type Daemon struct {
	Afk *AfkWatcher
	BranchWatcher *BranchWatcher
}

func (d *Daemon) Run(ctx context.Context) {
	go d.Afk.Start(ctx)
	go d.BranchWatcher.Start(ctx)

	<-ctx.Done()
}

