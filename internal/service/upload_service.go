package service

import (
	"fmt"
	"io"
	"log"
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
	absPath, err := filepath.Abs(destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
	}
	log.Printf("[upload] saving file to: %s (original dest: %s)", absPath, destPath)

	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	out, err := os.Create(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	written, err := io.Copy(out, src)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	log.Printf("[upload] saved file successfully: %s, size: %d", absPath, written)
	return &UploadedFile{
		Name:     filepath.Base(absPath),
		Original: filepath.Base(absPath),
		Size:     written,
	}, nil
}
