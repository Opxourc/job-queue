package storage

import (
	"context"
	"uuid"

	"github.com/Opxourc/job-queue/internal/job"
	"github.com/jmoiron/sqlx"
)

// SqliteJobRepository references the database and implemented JobRepository.
type SqliteJobRepository struct {
	db *sqlx.DB
}

// Constructs a SqliteJobRepository object.
func NewSqliteJobRepository(db *sqlx.DB) *SqliteJobRepository {
	return &SqliteJobRepository{
		db: db,
	}
}

// Create makes a new record for a new job in the repository.
// Presistence of jobs takes place here, not construction!
func (r *SqliteJobRepository) Create(ctx context.Context, new *job.Job) error {
	if new == nil {
		return ErrInvalidJobProvided
	}

	if existing, err := r.GetByID(ctx, new.Id); err != nil {
		return err
	} else if existing != nil {
		return ErrJobAlreadyExists
	}

	if _, err := r.db.NamedExecContext(ctx,
		`INSERT INTO jobs (job_id, status, retries, lifetime)
		VALUES (:job_id, :status, :retries, :lifetime)`, new); err != nil {
		return ErrQueryFailed
	}

	return nil
}

// GetByID finds a job that matches with the provided ID.
func (r *SqliteJobRepository) GetByID(ctx context.Context, id uuid.UUID) (*job.Job, error) {
	var foundJob job.Job
	if err := r.db.GetContext(ctx, &foundJob,
		"SELECT * FROM jobs WHERE job_id = ?",
		id.String()); err != nil {
		return nil, err
	}

	return &foundJob, nil
}

// List returns an slice of all jobs that are currently in the repository.
func (r *SqliteJobRepository) List(ctx context.Context) ([]*job.Job, error) {
	var jobs []*job.Job
	if err := r.db.SelectContext(ctx, &jobs, "SELECT * FROM jobs"); err != nil {
		return nil, ErrQueryFailed
	}

	return jobs, nil
}

// UpdateStatus updates the status of a job.
func (r *SqliteJobRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status job.JobStatus) error {
	if _, err := r.db.ExecContext(ctx,
		"UPDATE jobs SET status = ? WHERE job_id = ?",
		status, id.String(),
	); err != nil {
		return ErrQueryFailed
	}

	return nil
}

// Delete removes an existing job from the repository.
func (r *SqliteJobRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := r.db.ExecContext(ctx, "DELETE FROM jobs WHERE job_id = ?", id.String()); err != nil {
		return ErrQueryFailed
	}

	return nil
}
