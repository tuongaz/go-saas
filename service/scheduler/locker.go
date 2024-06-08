package scheduler

import (
	"time"

	"github.com/tuongaz/go-saas/pkg/log"
)

const lockID = 123456

func (s *Scheduler) waitToAcquireAdvisoryLock() {
	go func() {
		for {
			if v := s.tryAdvisoryLock(); v {
				if s.isLeader == false {
					log.Info("Acquired scheduler lock, become leader")
				}
				s.isLeader = true
			}
			time.Sleep(30 * time.Second)
		}
	}()
}

func (s *Scheduler) tryAdvisoryLock() bool {
	var success bool
	err := s.app.Store().DB().QueryRow("SELECT pg_try_advisory_lock($1)", lockID).Scan(&success)
	if err != nil {
		log.Error("Error acquiring lock:", err)
		return false
	}
	return success
}

func (s *Scheduler) releaseAdvisoryLock() {
	_, err := s.app.Store().DB().Exec("SELECT pg_advisory_unlock($1)", lockID)
	if err != nil {
		log.Error("Error releasing lock:", err)
	}
}
