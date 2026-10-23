package job_test

import (
	"context"
	"testing"
	"time"

	"github.com/Opxourc/job-queue/internal/job"
)

type workerTest struct {
	Channel chan (*job.JobResponse)
	Name    string
	Cancel  bool
}

// TestRunJob tests RunJob function, seeing if running a job can fail or have unintended behavior.
func TestRunJob(t *testing.T) {
	tests := []workerTest{
		{
			Channel: make(chan (*job.JobResponse)),
			Name:    "natural run off",
			Cancel:  false,
		},
		{
			Channel: make(chan (*job.JobResponse)),
			Name:    "force cancelation",
			Cancel:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()

			go job.RunJob(ctx, job.NewJob(time.Second*5, 3), test.Channel)
			if test.Cancel {
				cancel()
			}

			res := <-test.Channel
			if test.Cancel && res.Status != job.StatusCanceled {
				t.Errorf("Context was canceled but status didn't reflect it. Was: %s", res.Status)
				return
			}

			t.Logf("Job was successfully ran! Status: %s, Attempts: %d", res.Status, res.Attempts)
		})
	}
}
