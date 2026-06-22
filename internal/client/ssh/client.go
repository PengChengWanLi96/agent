package ssh

import (
	"fmt"
	"io"
	"net"
	"path"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type ConnectOptions struct {
	Host       string
	Port       int
	User       string
	Password   string
	PrivateKey string
	Local      bool
}

type Client struct {
	client *ssh.Client
	sftp   *sftp.Client
}

func NewClient(opts ConnectOptions) (*Client, error) {
	if opts.Port == 0 {
		opts.Port = 22
	}

	authMethods := make([]ssh.AuthMethod, 0)
	if opts.Password != "" {
		authMethods = append(authMethods, ssh.Password(opts.Password))
	}
	if opts.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(opts.PrivateKey))
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}
	if len(authMethods) == 0 {
		return nil, fmt.Errorf("password or private_key is required")
	}

	config := &ssh.ClientConfig{
		User:            opts.User,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}

	addr := net.JoinHostPort(opts.Host, fmt.Sprintf("%d", opts.Port))
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to dial ssh %s: %w", addr, err)
	}

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create sftp client: %w", err)
	}

	return &Client{
		client: client,
		sftp:   sftpClient,
	}, nil
}

func (c *Client) Close() error {
	if c.sftp != nil {
		c.sftp.Close()
	}
	if c.client != nil {
		c.client.Close()
	}
	return nil
}

func (c *Client) ListDir(path string) ([]FileInfo, error) {
	entries, err := c.sftp.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read dir %s: %w", path, err)
	}

	items := make([]FileInfo, 0, len(entries))
	for _, entry := range entries {
		items = append(items, FileInfo{
			Name:    entry.Name(),
			Path:    c.sftp.Join(path, entry.Name()),
			Size:    entry.Size(),
			IsDir:   entry.IsDir(),
			Mode:    entry.Mode().String(),
			ModTime: entry.ModTime().Unix(),
		})
	}
	return items, nil
}

func (c *Client) Download(path string) (io.ReadCloser, error) {
	file, err := c.sftp.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", path, err)
	}
	return file, nil
}

func (c *Client) Upload(pathStr string, reader io.Reader) error {
	if err := c.sftp.MkdirAll(path.Dir(pathStr)); err != nil {
		return fmt.Errorf("failed to create parent dir: %w", err)
	}
	file, err := c.sftp.Create(pathStr)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", pathStr, err)
	}
	defer file.Close()

	if _, err := io.Copy(file, reader); err != nil {
		return fmt.Errorf("failed to write file %s: %w", pathStr, err)
	}
	return nil
}

func (c *Client) Remove(path string) error {
	info, err := c.sftp.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat %s: %w", path, err)
	}
	if info.IsDir() {
		if err := c.sftp.RemoveDirectory(path); err != nil {
			return fmt.Errorf("failed to remove directory %s: %w", path, err)
		}
	} else {
		if err := c.sftp.Remove(path); err != nil {
			return fmt.Errorf("failed to remove file %s: %w", path, err)
		}
	}
	return nil
}

func (c *Client) Mkdir(path string) error {
	if err := c.sftp.MkdirAll(path); err != nil {
		return fmt.Errorf("failed to mkdir %s: %w", path, err)
	}
	return nil
}

func (c *Client) Rename(oldPath, newPath string) error {
	if err := c.sftp.Rename(oldPath, newPath); err != nil {
		return fmt.Errorf("failed to rename %s -> %s: %w", oldPath, newPath, err)
	}
	return nil
}

func (c *Client) Exec(command string) (string, int, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", 0, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			exitCode = exitErr.ExitStatus()
		} else {
			return "", 0, fmt.Errorf("failed to run command: %w", err)
		}
	}
	return string(output), exitCode, nil
}

type FileInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	IsDir   bool   `json:"is_dir"`
	Mode    string `json:"mode"`
	ModTime int64  `json:"mod_time"`
}
