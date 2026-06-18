package api

import (
	"io"
	"net/http"
	"strconv"

	"agent/internal/model"
	"agent/internal/service"
	"github.com/gin-gonic/gin"
)

type DockerHandler struct {
	svc *service.DockerService
}

func NewDockerHandler(svc *service.DockerService) *DockerHandler {
	return &DockerHandler{svc: svc}
}

func (h *DockerHandler) ListContainers(c *gin.Context) {
	all, _ := strconv.ParseBool(c.Query("all"))

	filters := make(map[string][]string)
	if status := c.Query("status"); status != "" {
		filters["status"] = []string{status}
	}
	if label := c.Query("label"); label != "" {
		filters["label"] = []string{label}
	}
	if name := c.Query("name"); name != "" {
		filters["name"] = []string{name}
	}

	containers, err := h.svc.ListContainers(c.Request.Context(), all, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success", Data: containers})
}

func (h *DockerHandler) InspectContainer(c *gin.Context) {
	id := c.Param("id")
	ctn, err := h.svc.InspectContainer(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success", Data: ctn})
}

func (h *DockerHandler) CreateContainer(c *gin.Context) {
	var req model.ContainerCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Code: 400, Message: err.Error()})
		return
	}

	resp, err := h.svc.CreateContainer(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success", Data: resp})
}

func (h *DockerHandler) StartContainer(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.StartContainer(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "container started"})
}

func (h *DockerHandler) StopContainer(c *gin.Context) {
	id := c.Param("id")
	timeout := 10
	if t := c.Query("timeout"); t != "" {
		if v, err := strconv.Atoi(t); err == nil {
			timeout = v
		}
	}
	if err := h.svc.StopContainer(c.Request.Context(), id, timeout); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "container stopped"})
}

func (h *DockerHandler) RestartContainer(c *gin.Context) {
	id := c.Param("id")
	timeout := 10
	if t := c.Query("timeout"); t != "" {
		if v, err := strconv.Atoi(t); err == nil {
			timeout = v
		}
	}
	if err := h.svc.RestartContainer(c.Request.Context(), id, timeout); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "container restarted"})
}

func (h *DockerHandler) KillContainer(c *gin.Context) {
	id := c.Param("id")
	signal := c.DefaultQuery("signal", "SIGKILL")
	if err := h.svc.KillContainer(c.Request.Context(), id, signal); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "container killed"})
}

func (h *DockerHandler) RemoveContainer(c *gin.Context) {
	id := c.Param("id")
	force, _ := strconv.ParseBool(c.Query("force"))
	if err := h.svc.RemoveContainer(c.Request.Context(), id, force); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "container removed"})
}

func (h *DockerHandler) PauseContainer(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.PauseContainer(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "container paused"})
}

func (h *DockerHandler) UnpauseContainer(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.UnpauseContainer(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "container unpaused"})
}

func (h *DockerHandler) RenameContainer(c *gin.Context) {
	id := c.Param("id")
	var req model.ContainerRenameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Code: 400, Message: err.Error()})
		return
	}
	if err := h.svc.RenameContainer(c.Request.Context(), id, req.NewName); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "container renamed"})
}

func (h *DockerHandler) ContainerLogs(c *gin.Context) {
	id := c.Param("id")
	tail := c.DefaultQuery("tail", "100")
	timestamps, _ := strconv.ParseBool(c.Query("timestamps"))
	follow, _ := strconv.ParseBool(c.Query("follow"))
	since := c.Query("since")
	until := c.Query("until")

	logs, err := h.svc.ContainerLogs(c.Request.Context(), id, tail, timestamps, follow, since, until)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success", Data: logs})
}

func (h *DockerHandler) PruneContainers(c *gin.Context) {
	report, err := h.svc.PruneContainers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success", Data: report})
}

func (h *DockerHandler) ExecContainer(c *gin.Context) {
	id := c.Param("id")
	var req model.ContainerExecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Code: 400, Message: err.Error()})
		return
	}

	resp, err := h.svc.ExecContainer(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success", Data: resp})
}

// ==================== Image Handlers ====================

type ImageHandler struct {
	svc *service.DockerService
}

func NewImageHandler(svc *service.DockerService) *ImageHandler {
	return &ImageHandler{svc: svc}
}

