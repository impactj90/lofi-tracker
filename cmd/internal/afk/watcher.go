package afk

import "context"


type Watcher interface {
	Start(ctx context.Context) error
}

