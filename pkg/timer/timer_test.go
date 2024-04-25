package timer

import (
	"testing"
	"time"
)

type mockTimer struct {
	currentTime time.Time
}

func (m mockTimer) Now() time.Time {
	return m.currentTime
}

func TestNow(t *testing.T) {
	expectedTime := time.Date(2024, 1, 17, 10, 0, 0, 0, time.UTC)
	DefaultTimer = mockTimer{currentTime: expectedTime}

	actualTime := Now()

	if !actualTime.Equal(expectedTime) {
		t.Errorf("Now() = %v; want %v", actualTime, expectedTime)
	}
}

func TestDefaultTimerImpl_Now(t *testing.T) {
	timer := DefaultTimerImpl{}
	now := timer.Now()
	currentTime := time.Now().UTC()

	if now.Sub(currentTime) > time.Second {
		t.Errorf("DefaultTimerImpl Now() is not accurate, got %v, want around %v", now, currentTime)
	}
}
