//go:build !linux

package service

import (
	"errors"
	"io"
)

type localTerminal struct{}

func newLocalTerminal() (*localTerminal, error) {
	return nil, errors.New("interactive terminal is only available on Linux")
}

func (t *localTerminal) Read(p []byte) (int, error) {
	return 0, io.EOF
}

func (t *localTerminal) Write(p []byte) (int, error) {
	return 0, errors.New("interactive terminal is only available on Linux")
}

func (t *localTerminal) Close() error {
	return nil
}

func (t *localTerminal) Resize(rows, cols int) error {
	return errors.New("interactive terminal is only available on Linux")
}

var _ terminalSession = (*localTerminal)(nil)
var _ io.ReadWriteCloser = (*localTerminal)(nil)
