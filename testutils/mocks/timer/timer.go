package timer

import (
	"time"

	"github.com/tuongaz/go-saas/pkg/timer"
)

var _ timer.Timer = &mockTimer{}

func MockTimer(now time.Time) {
	timer.SetDefaultTimer(&mockTimer{now: now})
}

func ResetTimer() {
	timer.SetDefaultTimer(timer.DefaultTimerImpl{})
}

type mockTimer struct {
	now time.Time
}

func (m *mockTimer) Now() time.Time {
	return m.now
}
