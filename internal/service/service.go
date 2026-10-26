package service

import (
	"context"
	"log"
	"math/rand/v2"
	"sync"
	"time"
	"uuid"

	"github.com/Opxourc/job-queue/internal/job"
)

// JobService implements job.Service and holds a reference to job.Repository.
type JobService struct {
	repo   job.Repository
	mu     sync.Mutex                       // Prevents two gorountines from accessing the cancel map.
	cancel map[uuid.UUID]context.CancelFunc // Associated cancel function for a specific job/context.
}

// NewJobService constructs a JobService object.
func NewJobService(r job.Repository) *JobService {
	return &JobService{
		repo:   r,
		cancel: make(map[uuid.UUID]context.CancelFunc),
	}
}

func (s *JobService) Repo() job.Repository {
	return s.repo
}

func (s *JobService) Cancel(id uuid.UUID) {
	s.mu.Lock()
	cancel, ok := s.cancel[id]
	s.mu.Unlock()
	if ok {
		cancel()
	}
}

func (s *JobService) Run(ctx context.Context, current job.Job) {
	runCtx, cancel := context.WithCancel(ctx)

	s.mu.Lock()
	s.cancel[current.Id] = cancel
	s.mu.Unlock()

	defer func() {
		cancel()

		s.mu.Lock()
		delete(s.cancel, current.Id)
		s.mu.Unlock()
	}()

	persistCtx := context.WithoutCancel(ctx)
	if runCtx.Err() != nil {
		current.Status = job.StatusCanceled
		if err := s.repo.Update(persistCtx, current); err != nil {
			log.Printf("persist canceled job %s: %v", current.Id, err)
		}
		return
	}

	current.Status = job.StatusRunning
	if err := s.repo.Update(persistCtx, current); err != nil {
		log.Printf("mark job %s as running: %v", current.Id, err)
		return
	}

	succeeded := false

	for current.Attempts <= current.Retries && !succeeded && runCtx.Err() == nil {
		duration := time.Duration(rand.IntN(10)+1) * time.Second
		timer := time.NewTimer(duration)

		select {
		case <-timer.C:
			current.DurationSec += duration
			current.Attempts++

			succeeded = rand.IntN(2) == 1
		case <-runCtx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
		}
	}

	switch {
	case runCtx.Err() != nil:
		current.Status = job.StatusCanceled
	case succeeded:
		current.Status = job.StatusSucceeded
	default:
		current.Status = job.StatusFailed
	}

	if err := s.repo.Update(persistCtx, current); err != nil {
		log.Printf("persist completed job %s: %v", current.Id, err)
	}
}
