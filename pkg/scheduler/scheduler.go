package scheduler

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

type Scheduler struct {
	cron *cron.Cron
}

func New() *Scheduler {
	return &Scheduler{
		cron: cron.New(),
	}
}

func (s *Scheduler) AddJob(spec string, job func()) error {
	_, err := s.cron.AddFunc(spec, job)
	return err
}

func (s *Scheduler) AddIntervalJob(hours int, job func()) error {
	spec := fmt.Sprintf("@every %dh", hours)
	return s.AddJob(spec, job)
}

func (s *Scheduler) Start() {
	s.cron.Start()
}

func (s *Scheduler) Stop() {
	ctx := s.cron.Stop()
	select {
	case <-ctx.Done():
	case <-time.After(5 * time.Second):
	}
}
