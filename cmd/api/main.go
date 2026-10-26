package main

import (
	"log"
	"net/http"
	"strconv"

	_ "modernc.org/sqlite"
	"github.com/Opxourc/job-queue/internal/config"
	"github.com/Opxourc/job-queue/internal/handler"
	"github.com/Opxourc/job-queue/internal/job"
	"github.com/Opxourc/job-queue/internal/service"
	"github.com/Opxourc/job-queue/internal/storage"
	"github.com/gin-contrib/timeout"
	"github.com/gin-gonic/gin"
)

func main() {
	db, err := storage.ConnectToDatabase(config.DbDriver, config.DbFileName)
	if err != nil {
		log.Fatal(err)
	}

	var repo job.Repository = storage.NewSqliteJobRepository(db)
	var serv job.Service = service.NewJobService(repo)
	var hdlr job.Handler = handler.NewJobHandler(serv)

	router := gin.New()
	router.Use(gin.Logger())
	router.Use(timeout.New(
		timeout.WithTimeout(config.ApiTimeout),
		timeout.WithResponse(func(c *gin.Context) {
			c.String(http.StatusGatewayTimeout, "request timed out")
		}),
	))

	api := router.Group("/api")
	{
		jobs := api.Group("/jobs")
		{
			jobs.GET("/", hdlr.List)
			jobs.GET("/:id", hdlr.GetByID)
			jobs.POST("/", hdlr.Create)
			jobs.POST("/:id/cancel", hdlr.Cancel)
		}
	}

	if err := router.Run(":" + strconv.Itoa(config.AddressPort)); err != nil {
		log.Fatal(err)
	}
}
