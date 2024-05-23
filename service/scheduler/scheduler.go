package scheduler

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
)

var _ Interface = &Scheduler{}

type Interface interface {
	RunEvery(d time.Duration, job func(), tags ...string) (id string, err error)
	RunCron(cron string, job func(), tags ...string) (id string, err error)
	RunAt(t time.Time, job func(), tags ...string) (id string, err error)
	RemoveJob(id string) error
	RemoveJobByTags(tags ...string)
}

type Scheduler struct {
	scheduler gocron.Scheduler
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

func MustNew() *Scheduler {
	s, err := New()
	if err != nil {
		panic(fmt.Errorf("failed to create a new scheduler: %w", err))
	}
	return s
}

func (s *Scheduler) Start() {
	s.scheduler.Start()
}

// RemoveJob removes a job by its id
func (s *Scheduler) RemoveJob(id string) error {
	uid, err := s.uuid(id)
	if err != nil {
		return err
	}

	if err := s.scheduler.RemoveJob(uid); err != nil {
		return fmt.Errorf("remove job: %w", err)
	}
	return nil
}

// RemoveJobByTags removes jobs by their tags
func (s *Scheduler) RemoveJobByTags(tags ...string) {
	s.scheduler.RemoveByTags(tags...)
}

// RunEvery runs a job every d duration
func (s *Scheduler) RunEvery(d time.Duration, job func(), tags ...string) (id string, err error) {
	j, err := s.scheduler.NewJob(
		gocron.DurationJob(d),
		gocron.NewTask(job),
		gocron.WithTags(tags...),
	)
	if err != nil {
		return "", fmt.Errorf("new duration job: %w", err)
	}

	return j.ID().String(), nil
}

// RunCron runs a job every cron duration
func (s *Scheduler) RunCron(cron string, job func(), tags ...string) (id string, err error) {
	j, err := s.scheduler.NewJob(
		gocron.CronJob(cron, true),
		gocron.NewTask(job),
		gocron.WithTags(tags...),
	)
	if err != nil {
		return "", fmt.Errorf("new cron job: %w", err)
	}

	return j.ID().String(), nil
}

// RunAt runs a job at a specific time (UTC)
func (s *Scheduler) RunAt(t time.Time, job func(), tags ...string) (id string, err error) {
	j, err := s.scheduler.NewJob(
		gocron.OneTimeJob(gocron.OneTimeJobStartDateTime(t)),
		gocron.NewTask(job),
		gocron.WithTags(tags...),
	)
	if err != nil {
		return "", fmt.Errorf("new one time job: %w", err)
	}

	return j.ID().String(), nil
}

func (s *Scheduler) uuid(id string) (uuid.UUID, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("parse uuid: %w", err)
	}
	return uid, nil
}
