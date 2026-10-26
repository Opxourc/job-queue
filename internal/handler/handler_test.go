package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"uuid"

	"github.com/Opxourc/job-queue/internal/handler"
	"github.com/Opxourc/job-queue/internal/job"
	"github.com/gin-gonic/gin"
)

type testRepository struct {
	createdJob job.Job
	foundJob   job.Job
	jobs       []job.Job
	createErr  error
	getErr     error
	listErr    error
	createReq  job.Request
	createCall int
	getID      uuid.UUID
	getCall    int
	listCall   int
}

func (r *testRepository) Create(_ context.Context, req job.Request) (job.Job, error) {
	r.createCall++
	r.createReq = req
	return r.createdJob, r.createErr
}

func (r *testRepository) GetByID(_ context.Context, id uuid.UUID) (job.Job, error) {
	r.getCall++
	r.getID = id
	return r.foundJob, r.getErr
}

func (r *testRepository) List(context.Context) ([]job.Job, error) {
	r.listCall++
	return r.jobs, r.listErr
}

func (*testRepository) Update(context.Context, job.Job) error {
	return nil
}

func (*testRepository) Delete(context.Context, uuid.UUID) error {
	return nil
}

type runCall struct {
	ctx context.Context
	job job.Job
}

type testService struct {
	repo      *testRepository
	runCalls  chan runCall
	cancelIDs []uuid.UUID
}

func (s *testService) Run(ctx context.Context, current job.Job) {
	s.runCalls <- runCall{ctx: ctx, job: current}
}

func (s *testService) Cancel(id uuid.UUID) {
	s.cancelIDs = append(s.cancelIDs, id)
}

func (s *testService) Repo() job.Repository {
	return s.repo
}

func newTestContext(method, path, body string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(method, path, strings.NewReader(body))
	return ctx, recorder
}

func TestCreate(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		name       string
		body       string
		createErr  error
		wantStatus int
		wantCalls  int
	}{
		{name: "invalid JSON", body: "{", wantStatus: http.StatusBadRequest},
		{name: "repository error", body: `{"retries":2,"lifetime":3000000000}`, createErr: errors.New("create failed"), wantStatus: http.StatusInternalServerError, wantCalls: 1},
		{name: "created", body: `{"retries":2,"lifetime":3000000000}`, wantStatus: http.StatusAccepted, wantCalls: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testRepository{createdJob: job.Job{Id: id}, createErr: tt.createErr}
			svc := &testService{repo: repo, runCalls: make(chan runCall, 1)}
			ctx, recorder := newTestContext(http.MethodPost, "/jobs", tt.body)

			handler.NewJobHandler(svc).Create(ctx)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			if repo.createCall != tt.wantCalls {
				t.Fatalf("Create calls = %d, want %d", repo.createCall, tt.wantCalls)
			}
			if tt.wantCalls == 0 {
				return
			}
			if repo.createReq != (job.Request{Retries: 2, Lifetime: 3 * time.Second}) {
				t.Errorf("Create request = %+v, want retries 2 and lifetime 3s", repo.createReq)
			}
			if tt.createErr != nil {
				select {
				case call := <-svc.runCalls:
					t.Errorf("Run called after create failure with job %+v", call.job)
				default:
				}
				return
			}

			var response uuid.UUID
			if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if response != id {
				t.Errorf("response ID = %s, want %s", response, id)
			}
			select {
			case call := <-svc.runCalls:
				if call.job.Id != id {
					t.Errorf("Run job ID = %s, want %s", call.job.Id, id)
				}
				if call.ctx == nil {
					t.Error("Run context is nil")
				}
			case <-time.After(time.Second):
				t.Fatal("Run was not called")
			}
		})
	}
}

