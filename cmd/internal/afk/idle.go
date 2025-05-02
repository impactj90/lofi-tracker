package afk

import "time"

type IdelTimeProvider interface {
	GetIdleTime() (time.Duration, error)
}

var idleProvider IdelTimeProvider

func GetIdleTime() (time.Duration, error) {
	return idleProvider.GetIdleTime()
}