func (h *ImageHandler) ListImages(c *gin.Context) {
	all, _ := strconv.ParseBool(c.Query("all"))

	filters := make(map[string][]string)
	if dangling := c.Query("dangling"); dangling != "" {
		filters["dangling"] = []string{dangling}
	}
	if label := c.Query("label"); label != "" {
		filters["label"] = []string{label}
	}

	images, err := h.svc.ListImages(c.Request.Context(), all, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success", Data: images})
}

func (h *ImageHandler) InspectImage(c *gin.Context) {
	id := c.Param("id")
	img, err := h.svc.InspectImage(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success", Data: img})
}

func (h *ImageHandler) PullImage(c *gin.Context) {
	var req model.ImagePullRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Code: 400, Message: err.Error()})
		return
	}

	reader, err := h.svc.PullImage(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	defer reader.Close()

	c.Header("Content-Type", "application/json")
	c.Status(http.StatusOK)
	io.Copy(c.Writer, reader)
}

func (h *ImageHandler) RemoveImage(c *gin.Context) {
	id := c.Param("id")
	force, _ := strconv.ParseBool(c.Query("force"))

	deleted, err := h.svc.RemoveImage(c.Request.Context(), id, force)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success", Data: map[string]interface{}{
		"deleted": deleted,
	}})
}

func (h *ImageHandler) PruneImages(c *gin.Context) {
	report, err := h.svc.PruneImages(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success", Data: report})
}

// ==================== Network Handlers ====================

type NetworkHandler struct {
	svc *service.DockerService
}

func NewNetworkHandler(svc *service.DockerService) *NetworkHandler {
	return &NetworkHandler{svc: svc}
}

func (h *NetworkHandler) ListNetworks(c *gin.Context) {
	filters := make(map[string][]string)
	if driver := c.Query("driver"); driver != "" {
		filters["driver"] = []string{driver}
	}
	if name := c.Query("name"); name != "" {
		filters["name"] = []string{name}
	}

	networks, err := h.svc.ListNetworks(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success", Data: networks})
}

func (h *NetworkHandler) CreateNetwork(c *gin.Context) {
	var req model.NetworkCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Code: 400, Message: err.Error()})
		return
	}
	resp, err := h.svc.CreateNetwork(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success", Data: resp})
}

func (h *NetworkHandler) RemoveNetwork(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.RemoveNetwork(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "network removed"})
}

func (h *NetworkHandler) ConnectNetwork(c *gin.Context) {
	id := c.Param("id")
	var req model.NetworkConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Code: 400, Message: err.Error()})
		return
	}
	if err := h.svc.ConnectNetwork(c.Request.Context(), id, &req); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "container connected to network"})
}

func (h *NetworkHandler) DisconnectNetwork(c *gin.Context) {
	id := c.Param("id")
	var req model.NetworkDisconnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Code: 400, Message: err.Error()})
		return
	}
	if err := h.svc.DisconnectNetwork(c.Request.Context(), id, &req); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "container disconnected from network"})
}

func (h *NetworkHandler) PruneNetworks(c *gin.Context) {
	report, err := h.svc.PruneNetworks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success", Data: report})
}

// ==================== Volume Handlers ====================

type VolumeHandler struct {
	svc *service.DockerService
}

func NewVolumeHandler(svc *service.DockerService) *VolumeHandler {
	return &VolumeHandler{svc: svc}
}

func (h *VolumeHandler) ListVolumes(c *gin.Context) {
	volumes, err := h.svc.ListVolumes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success", Data: volumes})
}

func (h *VolumeHandler) CreateVolume(c *gin.Context) {
	var req model.VolumeCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Code: 400, Message: err.Error()})
		return
	}
	vol, err := h.svc.CreateVolume(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success", Data: vol})
}

func (h *VolumeHandler) RemoveVolume(c *gin.Context) {
	name := c.Param("name")
	force, _ := strconv.ParseBool(c.Query("force"))
	if err := h.svc.RemoveVolume(c.Request.Context(), name, force); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "volume removed"})
}

func (h *VolumeHandler) PruneVolumes(c *gin.Context) {
	report, err := h.svc.PruneVolumes(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success", Data: report})
}
