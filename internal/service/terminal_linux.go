//go:build linux

package service

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/creack/pty"
)

type localTerminal struct {
	pty    *os.File
	cmd    *exec.Cmd
	mu     sync.Mutex
	closed bool
}

func newLocalTerminal() (*localTerminal, error) {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/sh"
	}
	cmd := exec.Command(shell)
	cmd.Env = os.Environ()

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to start local pty: %w", err)
	}

	return &localTerminal{pty: ptmx, cmd: cmd}, nil
}

func (t *localTerminal) Read(p []byte) (int, error) {
	return t.pty.Read(p)
}

func (t *localTerminal) Write(p []byte) (int, error) {
	return t.pty.Write(p)
}

func (t *localTerminal) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return nil
	}
	t.closed = true
	t.pty.Close()
	if t.cmd.Process != nil {
		_ = t.cmd.Process.Kill()
	}
	_ = t.cmd.Wait()
	return nil
}

func (t *localTerminal) Resize(rows, cols int) error {
	return pty.Setsize(t.pty, &pty.Winsize{Rows: uint16(rows), Cols: uint16(cols)})
}

var _ terminalSession = (*localTerminal)(nil)
var _ io.ReadWriteCloser = (*localTerminal)(nil)
