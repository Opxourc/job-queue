package handler

import (
	"context"
	"net/http"
	"uuid"

	"github.com/Opxourc/job-queue/internal/job"
	"github.com/gin-gonic/gin"
)

// JobHandler implements job.Handler for Jobs HTTP operations.
type JobHandler struct {
	srv job.Service
}

// NewJobHandler a JohHandler object.
func NewJobHandler(s job.Service) *JobHandler {
	return &JobHandler{
		srv: s,
	}
}

func (h *JobHandler) Create(c *gin.Context) {
	var req job.Request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	created, err := h.srv.Repo().Create(c.Request.Context(), req)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	go h.srv.Run(context.Background(), created)
	c.JSON(http.StatusAccepted, created.Id)
}

func (h *JobHandler) GetByID(c *gin.Context) {
	idParam, ok := c.Params.Get("id")
	if !ok {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(idParam)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	foundJob, err := h.srv.Repo().GetByID(c.Request.Context(), id)
	if err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, foundJob)
}

func (h *JobHandler) List(c *gin.Context) {
	jobs, err := h.srv.Repo().List(c.Request.Context())
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, jobs)
}

func (h *JobHandler) Cancel(c *gin.Context) {
	idParam, ok := c.Params.Get("id")
	if !ok {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(idParam)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if _, err := h.srv.Repo().GetByID(c.Request.Context(), id); err != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	h.srv.Cancel(id)
	c.Status(http.StatusNoContent)
}
