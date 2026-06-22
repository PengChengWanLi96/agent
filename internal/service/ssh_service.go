package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
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
	client, err := ssh.NewClient(opts)
	if err != nil {
		return nil, err
	}

	session := &SSHSession{
		ID:        generateSessionID(),
		Host:      opts.Host,
		User:      opts.User,
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

func (s *SSHService) ListDir(sessionID, path string) ([]ssh.FileInfo, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	return session.client.ListDir(path)
}

func (s *SSHService) Download(sessionID, path string) (io.ReadCloser, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	return session.client.Download(path)
}

func (s *SSHService) Upload(sessionID, path string, reader io.Reader) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}
	return session.client.Upload(path, reader)
}

func (s *SSHService) Remove(sessionID, path string) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}
	return session.client.Remove(path)
}

func (s *SSHService) Mkdir(sessionID, path string) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}
	return session.client.Mkdir(path)
}

func (s *SSHService) Rename(sessionID, oldPath, newPath string) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}
	return session.client.Rename(oldPath, newPath)
}

func (s *SSHService) Exec(sessionID, command string) (string, int, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return "", 0, err
	}
	return session.client.Exec(command)
}
