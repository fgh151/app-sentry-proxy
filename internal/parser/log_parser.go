package parser

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
)

type LogEntry struct {
	Timestamp time.Time
	Level     string
	Message   string
	Context   map[string]interface{}
	Stack     []string
}

type LogParser struct {
	// Regular expression to match PHP log entries
	// Example format: 2025-04-30 06:25:17 [172.19.0.2][-][1b9d93016fb9c0c4832c06294ef3d7f7][error][yii\web\HttpException:404] ...
	logPattern *regexp.Regexp
}

func NewLogParser() *LogParser {
	return &LogParser{
		logPattern: regexp.MustCompile(`^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}) \[(.*?)\]\[(.*?)\]\[(.*?)\]\[(.*?)\]\[(.*?)\] (.*)`),
	}
}

func (p *LogParser) ParseLogs(reader io.Reader) ([]*LogEntry, error) {
	var entries []*LogEntry
	scanner := bufio.NewScanner(reader)
	var currentEntry *LogEntry
	var stackTrace []string

	for scanner.Scan() {
		line := scanner.Text()

		// Check if this is a new log entry
		if matches := p.logPattern.FindStringSubmatch(line); len(matches) > 0 {
			// If we have a previous entry with stack trace, add it to entries
			if currentEntry != nil && len(stackTrace) > 0 {
				currentEntry.Stack = stackTrace
				entries = append(entries, currentEntry)
				stackTrace = nil
			}

			// Parse timestamp
			timestamp, err := time.Parse("2006-01-02 15:04:05", matches[1])
			if err != nil {
				continue // Skip invalid timestamp
			}

			// Create new entry
			currentEntry = &LogEntry{
				Timestamp: timestamp,
				Level:     matches[5], // error level
				Message:   matches[6], // error message
				Context: map[string]interface{}{
					"ip":        matches[2],
					"user_id":   matches[3],
					"session":   matches[4],
					"exception": matches[5],
				},
				Stack: make([]string, 0),
			}
		} else if currentEntry != nil && strings.HasPrefix(line, "#") {
			// This is a stack trace line
			stackTrace = append(stackTrace, line)
		}
	}

	// Add the last entry if it exists
	if currentEntry != nil && len(stackTrace) > 0 {
		currentEntry.Stack = stackTrace
		entries = append(entries, currentEntry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning logs: %w", err)
	}

	return entries, nil
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

	// Create stack trace frames
	frames := make([]sentry.Frame, 0, len(entry.Stack))
	for _, line := range entry.Stack {
		if frame := parseStackFrame(line); frame != nil {
			frames = append(frames, *frame)
		}
	}

	// Create exception
	exception := sentry.Exception{
		Type:       entry.Context["exception"].(string),
		Value:      entry.Message,
		Stacktrace: &sentry.Stacktrace{Frames: frames},
	}

	return &sentry.Event{
		Timestamp: entry.Timestamp,
		Level:     level,
		Message:   entry.Message,
		Extra:     entry.Context,
		Exception: []sentry.Exception{exception},
	}
}

func parseStackFrame(line string) *sentry.Frame {
	// Example line: #0 /app/vendor/yiisoft/yii2/base/Module.php(561): yii\base\Module->runAction('assets/61112d37...', Array)
	parts := strings.SplitN(line, " ", 2)
	if len(parts) != 2 {
		return nil
	}

	// Parse file and line number
	fileAndLine := strings.SplitN(parts[1], "(", 2)
	if len(fileAndLine) != 2 {
		return nil
	}

	file := fileAndLine[0]
	lineNumberStr := strings.TrimSuffix(strings.SplitN(fileAndLine[1], ")", 2)[0], ")")
	lineNumber, err := strconv.Atoi(lineNumberStr)
	if err != nil {
		return nil
	}

	// Parse function
	function := ""
	if strings.Contains(parts[1], ":") {
		funcParts := strings.SplitN(parts[1], ":", 2)
		if len(funcParts) == 2 {
			function = strings.TrimSpace(funcParts[1])
		}
	}

	return &sentry.Frame{
		Filename: file,
		Lineno:   lineNumber,
		Function: function,
	}
}
