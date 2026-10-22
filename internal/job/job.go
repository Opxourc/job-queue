package jobs

type JobStatus string

const (
	StatusRunning JobStatus = "running"
	StatusFailed  JobStatus = "failed"
	StatusDone    JobStatus = "done"

	maxLifetime = 180
	maxRetries  = 5
)

// Job is a actively existing task with a state of completion
type Job[T any] struct {
	Status   JobStatus
	Lifetime int
	Retries  int
	Task     T
}

// NewJob constructs a job and returns it.
// Certain values will be clamped to not exceed certain limits.
func NewJob[T any](lifetime, retries int, task T) *Job[T] {
	lifetime = min(max(lifetime, 0), maxLifetime)
	retries = min(max(retries, 0), maxRetries)

	return &Job[T]{
		Status:   StatusRunning,
		Lifetime: lifetime,
		Retries:  retries,
		Task:     task,
	}
}
