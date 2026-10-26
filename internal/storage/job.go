package storage

import (
	"context"
	"uuid"

	"github.com/Opxourc/job-queue/internal/job"
	"github.com/jmoiron/sqlx"
)

// JobRepository implements job.Repository for Job repository operations.
type JobRepository struct {
	db *sqlx.DB
}

// Constructs a JobRepository object.
func NewSqliteJobRepository(db *sqlx.DB) *JobRepository {
	return &JobRepository{
		db: db,
	}
}

func (r *JobRepository) Create(ctx context.Context, req job.Request) (job.Job, error) {
	current := job.NewJob(&req)

	if _, err := r.db.NamedExecContext(ctx,
		`INSERT INTO jobs (job_id, status, retries, lifetime)
		VALUES (:job_id, :status, :retries, :lifetime)`, current); err != nil {
		return job.Job{}, ErrQueryFailed
	}

	return *current, nil
}

func (r *JobRepository) GetByID(ctx context.Context, id uuid.UUID) (job.Job, error) {
	var foundJob job.Job
	if err := r.db.GetContext(ctx, &foundJob,
		"SELECT * FROM jobs WHERE job_id = ? LIMIT 1",
		id.String()); err != nil {
		return job.Job{}, err
	}

	return foundJob, nil
}

func (r *JobRepository) List(ctx context.Context) ([]job.Job, error) {
	var jobs []job.Job
	if err := r.db.SelectContext(ctx, &jobs, "SELECT * FROM jobs"); err != nil {
		return nil, ErrQueryFailed
	}

	return jobs, nil
}

func (r *JobRepository) Update(ctx context.Context, current job.Job) error {
	if _, err := r.db.ExecContext(ctx,
		"UPDATE jobs SET status = ?, attempts = ?, duration_sec = ? WHERE job_id = ?",
		current.Status, current.Attempts, current.DurationSec, current.Id.String(),
	); err != nil {
		return ErrQueryFailed
	}

	return nil
}

func (r *JobRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := r.db.ExecContext(ctx, "DELETE FROM jobs WHERE job_id = ?", id.String()); err != nil {
		return ErrQueryFailed
	}

	return nil
}
