package scheduler

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron/v2"
)

var _ Interface = &Scheduler{}

type Interface interface {
	NewDurationJob(d time.Duration, job func()) (id string, err error)
	NewCronJobWithMinutes(cron string, job func()) (id string, err error)
	NewCronJobWithSeconds(cron string, job func()) (id string, err error)
	NewOneTimeJob(t time.Time, job func()) (id string, err error)
}

func New() (*Scheduler, error) {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return nil, fmt.Errorf("new scheduler: %w", err)
	}

	return &Scheduler{
		scheduler: scheduler,
	}, nil
}

type Scheduler struct {
	scheduler gocron.Scheduler
}

func (s *Scheduler) Start() {
	s.scheduler.Start()
}

func (s *Scheduler) NewDurationJob(d time.Duration, job func()) (id string, err error) {
	j, err := s.scheduler.NewJob(gocron.DurationJob(d), gocron.NewTask(job))
	if err != nil {
		return "", fmt.Errorf("new duration job: %w", err)
	}

	return j.ID().String(), nil
}

func (s *Scheduler) NewCronJobWithMinutes(cron string, job func()) (id string, err error) {
	j, err := s.scheduler.NewJob(gocron.CronJob(cron, true), gocron.NewTask(job))
	if err != nil {
		return "", fmt.Errorf("new cron job: %w", err)
	}

	return j.ID().String(), nil
}

func (s *Scheduler) NewCronJobWithSeconds(cron string, job func()) (id string, err error) {
	j, err := s.scheduler.NewJob(gocron.CronJob(cron, true), gocron.NewTask(job))
	if err != nil {
		return "", fmt.Errorf("new cron job: %w", err)
	}

	return j.ID().String(), nil
}

func (s *Scheduler) NewOneTimeJob(t time.Time, job func()) (id string, err error) {
	j, err := s.scheduler.NewJob(gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(t)), gocron.NewTask(job))
	if err != nil {
		return "", fmt.Errorf("new one time job: %w", err)
	}

	return j.ID().String(), nil
}
