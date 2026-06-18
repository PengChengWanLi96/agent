package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"agent/internal/model"
	"agent/internal/service"
	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	svc *service.UploadService
}

func NewUploadHandler(svc *service.UploadService) *UploadHandler {
	return &UploadHandler{svc: svc}
}

func (h *UploadHandler) UploadFiles(c *gin.Context) {
	destDir := c.PostForm("dest_dir")
	if err := validateUploadDestDir(destDir); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Code: 400, Message: err.Error()})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Code: 400, Message: err.Error()})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, model.Response{Code: 400, Message: "no files uploaded"})
		return
	}

	results := make([]map[string]interface{}, 0, len(files))
	for _, fileHeader := range files {
		src, err := fileHeader.Open()
		if err != nil {
			results = append(results, map[string]interface{}{
				"name":    fileHeader.Filename,
				"success": false,
				"error":   err.Error(),
			})
			continue
		}
		func() {
			defer src.Close()
			destPath := filepath.Join(destDir, filepath.Base(fileHeader.Filename))
			uploaded, err := h.svc.SaveUploadedFile(src, destPath, fileHeader.Size)
			if err != nil {
				results = append(results, map[string]interface{}{
					"name":    fileHeader.Filename,
					"success": false,
					"error":   err.Error(),
				})
				return
			}
			results = append(results, map[string]interface{}{
				"name":    uploaded.Name,
				"size":    uploaded.Size,
				"success": true,
			})
		}()
	}

	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success", Data: results})
}

func (h *UploadHandler) ListUploadDir(c *gin.Context) {
	destDir := c.Query("dest_dir")
	if err := validateUploadDestDir(destDir); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Code: 400, Message: err.Error()})
		return
	}

	entries, err := os.ReadDir(destDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: 500, Message: err.Error()})
		return
	}

	items := make([]map[string]interface{}, 0, len(entries))
	for _, entry := range entries {
		info, _ := entry.Info()
		size := int64(0)
		if info != nil {
			size = info.Size()
		}
		items = append(items, map[string]interface{}{
			"name":  entry.Name(),
			"is_dir": entry.IsDir(),
			"size":  size,
		})
	}

	c.JSON(http.StatusOK, model.Response{Code: 0, Message: "success", Data: map[string]interface{}{
		"dest_dir": destDir,
		"items":    items,
	}})
}

func validateUploadDestDir(destDir string) error {
	if destDir == "" {
		return fmt.Errorf("dest_dir is required")
	}
	cleaned := filepath.Clean(destDir)
	if cleaned == "/" || cleaned == "." {
		return nil
	}
	return nil
}
