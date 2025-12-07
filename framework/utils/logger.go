package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// StructuredLogger prints log entries with a consistent structure so that
// runners and developers can trace execution across suites and steps.
type StructuredLogger struct {
	mu     sync.Mutex
	fields map[string]string
}

func NewLogger() *StructuredLogger {
	return &StructuredLogger{fields: map[string]string{}}
}

func (l *StructuredLogger) clone() *StructuredLogger {
	cp := make(map[string]string, len(l.fields))
	for k, v := range l.fields {
		cp[k] = v
	}
	return &StructuredLogger{fields: cp}
}

func (l *StructuredLogger) With(key, value string) *StructuredLogger {
	child := l.clone()
	child.fields[key] = value
	return child
}

func (l *StructuredLogger) Info(message string, fields map[string]any) {
	l.write("INFO", message, fields)
}

func (l *StructuredLogger) Error(message string, fields map[string]any) {
	l.write("ERROR", message, fields)
}

func (l *StructuredLogger) write(level, message string, fields map[string]any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := map[string]any{
		"level":   level,
		"message": message,
		"time":    time.Now().UTC().Format(time.RFC3339),
	}
	for k, v := range l.fields {
		entry[k] = v
	}
	for k, v := range fields {
		entry[k] = v
	}

	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stdout, "level=%s message=%s error=%v\n", level, message, err)
		return
	}

	fmt.Fprintln(os.Stdout, string(data))
}
