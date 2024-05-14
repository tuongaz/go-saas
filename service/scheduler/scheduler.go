package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/tuongaz/go-saas/app"
	"github.com/tuongaz/go-saas/pkg/log"
)

var _ Interface = &Scheduler{}

type Interface interface {
	NewDurationJob(d time.Duration, job func(), tags ...string) (id string, err error)
	NewCronJobWithMinutes(cron string, job func(), tags ...string) (id string, err error)
	NewCronJobWithSeconds(cron string, job func(), tags ...string) (id string, err error)
	NewOneTimeJob(t time.Time, job func(), tags ...string) (id string, err error)
	RemoveJob(id string) error
	RemoveJobByTags(tags ...string)
}

func Register(appInstance app.Interface) {
	s := &Scheduler{}

	appInstance.OnAfterBootstrap().Add(func(ctx context.Context, e *app.OnAfterBootstrapEvent) error {
		if err := s.bootstrap(); err != nil {
			return fmt.Errorf("scheduler bootstrap: %w", err)
		}

		return nil
	})

	appInstance.OnBeforeServe().Add(func(ctx context.Context, e *app.OnBeforeServeEvent) error {
		if err := s.start(ctx); err != nil {
			return fmt.Errorf("scheduler start: %w", err)
		}

		return nil
	})
}

type Scheduler struct {
	scheduler gocron.Scheduler
}

func (s *Scheduler) bootstrap() error {
	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return fmt.Errorf("new scheduler: %w", err)
	}

	s.scheduler = scheduler
	return nil
}

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

func (s *Scheduler) RemoveJobByTags(tags ...string) {
	s.scheduler.RemoveByTags(tags...)
}

func (s *Scheduler) start(ctx context.Context) error {
	log.Info("starting scheduler")
	s.scheduler.Start()
	return nil
}

func (s *Scheduler) NewDurationJob(d time.Duration, job func(), tags ...string) (id string, err error) {
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

func (s *Scheduler) NewCronJobWithMinutes(cron string, job func(), tags ...string) (id string, err error) {
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

func (s *Scheduler) NewCronJobWithSeconds(cron string, job func(), tags ...string) (id string, err error) {
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

func (s *Scheduler) NewOneTimeJob(t time.Time, job func(), tags ...string) (id string, err error) {
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
