package service_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"uuid"

	"github.com/Opxourc/job-queue/internal/job"
	"github.com/Opxourc/job-queue/internal/service"
)

type testRepository struct {
	mu      sync.Mutex
	stored  job.Job
	started sync.WaitGroup
}

func (r *testRepository) Create(context.Context, job.Request) (job.Job, error) {
	return job.Job{}, errors.New("not implemented")
}

func (r *testRepository) GetByID(context.Context, uuid.UUID) (job.Job, error) {
	return r.stored, nil
}

func (r *testRepository) List(context.Context) ([]job.Job, error) {
	return nil, nil
}

func (r *testRepository) Update(ctx context.Context, current job.Job) error {
	r.mu.Lock()
	r.stored = current
	r.mu.Unlock()
	return nil
}

func (r *testRepository) UpdateStatus(_ context.Context, id uuid.UUID, status job.Status) error {
	r.mu.Lock()
	r.stored.Id = id
	r.stored.Status = status
	r.mu.Unlock()
	r.started.Done()
	return nil
}

func (r *testRepository) Delete(context.Context, uuid.UUID) error {
	return nil
}

func TestRunCanceledContext(t *testing.T) {
	repo := &testRepository{}
	svc := service.NewJobService(repo)
	current := job.Job{Id: uuid.New(), Retries: 2, Status: job.StatusQueued}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	svc.Run(ctx, current)

	repo.mu.Lock()
	defer repo.mu.Unlock()
	if repo.stored.Status != job.StatusCanceled {
		t.Fatalf("status = %q, want %q", repo.stored.Status, job.StatusCanceled)
	}
}

func TestCancelRunningJob(t *testing.T) {
	repo := &testRepository{}
	repo.started.Add(1)
	svc := service.NewJobService(repo)

	current := job.Job{Id: uuid.New(), Retries: 0, Status: job.StatusQueued}
	var done sync.WaitGroup
	done.Add(1)

	go func() {
		svc.Run(t.Context(), current)
		done.Done()
	}()

	repo.started.Wait()
	svc.Cancel(current.Id)
	done.Wait()

	repo.mu.Lock()
	defer repo.mu.Unlock()
	if repo.stored.Status != job.StatusCanceled {
		t.Fatalf("status = %q, want %q", repo.stored.Status, job.StatusCanceled)
	}
}
