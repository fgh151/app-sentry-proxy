package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type LogClient struct {
	client   *http.Client
	baseURL  string
	username string
	password string
}

func NewLogClient(baseURL, username, password string) *LogClient {
	return &LogClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:  baseURL,
		username: username,
		password: password,
	}
}

func (c *LogClient) FetchLogs(ctx context.Context) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.password)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch logs: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp.Body, nil
} 