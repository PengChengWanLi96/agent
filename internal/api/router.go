package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"agent/internal/model"
	"agent/internal/service"
	"agent/internal/version"
	"agent/web"
)

func NewRouter(dockerSvc *service.DockerService, metricsSvc *service.MetricsService) *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, model.Response{Code: 0, Message: "ok"})
	})

	apiGroup := r.Group("/api/v1")
	{
		apiGroup.GET("/version", func(c *gin.Context) {
			c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success", Data: version.GetInfo()})
		})
	}

	dockerGroup := r.Group("/api/v1/docker")
	{
		h := NewDockerHandler(dockerSvc)
		dockerGroup.GET("/containers", h.ListContainers)
		dockerGroup.GET("/containers/:id", h.InspectContainer)
		dockerGroup.POST("/containers/:id/start", h.StartContainer)
		dockerGroup.POST("/containers/:id/stop", h.StopContainer)
		dockerGroup.DELETE("/containers/:id", h.RemoveContainer)
		dockerGroup.GET("/containers/:id/logs", h.ContainerLogs)
	}

	metricsGroup := r.Group("/api/v1/metrics")
	{
		h := NewMetricsHandler(metricsSvc)
		metricsGroup.GET("/collect", h.Collect)
		metricsGroup.GET("/prometheus", h.CollectRaw)
	}

	indexHTML, err := web.Content.ReadFile("index.html")
	if err != nil {
		panic(err)
	}
	r.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexHTML)
	})

	return r
}
