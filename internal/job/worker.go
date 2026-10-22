package job

import (
	"context"
	"math/rand/v2"
	"net/http"
	"time"
)

// RunJob runs the provided task and handles retries, durations, etc.
// When the job has finished, it will send the result of what happened back through a channel.
func RunJob(ctx context.Context, job *Job, res chan<- *JobResponse) {
	job.Status = StatusRunning
	attempts := 0

	var fullDuration time.Duration
	succeeded := false

outer:
	for attempts == job.Retries && !succeeded {
		// Simulate duration of a job
		duration := time.Duration(rand.IntN(10)+1) * time.Second

		select {
		case <-time.After(duration):
			time.Sleep(duration)
			fullDuration += duration

			// Simulate succeess or failure of job
			if ok := rand.IntN(2); ok == 0 {
				succeeded = false
				attempts++
			} else {
				succeeded = true
				job.Status = StatusSucceeded
			}

		case <-ctx.Done():
			job.Status = StatusCanceled
			break outer
		}
	}

	if attempts == job.Retries && !succeeded && ctx.Err() == nil {
		job.Status = StatusFailed
	}

	var statusCode int
	switch job.Status {
	case StatusCanceled:
		statusCode = http.StatusGone

	case StatusSucceeded:
		statusCode = http.StatusOK

	case StatusFailed:
		statusCode = http.StatusInternalServerError
	}

	res <- &JobResponse{
		Id:         job.Id,
		Status:     job.Status,
		Attempts:   attempts,
		StatusCode: statusCode,
		DurationMs: fullDuration,
	}
}
