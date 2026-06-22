package api

import (
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"agent/internal/client/ssh"
	"agent/internal/model"
	"agent/internal/service"
	"github.com/gin-gonic/gin"
)

type SSHHandler struct {
	svc *service.SSHService
}

func NewSSHHandler(svc *service.SSHService) *SSHHandler {
	return &SSHHandler{svc: svc}
}

func (h *SSHHandler) Connect(c *gin.Context) {
	var req model.SSHConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Code: 400, Message: err.Error()})
		return
	}

	opts := ssh.ConnectOptions{
		Host:       req.Host,
		Port:       req.Port,
		User:       req.User,
		Password:   req.Password,
		PrivateKey: req.PrivateKey,
	}

	session, err := h.svc.Connect(opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success", Data: model.SSHSessionResponse{
		ID:        session.ID,
		Host:      session.Host,
		User:      session.User,
		CreatedAt: session.CreatedAt.Unix(),
	}})
}

func (h *SSHHandler) ListSessions(c *gin.Context) {
	sessions := h.svc.ListSessions()
	resp := make([]model.SSHSessionResponse, 0, len(sessions))
	for _, s := range sessions {
		resp = append(resp, model.SSHSessionResponse{
			ID:        s.ID,
			Host:      s.Host,
			User:      s.User,
			CreatedAt: s.CreatedAt.Unix(),
		})
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success", Data: resp})
}

func (h *SSHHandler) CloseSession(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.CloseSession(id); err != nil {
		c.JSON(http.StatusNotFound, model.Response{Code: 404, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success"})
}

func (h *SSHHandler) ListFiles(c *gin.Context) {
	id := c.Param("id")
	path := c.Query("path")
	if path == "" {
		path = "/"
	}

	items, err := h.svc.ListDir(id, path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}

	resp := make([]model.SSHFileInfo, 0, len(items))
	for _, item := range items {
		resp = append(resp, model.SSHFileInfo{
			Name:    item.Name,
			Path:    item.Path,
			Size:    item.Size,
			IsDir:   item.IsDir,
			Mode:    item.Mode,
			ModTime: item.ModTime,
		})
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success", Data: resp})
}

func (h *SSHHandler) DownloadFile(c *gin.Context) {
	id := c.Param("id")
	filePath := c.Query("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, model.Response{Code: 400, Message: "path is required"})
		return
	}

	reader, err := h.svc.Download(id, filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	defer reader.Close()

	filename := path.Base(filePath)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Type", "application/octet-stream")
	c.Status(http.StatusOK)
	if _, err := io.Copy(c.Writer, reader); err != nil {
		_ = c.Error(err)
	}
}

func (h *SSHHandler) UploadFile(c *gin.Context) {
	id := c.Param("id")
	filePath := c.PostForm("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, model.Response{Code: 400, Message: "path is required"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Code: 400, Message: err.Error()})
		return
	}

	src, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	defer src.Close()

	if err := h.svc.Upload(id, filePath, src); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success", Data: map[string]interface{}{
		"path": filePath,
		"size": fileHeader.Size,
	}})
}

func (h *SSHHandler) RemoveFile(c *gin.Context) {
	id := c.Param("id")
	filePath := c.Query("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, model.Response{Code: 400, Message: "path is required"})
		return
	}

	if err := h.svc.Remove(id, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success"})
}

func (h *SSHHandler) Mkdir(c *gin.Context) {
	id := c.Param("id")
	filePath := c.Query("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, model.Response{Code: 400, Message: "path is required"})
		return
	}

	if err := h.svc.Mkdir(id, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success"})
}

func (h *SSHHandler) RenameFile(c *gin.Context) {
	id := c.Param("id")
	var req model.SSHRenameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Code: 400, Message: err.Error()})
		return
	}

	if err := h.svc.Rename(id, req.OldPath, req.NewPath); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success"})
}

func (h *SSHHandler) ExecCommand(c *gin.Context) {
	id := c.Param("id")
	var req model.SSHExecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Code: 400, Message: err.Error()})
		return
	}

	output, exitCode, err := h.svc.Exec(id, strings.TrimSpace(req.Command))
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success", Data: model.SSHExecResponse{
		Output:   output,
		ExitCode: exitCode,
	}})
}

func (h *SSHHandler) ExecCommandGet(c *gin.Context) {
	id := c.Param("id")
	command := strings.TrimSpace(c.Query("command"))
	if command == "" {
		c.JSON(http.StatusBadRequest, model.Response{Code: 400, Message: "command is required"})
		return
	}

	output, exitCode, err := h.svc.Exec(id, command)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success", Data: model.SSHExecResponse{
		Output:   output,
		ExitCode: exitCode,
	}})
}
