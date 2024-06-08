package scheduler

import (
	"fmt"
	"log"
)

const lockID = 123456789

func (s *Scheduler) tryAdvisoryLock() bool {
	var success bool
	err := s.app.Store().DB().QueryRow("SELECT pg_try_advisory_lock($1)", lockID).Scan(&success)
	if err != nil {
		log.Println("Error acquiring lock:", err)
		return false
	}
	return success
}

func (s *Scheduler) releaseAdvisoryLock() {
	fmt.Println("Releasing lock")
	_, err := s.app.Store().DB().Exec("SELECT pg_advisory_unlock($1)", lockID)
	if err != nil {
		log.Println("Error releasing lock:", err)
	}
}
