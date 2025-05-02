package sentry

import (
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"
)

type Client struct {
	dsn         string
	environment string
	project     string
}

func NewClient(dsn, environment, project string) (*Client, error) {
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		Environment:      environment,
		TracesSampleRate: 1.0,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Sentry client: %w", err)
	}

	return &Client{
		dsn:         dsn,
		environment: environment,
		project:     project,
	}, nil
}

func (c *Client) SendEvent(event *sentry.Event) error {
	eventID := sentry.CaptureEvent(event)
	if eventID == nil {
		return fmt.Errorf("failed to capture event")
	}
	return nil
}

func (c *Client) Flush() {
	sentry.Flush(2 * time.Second)
}
