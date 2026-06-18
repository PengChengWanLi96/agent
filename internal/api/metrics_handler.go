package api

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"agent/internal/model"
	"agent/internal/service"
)

type MetricsHandler struct {
	svc *service.MetricsService
}

func NewMetricsHandler(svc *service.MetricsService) *MetricsHandler {
	return &MetricsHandler{svc: svc}
}

func (h *MetricsHandler) Collect(c *gin.Context) {
	metrics, err := h.svc.Collect()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success", Data: metrics})
}

func (h *MetricsHandler) CollectRaw(c *gin.Context) {
	raw, err := h.svc.CollectRaw()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.String(http.StatusOK, raw)
}
