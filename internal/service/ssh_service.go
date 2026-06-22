package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"agent/internal/client/ssh"
)

func generateSessionID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

type SSHSession struct {
	ID        string
	Host      string
	User      string
	Local     bool
	CreatedAt time.Time
	client    *ssh.Client
}

type SSHService struct {
	sessions map[string]*SSHSession
	mu       sync.RWMutex
}

func NewSSHService() *SSHService {
	return &SSHService{
		sessions: make(map[string]*SSHSession),
	}
}

func (s *SSHService) Connect(opts ssh.ConnectOptions) (*SSHSession, error) {
	var client *ssh.Client
	var err error
	if opts.Local {
		client, err = nil, nil
	} else {
		client, err = ssh.NewClient(opts)
		if err != nil {
			return nil, err
		}
	}

	session := &SSHSession{
		ID:        generateSessionID(),
		Host:      opts.Host,
		User:      opts.User,
		Local:     opts.Local,
		CreatedAt: time.Now(),
		client:    client,
	}

	s.mu.Lock()
	s.sessions[session.ID] = session
	s.mu.Unlock()

	return session, nil
}

func (s *SSHService) CloseSession(id string) error {
	s.mu.Lock()
	session, ok := s.sessions[id]
	delete(s.sessions, id)
	s.mu.Unlock()

	if !ok {
		return fmt.Errorf("session not found")
	}
	if session.Local {
		return nil
	}
	return session.client.Close()
}

func (s *SSHService) GetSession(id string) (*SSHSession, error) {
	s.mu.RLock()
	session, ok := s.sessions[id]
	s.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("session not found")
	}
	return session, nil
}

func (s *SSHService) ListSessions() []*SSHSession {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sessions := make([]*SSHSession, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

func (s *SSHService) ListDir(sessionID, p string) ([]ssh.FileInfo, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	if session.Local {
		return s.localListDir(p)
	}
	return session.client.ListDir(p)
}

func (s *SSHService) Download(sessionID, p string) (io.ReadCloser, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	if session.Local {
		return s.localDownload(p)
	}
	return session.client.Download(p)
}

func (s *SSHService) Upload(sessionID, p string, reader io.Reader) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}
	if session.Local {
		return s.localUpload(p, reader)
	}
	return session.client.Upload(p, reader)
}

func (s *SSHService) Remove(sessionID, p string) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}
	if session.Local {
		return s.localRemove(p)
	}
	return session.client.Remove(p)
}

func (s *SSHService) Mkdir(sessionID, p string) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}
	if session.Local {
		return s.localMkdir(p)
	}
	return session.client.Mkdir(p)
}

func (s *SSHService) Rename(sessionID, oldPath, newPath string) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}
	if session.Local {
		return s.localRename(oldPath, newPath)
	}
	return session.client.Rename(oldPath, newPath)
}

func (s *SSHService) Exec(sessionID, command string) (string, int, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return "", 0, err
	}
	if session.Local {
		return s.localExec(command)
	}
	return session.client.Exec(command)
}

// local helpers

func (s *SSHService) localListDir(p string) ([]ssh.FileInfo, error) {
	entries, err := os.ReadDir(p)
	if err != nil {
		return nil, err
	}
	items := make([]ssh.FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, _ := entry.Info()
		fi := ssh.FileInfo{
			Name:  entry.Name(),
			Path:  filepath.Join(p, entry.Name()),
			IsDir: entry.IsDir(),
		}
		if info != nil {
			fi.Size = info.Size()
			fi.ModTime = info.ModTime().Unix()
			fi.Mode = info.Mode().String()
		}
		items = append(items, fi)
	}
	return items, nil
}

func (s *SSHService) localDownload(p string) (io.ReadCloser, error) {
	return os.Open(p)
}

func (s *SSHService) localUpload(p string, reader io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	out, err := os.Create(p)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, reader)
	return err
}

func (s *SSHService) localRemove(p string) error {
	return os.RemoveAll(p)
}

func (s *SSHService) localMkdir(p string) error {
	return os.MkdirAll(p, 0o755)
}

func (s *SSHService) localRename(oldPath, newPath string) error {
	return os.Rename(oldPath, newPath)
}

func (s *SSHService) localExec(command string) (string, int, error) {
	cmd := exec.Command("sh", "-c", command)
	out, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return "", 0, err
		}
	}
	return string(out), exitCode, nil
}
