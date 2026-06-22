package api

import (
	"net/http"

	"agent/internal/model"
	"agent/internal/service"
	"agent/internal/version"
	"agent/web"
	"github.com/gin-gonic/gin"
)

func NewRouter(dockerSvc *service.DockerService, metricsSvc *service.MetricsService, sshSvc *service.SSHService) *gin.Engine {
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

	uploadGroup := r.Group("/api/v1/upload")
	{
		uh := NewUploadHandler(service.NewUploadService())
		uploadGroup.POST("/files", uh.UploadFiles)
		uploadGroup.GET("/files", uh.ListUploadDir)
	}

	dockerGroup := r.Group("/api/v1/docker")
	{
		dh := NewDockerHandler(dockerSvc)
		dockerGroup.GET("/containers", dh.ListContainers)
		dockerGroup.POST("/containers", dh.CreateContainer)
		dockerGroup.GET("/containers/:id", dh.InspectContainer)
		dockerGroup.POST("/containers/:id/start", dh.StartContainer)
		dockerGroup.POST("/containers/:id/stop", dh.StopContainer)
		dockerGroup.POST("/containers/:id/restart", dh.RestartContainer)
		dockerGroup.POST("/containers/:id/kill", dh.KillContainer)
		dockerGroup.DELETE("/containers/:id", dh.RemoveContainer)
		dockerGroup.POST("/containers/:id/pause", dh.PauseContainer)
		dockerGroup.POST("/containers/:id/unpause", dh.UnpauseContainer)
		dockerGroup.POST("/containers/:id/rename", dh.RenameContainer)
		dockerGroup.POST("/containers/:id/exec", dh.ExecContainer)
		dockerGroup.GET("/containers/:id/logs", dh.ContainerLogs)
		dockerGroup.POST("/containers/prune", dh.PruneContainers)

		ih := NewImageHandler(dockerSvc)
		dockerGroup.GET("/images", ih.ListImages)
		dockerGroup.GET("/images/:id", ih.InspectImage)
		dockerGroup.POST("/images/pull", ih.PullImage)
		dockerGroup.DELETE("/images/:id", ih.RemoveImage)
		dockerGroup.POST("/images/prune", ih.PruneImages)

		nh := NewNetworkHandler(dockerSvc)
		dockerGroup.GET("/networks", nh.ListNetworks)
		dockerGroup.POST("/networks", nh.CreateNetwork)
		dockerGroup.DELETE("/networks/:id", nh.RemoveNetwork)
		dockerGroup.POST("/networks/:id/connect", nh.ConnectNetwork)
		dockerGroup.POST("/networks/:id/disconnect", nh.DisconnectNetwork)
		dockerGroup.POST("/networks/prune", nh.PruneNetworks)

		vh := NewVolumeHandler(dockerSvc)
		dockerGroup.GET("/volumes", vh.ListVolumes)
		dockerGroup.POST("/volumes", vh.CreateVolume)
		dockerGroup.DELETE("/volumes/:name", vh.RemoveVolume)
		dockerGroup.POST("/volumes/prune", vh.PruneVolumes)
	}

	metricsGroup := r.Group("/api/v1/metrics")
	{
		h := NewMetricsHandler(metricsSvc)
		metricsGroup.GET("/collect", h.Collect)
		metricsGroup.GET("/prometheus", h.CollectRaw)
	}

	sshGroup := r.Group("/api/v1/ssh")
	{
		h := NewSSHHandler(sshSvc)
		sshGroup.GET("/sessions", h.ListSessions)
		sshGroup.POST("/connect", h.Connect)
		sshGroup.DELETE("/sessions/:id", h.CloseSession)
		sshGroup.GET("/sessions/:id/files", h.ListFiles)
		sshGroup.GET("/sessions/:id/download", h.DownloadFile)
		sshGroup.POST("/sessions/:id/upload", h.UploadFile)
		sshGroup.DELETE("/sessions/:id/files", h.RemoveFile)
		sshGroup.POST("/sessions/:id/mkdir", h.Mkdir)
		sshGroup.POST("/sessions/:id/rename", h.RenameFile)
		sshGroup.POST("/sessions/:id/exec", h.ExecCommand)
		sshGroup.GET("/sessions/:id/exec", h.ExecCommandGet)
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
