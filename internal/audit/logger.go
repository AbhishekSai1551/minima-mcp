package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type EventType string

const (
	EventToolCall   EventType = "tool_call"
	EventToolResult EventType = "tool_result"
	EventAuth       EventType = "auth"
	EventRateLimit  EventType = "rate_limit"
	EventError      EventType = "error"
	EventSystem     EventType = "system"
)

type Event struct {
	Timestamp  time.Time              `json:"timestamp"`
	EventType  EventType              `json:"event_type"`
	Actor      string                 `json:"actor"`
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource,omitempty"`
	RequestID  string                 `json:"request_id"`
	SourceIP   string                 `json:"source_ip,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	Result     string                 `json:"result,omitempty"`
	Error      string                 `json:"error,omitempty"`
	StatusCode int                    `json:"status_code,omitempty"`
	Duration   int64                  `json:"duration_ns,omitempty"`
}

type Logger struct {
	mu       sync.Mutex
	logDir   string
	maxAge   int
	format   string
	file     *os.File
	current  string
	disabled bool
}

func NewLogger(logDir string, maxAgeDays int, format string) (*Logger, error) {
	l := &Logger{
		logDir: logDir,
		maxAge: maxAgeDays,
		format: format,
	}

	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("create audit log dir: %w", err)
	}

	if err := l.rotateFile(time.Now()); err != nil {
		return nil, fmt.Errorf("initial audit log rotation: %w", err)
	}

	return l, nil
}

func (l *Logger) Log(ctx context.Context, evt Event) error {
	if l.noop() {
		return nil
	}

	if evt.Timestamp.IsZero() {
		evt.Timestamp = time.Now().UTC()
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	day := evt.Timestamp.Format("2006-01-02")
	if day != l.current {
		if err := l.rotateFile(evt.Timestamp); err != nil {
			return err
		}
	}

	var line []byte
	var err error
	switch l.format {
	case "json":
		line, err = json.Marshal(evt)
	default:
		line, err = json.Marshal(evt)
	}
	if err != nil {
		return fmt.Errorf("marshal audit event: %w", err)
	}

	line = append(line, '\n')
	if _, err := l.file.Write(line); err != nil {
		return fmt.Errorf("write audit event: %w", err)
	}

	return nil
}

func (l *Logger) LogToolCall(ctx context.Context, actor, toolName string, params map[string]interface{}, requestID, sourceIP string) error {
	sanitized := sanitizeParams(params)
	return l.Log(ctx, Event{
		EventType:  EventToolCall,
		Actor:      actor,
		Action:     toolName,
		RequestID:  requestID,
		SourceIP:   sourceIP,
		Parameters: sanitized,
	})
}

func (l *Logger) LogToolResult(ctx context.Context, actor, toolName, result string, requestID string, duration int64) error {
	return l.Log(ctx, Event{
		EventType: EventToolResult,
		Actor:     actor,
		Action:    toolName,
		RequestID: requestID,
		Result:    truncate(result, 4096),
		Duration:  duration,
	})
}

func (l *Logger) LogAuth(ctx context.Context, actor, action, sourceIP string, success bool) error {
	evt := Event{
		EventType: EventAuth,
		Actor:     actor,
		Action:    action,
		SourceIP:  sourceIP,
	}
	if success {
		evt.Result = "success"
	} else {
		evt.Result = "denied"
		evt.Error = "authentication_failed"
	}
	return l.Log(ctx, evt)
}

func (l *Logger) LogRateLimit(ctx context.Context, sourceIP, toolName string) error {
	return l.Log(ctx, Event{
		EventType: EventRateLimit,
		Action:    toolName,
		SourceIP:  sourceIP,
		Error:     "rate_limit_exceeded",
	})
}

func (l *Logger) LogError(ctx context.Context, actor, action, errMsg string) error {
	return l.Log(ctx, Event{
		EventType: EventError,
		Actor:     actor,
		Action:    action,
		Error:     truncate(errMsg, 4096),
	})
}

func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

func (l *Logger) PurgeOldLogs() error {
	if l.maxAge <= 0 {
		return nil
	}

	entries, err := os.ReadDir(l.logDir)
	if err != nil {
		return fmt.Errorf("read audit log dir: %w", err)
	}

	cutoff := time.Now().AddDate(0, 0, -l.maxAge)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			os.Remove(filepath.Join(l.logDir, entry.Name()))
		}
	}
	return nil
}

func (l *Logger) rotateFile(t time.Time) error {
	if l.file != nil {
		l.file.Close()
	}

	day := t.Format("2006-01-02")
	filename := filepath.Join(l.logDir, fmt.Sprintf("audit-%s.log", day))

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open audit log file: %w", err)
	}

	l.file = f
	l.current = day
	return nil
}

func sanitizeParams(params map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{}, len(params))
	sensitiveKeys := map[string]bool{
		"privatekey":  true,
		"private_key": true,
		"secret":      true,
		"password":    true,
		"seed":        true,
		"mnemonic":    true,
	}
	for k, v := range params {
		if sensitiveKeys[k] || sensitiveKeys[strings.ToLower(k)] {
			sanitized[k] = "[REDACTED]"
		} else {
			sanitized[k] = v
		}
	}
	return sanitized
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