func TestGetByID(t *testing.T) {
	id := uuid.New()
	expected := job.Job{Id: id, Status: job.StatusQueued}
	tests := []struct {
		name       string
		param      string
		setParam   bool
		getErr     error
		wantStatus int
		wantCalls  int
	}{
		{name: "missing ID", wantStatus: http.StatusBadRequest},
		{name: "invalid ID", param: "not-a-uuid", setParam: true, wantStatus: http.StatusBadRequest},
		{name: "job not found", param: id.String(), setParam: true, getErr: errors.New("not found"), wantStatus: http.StatusNotFound, wantCalls: 1},
		{name: "found", param: id.String(), setParam: true, wantStatus: http.StatusOK, wantCalls: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testRepository{foundJob: expected, getErr: tt.getErr}
			svc := &testService{repo: repo}
			ctx, recorder := newTestContext(http.MethodGet, "/jobs/"+tt.param, "")
			if tt.setParam {
				ctx.Params = gin.Params{{Key: "id", Value: tt.param}}
			}

			handler.NewJobHandler(svc).GetByID(ctx)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			if repo.getCall != tt.wantCalls {
				t.Fatalf("GetByID calls = %d, want %d", repo.getCall, tt.wantCalls)
			}
			if tt.wantCalls == 0 {
				return
			}
			if repo.getID != id {
				t.Errorf("GetByID ID = %s, want %s", repo.getID, id)
			}
			if tt.getErr == nil {
				var response job.Job
				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				if response != expected {
					t.Errorf("response = %+v, want %+v", response, expected)
				}
			}
		})
	}
}

func TestList(t *testing.T) {
	expected := []job.Job{{Id: uuid.New(), Status: job.StatusQueued}}
	tests := []struct {
		name       string
		listErr    error
		wantStatus int
	}{
		{name: "repository error", listErr: errors.New("list failed"), wantStatus: http.StatusInternalServerError},
		{name: "listed", wantStatus: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testRepository{jobs: expected, listErr: tt.listErr}
			ctx, recorder := newTestContext(http.MethodGet, "/jobs", "")

			handler.NewJobHandler(&testService{repo: repo}).List(ctx)

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			if repo.listCall != 1 {
				t.Fatalf("List calls = %d, want 1", repo.listCall)
			}
			if tt.listErr == nil {
				var response []job.Job
				if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				if len(response) != len(expected) || response[0] != expected[0] {
					t.Errorf("response = %+v, want %+v", response, expected)
				}
			}
		})
	}
}

func TestCancel(t *testing.T) {
	id := uuid.New()
	tests := []struct {
		name       string
		param      string
		setParam   bool
		getErr     error
		wantStatus int
		wantCalls  int
		wantCancel bool
	}{
		{name: "missing ID", wantStatus: http.StatusBadRequest},
		{name: "invalid ID", param: "not-a-uuid", setParam: true, wantStatus: http.StatusBadRequest},
		{name: "job not found", param: id.String(), setParam: true, getErr: errors.New("not found"), wantStatus: http.StatusNotFound, wantCalls: 1},
		{name: "canceled", param: id.String(), setParam: true, wantStatus: http.StatusNoContent, wantCalls: 1, wantCancel: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &testRepository{getErr: tt.getErr}
			svc := &testService{repo: repo}
			ctx, recorder := newTestContext(http.MethodPost, "/jobs/"+tt.param+"/cancel", "")
			if tt.setParam {
				ctx.Params = gin.Params{{Key: "id", Value: tt.param}}
			}

			handler.NewJobHandler(svc).Cancel(ctx)
			ctx.Writer.WriteHeaderNow()

			if recorder.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			if repo.getCall != tt.wantCalls {
				t.Fatalf("GetByID calls = %d, want %d", repo.getCall, tt.wantCalls)
			}
			if tt.wantCancel {
				if len(svc.cancelIDs) != 1 || svc.cancelIDs[0] != id {
					t.Errorf("Cancel IDs = %v, want [%s]", svc.cancelIDs, id)
				}
			} else if len(svc.cancelIDs) != 0 {
				t.Errorf("Cancel called with IDs %v, want no calls", svc.cancelIDs)
			}
			if tt.wantCalls > 0 && repo.getID != id {
				t.Errorf("GetByID ID = %s, want %s", repo.getID, id)
			}
		})
	}
}
