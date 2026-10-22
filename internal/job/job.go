package job

import (
	"crypto/rand"
	"time"
)

// JobStatus is the state of a Job.
type JobStatus string

const (
	StatusSucceeded JobStatus = "succeeded" // Completed successfuly.
	StatusCanceled  JobStatus = "canceled"  // Canceled before completion.
	StatusRunning   JobStatus = "running"   // Currently executing.
	StatusFailed    JobStatus = "failed"    // Failed permanently or exhausted retries.
	StatusQueued    JobStatus = "queued"    // Waiting for a worker.

	MaxLifetime = time.Minute * 3
	MaxRetries  = 5
)

// Job is a actively existing task with a state of completion
type Job struct {
	Lifetime time.Duration
	Status   JobStatus
	Retries  int
	Id       string
}

// JobResponse is the the result of a Job being finished. It can be used as a HTTP
type JobResponse struct {
	Id         string        `json:"id"`
	Status     JobStatus     `json:"status"`
	Attempts   int           `json:"attempts"`
	StatusCode int           `json:"status_code"`
	DurationMs time.Duration `json:"duration_ms"`
}

// NewJob constructs a job and returns it.
// Lifetime and retries will be clamped so they don't exceed certain limits.
func NewJob(lifetime time.Duration, retries int) *Job {
	lifetime = min(max(lifetime, 0), MaxLifetime)
	retries = min(max(retries, 0), MaxRetries)

	return &Job{
		Status:   StatusQueued,
		Lifetime: lifetime,
		Retries:  retries,
		Id:       rand.Text(),
	}
}
