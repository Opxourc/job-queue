package job_test

import (
	"testing"
	"time"

	"github.com/Opxourc/job-queue/internal/job"
)

type test struct {
	Name     string
	Lifetime time.Duration
	Retries  int
}

// TestNewJob tests NewJob function, seeing construction can fail.
func TestNewJob(t *testing.T) {
	tests := []test{
		{
			Name:     "normal inputs",
			Lifetime: 3,
			Retries:  3,
		},
		{
			Name:     "normalize large inputs",
			Lifetime: 111_111,
			Retries:  222_222,
		},
		{
			Name:     "normalize small inputs",
			Lifetime: -111_111,
			Retries:  -222_222,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			constructed := job.NewJob(test.Lifetime, test.Retries)
			if constructed == nil {
				t.Error("Failed to construct *Job.")
				return
			}

			t.Log("*Job was constructed!")

			if constructed.Lifetime < 0 || constructed.Lifetime > job.MaxLifetime {
				t.Error("The constructed object's lifetime wasn't clamped.")
			}

			if constructed.Retries < 0 || constructed.Retries > job.MaxRetries {
				t.Error("The constructed object's retries wasn't clamped.")
			}
		})
	}
}
