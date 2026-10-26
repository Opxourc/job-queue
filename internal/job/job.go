package job

import (
	"context"
	"time"
	"uuid"

	"github.com/gin-gonic/gin"
)

const (
	StatusSucceeded Status = "succeeded" // Completed successfuly.
	StatusCanceled  Status = "canceled"  // Canceled before completion.
	StatusRunning   Status = "running"   // Currently executing.
	StatusFailed    Status = "failed"    // Failed permanently or exhausted retries.
	StatusQueued    Status = "queued"    // Waiting for a worker.

	MaxLifetime = time.Minute * 3
	MaxRetries  = 5
)

// Status is the state of a Job.
type Status string

// Job is a actively existing job that is in known state.
type Job struct {
	DurationSec time.Duration `db:"duration_sec" json:"duration_sec"`
	Lifetime    time.Duration `db:"lifetime" json:"lifetime"`
	Status      Status        `db:"status" json:"status"`
	Retries     int           `db:"retries" json:"retries"`
	Id          uuid.UUID     `db:"job_id" json:"job_id"`
	Attempts    int           `db:"attempts" json:"attempts"`
}

// JobRequest is the wanted structure needed to create a new job.
type Request struct {
	Retries  int           `json:"retries"`
	Lifetime time.Duration `json:"lifetime"`
}

// Service is an object that can handle lifecycle of Jobs.
type Service interface {
	Run(ctx context.Context, job Job) // Run begins the functionality of the provided job.
	Cancel(id uuid.UUID)              // Cancel sets the status the provided job with the provided ID.
	Repo() Repository                 // Repo returns the reference to the repository.
}

// Handler is an object that can take HTTP requests for Job related operations.
type Handler interface {
	GetByID(c *gin.Context) // GetByID handles HTTP GET requests for retrieving an existing Job.
	Create(c *gin.Context)  // Create handles HTTP POST requests for creating a new Job.
	Cancel(c *gin.Context)  // Cancel handles HTTP POST requests for canceling a running Job.
	List(c *gin.Context)    // List handles HTTP GET requests for retrieving all existing Jobs.
}

// Repository is a object that can handle job operations to a repository.
type Repository interface {
	Create(ctx context.Context, req Request) (Job, error)   // Create makes a new record for a Job in the database.
	GetByID(ctx context.Context, id uuid.UUID) (Job, error) // GetByID finds a Job that matches with the provided ID.
	List(ctx context.Context) ([]Job, error)                // List returns an slice of all Jobs that are currently in the database.
	Update(ctx context.Context, job Job) error              // Update persists the mutable state of a Job.
	Delete(ctx context.Context, id uuid.UUID) error         // Delete removes an existing Job from the database.
}

// NewJob constructs a job and returns it.
// Request values will be clamped so they don't exceed excessive values.
func NewJob(req *Request) *Job {
	return &Job{
		Status:      StatusQueued,
		Lifetime:    min(max(req.Lifetime, 0), MaxLifetime),
		Retries:     min(max(req.Retries, 0), MaxRetries),
		DurationSec: time.Second * 0,
		Attempts:    0,
		Id:          uuid.New(),
	}
}
