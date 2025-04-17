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

func NowString() string {
	return Now().Format(time.RFC3339)
}

func SetDefaultTimer(timer Timer) {
	DefaultTimer = timer
}
