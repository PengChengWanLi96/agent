package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"agent/internal/config"
	"agent/internal/model"
)

type Client struct {
	client     *http.Client
	baseURL    string
	apiVersion string
}

type containerListResponse struct {
	ID      string   `json:"Id"`
	Names   []string `json:"Names"`
	Image   string   `json:"Image"`
	State   string   `json:"State"`
	Status  string   `json:"Status"`
	Created int64    `json:"Created"`
}

type containerInspectResponse struct {
	ID     string `json:"Id"`
	Name   string `json:"Name"`
	Config struct {
		Image    string   `json:"Image"`
		Hostname string   `json:"Hostname"`
		Env      []string `json:"Env"`
		Cmd      []string `json:"Cmd"`
	} `json:"Config"`
	State struct {
		Status     string    `json:"Status"`
		Running    bool      `json:"Running"`
		Paused     bool      `json:"Paused"`
		Restarting bool      `json:"Restarting"`
		StartedAt  time.Time `json:"StartedAt"`
		FinishedAt time.Time `json:"FinishedAt"`
	} `json:"State"`
	NetworkSettings struct {
		Ports map[string]interface{} `json:"Ports"`
	} `json:"NetworkSettings"`
}

func NewClient(cfg config.DockerConfig) (*Client, error) {
	transport := &http.Transport{}
	
	if cfg.Host == "" {
		cfg.Host = "unix:///var/run/docker.sock"
	}

	var baseURL string
	if cfg.Host[:7] == "unix://" {
		socketPath := cfg.Host[7:]
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return net.Dial("unix", socketPath)
		}
		baseURL = "http://localhost"
	} else {
		baseURL = cfg.Host
	}

	return &Client{
		client: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
		baseURL:    baseURL,
		apiVersion: cfg.APIVersion,
	}, nil
}

func (c *Client) apiURL(path string) string {
	return fmt.Sprintf("%s/%s%s", c.baseURL, c.apiVersion, path)
}

func (c *Client) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.apiURL("/_ping"), nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("docker ping failed: %s", resp.Status)
	}
	return nil
}

func (c *Client) ListContainers(ctx context.Context, all bool) ([]model.Container, error) {
	url := c.apiURL("/containers/json")
	if all {
		url += "?all=1"
	}
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list containers failed: %s - %s", resp.Status, string(body))
	}

	var containers []containerListResponse
	if err := json.NewDecoder(resp.Body).Decode(&containers); err != nil {
		return nil, err
	}

	var result []model.Container
	for _, ctn := range containers {
		result = append(result, model.Container{
			ID:      ctn.ID[:min(len(ctn.ID), 12)],
			Names:   ctn.Names,
			Image:   ctn.Image,
			State:   ctn.State,
			Status:  ctn.Status,
			Created: ctn.Created,
		})
	}
	return result, nil
}

func (c *Client) InspectContainer(ctx context.Context, id string) (*model.ContainerDetail, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.apiURL(fmt.Sprintf("/containers/%s/json", id)), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("inspect container failed: %s - %s", resp.Status, string(body))
	}

	var ctn containerInspectResponse
	if err := json.NewDecoder(resp.Body).Decode(&ctn); err != nil {
		return nil, err
	}

	return &model.ContainerDetail{
		ID:    ctn.ID,
		Name:  ctn.Name,
		Image: ctn.Config.Image,
		State: model.ContainerState{
			Status:     ctn.State.Status,
			Running:    ctn.State.Running,
			Paused:     ctn.State.Paused,
			Restarting: ctn.State.Restarting,
			StartedAt:  ctn.State.StartedAt,
			FinishedAt: ctn.State.FinishedAt,
		},
		Config: map[string]interface{}{
			"hostname": ctn.Config.Hostname,
			"env":      ctn.Config.Env,
			"cmd":      ctn.Config.Cmd,
		},
		Network: map[string]interface{}{
			"ports": ctn.NetworkSettings.Ports,
		},
	}, nil
}

func (c *Client) StartContainer(ctx context.Context, id string) error {
	req, err := http.NewRequestWithContext(ctx, "POST", c.apiURL(fmt.Sprintf("/containers/%s/start", id)), nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("start container failed: %s - %s", resp.Status, string(body))
	}
	return nil
}

func (c *Client) StopContainer(ctx context.Context, id string, timeout int) error {
	url := c.apiURL(fmt.Sprintf("/containers/%s/stop?t=%d", id, timeout))
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("stop container failed: %s - %s", resp.Status, string(body))
	}
	return nil
}

func (c *Client) RemoveContainer(ctx context.Context, id string, force bool) error {
	url := c.apiURL(fmt.Sprintf("/containers/%s", id))
	if force {
		url += "?force=1"
	}
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("remove container failed: %s - %s", resp.Status, string(body))
	}
	return nil
}

func (c *Client) ContainerLogs(ctx context.Context, id string, tail string) (string, error) {
	url := c.apiURL(fmt.Sprintf("/containers/%s/logs?stdout=1&stderr=1&tail=%s", id, tail))
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("container logs failed: %s - %s", resp.Status, string(body))
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (c *Client) Close() error {
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
