package audit

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// Entry represents a single audit log record.
type Entry struct {
	Timestamp time.Time `json:"timestamp"`
	Operation string    `json:"operation"`
	Namespace string    `json:"namespace"`
	Path      string    `json:"path"`
	DryRun    bool      `json:"dry_run"`
	Status    string    `json:"status"`
	Message   string    `json:"message,omitempty"`
}

// Logger writes structured audit entries to a writer.
type Logger struct {
	out io.Writer
}

// New creates a Logger writing to the given writer.
// If w is nil, os.Stdout is used.
func New(w io.Writer) *Logger {
	if w == nil {
		w = os.Stdout
	}
	return &Logger{out: w}
}

// Log writes a single audit entry as a JSON line.
func (l *Logger) Log(op, namespace, path string, dryRun bool, status, message string) error {
	e := Entry{
		Timestamp: time.Now().UTC(),
		Operation: op,
		Namespace: namespace,
		Path:      path,
		DryRun:    dryRun,
		Status:    status,
		Message:   message,
	}
	b, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("audit: marshal entry: %w", err)
	}
	_, err = fmt.Fprintf(l.out, "%s\n", b)
	return err
}

// LogRotate is a convenience wrapper for rotation audit events.
func (l *Logger) LogRotate(namespace, path string, dryRun bool, status string) error {
	return l.Log("rotate", namespace, path, dryRun, status, "")
}

// LogSync is a convenience wrapper for sync audit events.
func (l *Logger) LogSync(srcNamespace, dstNamespace, path string, dryRun bool, status string) error {
	msg := fmt.Sprintf("src=%s dst=%s", srcNamespace, dstNamespace)
	return l.Log("sync", dstNamespace, path, dryRun, status, msg)
}
