package parser

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"time"

	"github.com/getsentry/sentry-go"
)

type LogEntry struct {
	Timestamp time.Time
	Level     string
	Message   string
	Context   map[string]interface{}
}

type LogParser struct {
	// Regular expression to match PHP log entries
	// Example format: [2024-03-20 10:00:00] [error] Message here {"key": "value"}
	logPattern *regexp.Regexp
}

func NewLogParser() *LogParser {
	return &LogParser{
		logPattern: regexp.MustCompile(`^\[(.*?)\] \[(.*?)\] (.*?)(?:\s+(\{.*\}))?$`),
	}
}

func (p *LogParser) ParseLogs(reader io.Reader) ([]*LogEntry, error) {
	var entries []*LogEntry
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		line := scanner.Text()
		entry, err := p.parseLine(line)
		if err != nil {
			continue // Skip invalid lines
		}
		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning logs: %w", err)
	}

	return entries, nil
}

func (p *LogParser) parseLine(line string) (*LogEntry, error) {
	matches := p.logPattern.FindStringSubmatch(line)
	if len(matches) < 4 {
		return nil, fmt.Errorf("invalid log format: %s", line)
	}

	timestamp, err := time.Parse("2006-01-02 15:04:05", matches[1])
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp: %w", err)
	}

	entry := &LogEntry{
		Timestamp: timestamp,
		Level:     matches[2],
		Message:   matches[3],
		Context:   make(map[string]interface{}),
	}

	// Parse additional context if present
	if len(matches) > 4 && matches[4] != "" {
		// TODO: Parse JSON context
		// For now, we'll just store the raw JSON string
		entry.Context["raw_context"] = matches[4]
	}

	return entry, nil
}

func (p *LogParser) ToSentryEvent(entry *LogEntry) *sentry.Event {
	level := sentry.LevelInfo
	switch entry.Level {
	case "error":
		level = sentry.LevelError
	case "warning":
		level = sentry.LevelWarning
	case "debug":
		level = sentry.LevelDebug
	}

	return &sentry.Event{
		Timestamp: entry.Timestamp,
		Level:     level,
		Message:   entry.Message,
		Extra:     entry.Context,
	}
} 