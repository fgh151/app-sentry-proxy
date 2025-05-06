package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/openitstudio/app-sentry-proxy/internal/client"
	"github.com/openitstudio/app-sentry-proxy/internal/parser"
	"github.com/openitstudio/app-sentry-proxy/internal/sentry"
	"github.com/openitstudio/app-sentry-proxy/pkg/config"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize state manager
	state, err := client.NewLogState(cfg.Server.StateFile)
	if err != nil {
		log.Fatalf("Failed to initialize state manager: %v", err)
	}

	// Initialize clients
	logClient := client.NewLogClient(cfg.Server.LogURL, cfg.Server.Username, cfg.Server.Password, state)
	logParser := parser.NewLogParser()
	sentryClient, err := sentry.NewClient(cfg.Sentry.DSN, cfg.Sentry.Environment, cfg.Sentry.Project)
	if err != nil {
		log.Fatalf("Failed to initialize Sentry client: %v", err)
	}
	defer sentryClient.Flush()

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Main loop
	ticker := time.NewTicker(cfg.Server.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := processLogs(ctx, logClient, logParser, sentryClient); err != nil {
				log.Printf("Error processing logs: %v", err)
			}
		case sig := <-sigChan:
			log.Printf("Received signal %v, shutting down...", sig)
			return
		}
	}
}

func processLogs(ctx context.Context, logClient *client.LogClient, logParser *parser.LogParser, sentryClient *sentry.Client) error {
	// Fetch logs
	reader, err := logClient.FetchLogs(ctx)
	if err != nil {
		return err
	}
	defer reader.Close()

	// Parse logs
	entries, err := logParser.ParseLogs(reader)
	if err != nil {
		return err
	}

	// Send to Sentry
	for _, entry := range entries {
		event := logParser.ToSentryEvent(entry)
		if err := sentryClient.SendEvent(event); err != nil {
			log.Printf("Failed to send event to Sentry: %v", err)
		}
	}

	return nil
}
