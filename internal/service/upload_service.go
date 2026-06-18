package service

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type UploadService struct{}

func NewUploadService() *UploadService {
	return &UploadService{}
}

type UploadedFile struct {
	Name     string `json:"name"`
	Original string `json:"original"`
	Size     int64  `json:"size"`
}

func (s *UploadService) SaveFiles(destDir string, files []UploadedFile) ([]UploadedFile, error) {
	if destDir == "" {
		destDir = "."
	}

	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", destDir, err)
	}

	saved := make([]UploadedFile, 0, len(files))
	for _, f := range files {
		if f.Name == "" {
			continue
		}
		saved = append(saved, f)
	}
	return saved, nil
}

func (s *UploadService) SaveUploadedFile(src io.Reader, destPath string, size int64) (*UploadedFile, error) {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	written, err := io.Copy(out, src)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return &UploadedFile{
		Name:     filepath.Base(destPath),
		Original: filepath.Base(destPath),
		Size:     written,
	}, nil
}
