package storage

import (
	"context"
	"errors"
	"uuid"

	"github.com/Opxourc/job-queue/internal/job"
)

var (
	ErrQueryFailed        = errors.New("query failed to complete")
	ErrJobAlreadyExists   = errors.New("job already exists")
	ErrJobDoesNotExist    = errors.New("job doesn't exist")
	ErrInvalidJobProvided = errors.New("provided job is invalid")
)

// JobRepository is a object that can handle job operations to a repository.
type JobRepository interface {
	Create(ctx context.Context, req *job.JobRequest) (*job.Job, error)
	GetByID(ctx context.Context, id uuid.UUID) (*job.Job, error)
	List(ctx context.Context) ([]*job.Job, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status job.JobStatus) error
	Delete(ctx context.Context, id uuid.UUID) error
}
