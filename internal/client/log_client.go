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
	state    *LogState
}

func NewLogClient(baseURL, username, password string, state *LogState) *LogClient {
	return &LogClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:  baseURL,
		username: username,
		password: password,
		state:    state,
	}
}

func (c *LogClient) FetchLogs(ctx context.Context) (io.ReadCloser, error) {
	lastFile, lastPos := c.state.GetLastPosition()

	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth(c.username, c.password)

	// Add Range header to request only new content
	if lastPos > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", lastPos))
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch logs: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		resp.Body.Close()
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Get content length from response
	contentLength := resp.ContentLength
	if resp.StatusCode == http.StatusPartialContent {
		// For partial content, add the range start to content length
		contentRange := resp.Header.Get("Content-Range")
		if contentRange != "" {
			// Parse Content-Range header (e.g., "bytes 1000-2000/3000")
			var start, end, total int64
			fmt.Sscanf(contentRange, "bytes %d-%d/%d", &start, &end, &total)
			contentLength = end - start + 1
		}
	}

	// Create a wrapper around the response body to track position
	reader := &positionTrackingReader{
		reader: resp.Body,
		state:  c.state,
		file:   lastFile,
		pos:    lastPos,
		length: contentLength,
	}

	return reader, nil
}

type positionTrackingReader struct {
	reader io.ReadCloser
	state  *LogState
	file   string
	pos    int64
	length int64
}

func (r *positionTrackingReader) Read(p []byte) (n int, err error) {
	n, err = r.reader.Read(p)
	if n > 0 {
		r.pos += int64(n)
		// Update state periodically (every 1MB or on error)
		if r.pos%1024*1024 == 0 || err != nil {
			if err := r.state.UpdatePosition(r.file, r.pos); err != nil {
				fmt.Printf("Failed to update state: %v\n", err)
			}
		}
	}
	return n, err
}

func (r *positionTrackingReader) Close() error {
	// Update final position before closing
	if err := r.state.UpdatePosition(r.file, r.pos); err != nil {
		fmt.Printf("Failed to update final state: %v\n", err)
	}
	return r.reader.Close()
}
