package afk

import (
	"context"
)

type Daemon struct {
	Afk *AfkWatcher
}

func (d *Daemon) Run(ctx context.Context) {
	go d.Afk.Start(ctx)

	<-ctx.Done()
}

