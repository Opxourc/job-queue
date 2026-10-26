package job_test

import (
	"testing"
	"time"

	"github.com/Opxourc/job-queue/internal/job"
)

func TestNewJob(t *testing.T) {
	tests := []struct {
		Name string
		Req  job.Request
	}{
		{
			Name: "normal inputs",
			Req: job.Request{
				Lifetime: time.Second * 3,
				Retries:  3,
			},
		},
		{
			Name: "normalize large inputs",
			Req: job.Request{
				Lifetime: time.Second * 111_111,
				Retries:  222_222,
			},
		},
		{
			Name: "normalize small inputs",
			Req: job.Request{
				Lifetime: time.Second * -111_111,
				Retries:  -222_222,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			constructed := job.NewJob(&test.Req)
			if constructed == nil {
				t.Error("failed to construct")
				return
			}

			if constructed.Lifetime < 0 || constructed.Lifetime > job.MaxLifetime {
				t.Error("lifetime wasn't clamped")
			}

			if constructed.Retries < 0 || constructed.Retries > job.MaxRetries {
				t.Error("retries wasn't clamped")
			}
		})
	}
}
