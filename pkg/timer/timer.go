package timer

import (
	"time"
)

var DefaultTimer Timer

func init() {
	DefaultTimer = DefaultTimerImpl{}
}

type Timer interface {
	Now() time.Time
}

type DefaultTimerImpl struct{}

func (DefaultTimerImpl) Now() time.Time {
	return time.Now().UTC()
}

func Now() time.Time {
	return DefaultTimer.Now()
}
