package api

import (
	"net/http"
	"strconv"
	"github.com/gin-gonic/gin"
	"agent/internal/model"
	"agent/internal/service"
)

type DockerHandler struct {
	svc *service.DockerService
}

func NewDockerHandler(svc *service.DockerService) *DockerHandler {
	return &DockerHandler{svc: svc}
}

func (h *DockerHandler) ListContainers(c *gin.Context) {
	all, _ := strconv.ParseBool(c.Query("all"))
	containers, err := h.svc.ListContainers(c.Request.Context(), all)
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

func (h *DockerHandler) RemoveContainer(c *gin.Context) {
	id := c.Param("id")
	force, _ := strconv.ParseBool(c.Query("force"))
	if err := h.svc.RemoveContainer(c.Request.Context(), id, force); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "container removed"})
}

func (h *DockerHandler) ContainerLogs(c *gin.Context) {
	id := c.Param("id")
	tail := c.DefaultQuery("tail", "100")
	logs, err := h.svc.ContainerLogs(c.Request.Context(), id, tail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success", Data: logs})
}
