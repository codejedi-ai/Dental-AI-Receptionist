package util

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ToolCallLog records a single tool call entry.
type ToolCallLog struct {
	Timestamp  string          `json:"timestamp"`
	ToolName   string          `json:"tool_name"`
	Arguments  json.RawMessage `json:"arguments"`
	Result     string          `json:"result"`
	Status     string          `json:"status"`
}

var (
	logFile *os.File
	logMu   sync.Mutex
	logInit sync.Once
)

// openLogFile opens (or creates) the tool call log file for append.
func openLogFile() error {
	var err error
	logDir := os.Getenv("LOG_DIR")
	if logDir == "" {
		logDir = "logs"
	}
	if err = os.MkdirAll(logDir, 0755); err != nil {
		return err
	}
	today := time.Now().Format("2006-01-02")
	path := filepath.Join(logDir, fmt.Sprintf("tool_calls_%s.log", today))
	logFile, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	return err
}

// LogToolCall writes a tool call entry to the daily log file.
// It is safe for concurrent use.
func LogToolCall(toolName string, arguments json.RawMessage, result, status string) {
	logInit.Do(func() {
		if err := openLogFile(); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Failed to open tool call log: %v\n", err)
		}
	})

	logMu.Lock()
	defer logMu.Unlock()

	entry := ToolCallLog{
		Timestamp:  time.Now().Format(time.RFC3339),
		ToolName:   toolName,
		Arguments:  arguments,
		Result:     result,
		Status:     status,
	}

	line, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Failed to marshal tool call log: %v\n", err)
		return
	}

	if logFile == nil {
		// Fallback: print to stderr if file couldn't be opened
		fmt.Fprintf(os.Stderr, "⚠️  Tool call log file not available: %s\n", string(line))
		return
	}

	if _, err := logFile.Write(append(line, '\n')); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Failed to write tool call log: %v\n", err)
	}
}
