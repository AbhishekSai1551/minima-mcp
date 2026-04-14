package audit

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoggerLogAndRotate(t *testing.T) {
	dir := t.TempDir()
	l, err := NewLogger(dir, 90, "json")
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	defer l.Close()

	evt := Event{
		EventType: EventToolCall,
		Actor:     "test-actor",
		Action:    "minima_get_status",
		RequestID: "req_test_123",
		SourceIP:  "127.0.0.1",
	}

	if err := l.Log(context.Background(), evt); err != nil {
		t.Fatalf("Log: %v", err)
	}

	day := time.Now().UTC().Format("2006-01-02")
	filename := filepath.Join(dir, "audit-"+day+".log")

	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var logged Event
	if err := json.Unmarshal(data, &logged); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if logged.Actor != "test-actor" {
		t.Errorf("logged actor = %q, want %q", logged.Actor, "test-actor")
	}
	if logged.Action != "minima_get_status" {
		t.Errorf("logged action = %q, want %q", logged.Action, "minima_get_status")
	}
}

func TestLoggerLogToolCall(t *testing.T) {
	dir := t.TempDir()
	l, err := NewLogger(dir, 90, "json")
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	defer l.Close()

	params := map[string]interface{}{
		"privatekey": "super-secret-key",
		"to":         "0xabc",
	}

	if err := l.LogToolCall(context.Background(), "actor", "minima_send", params, "req_1", "10.0.0.1"); err != nil {
		t.Fatalf("LogToolCall: %v", err)
	}

	day := time.Now().UTC().Format("2006-01-02")
	filename := filepath.Join(dir, "audit-"+day+".log")

	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var logged Event
	if err := json.Unmarshal(data, &logged); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if logged.EventType != EventToolCall {
		t.Errorf("event type = %q, want %q", logged.EventType, EventToolCall)
	}

	if logged.Parameters["privatekey"] != "[REDACTED]" {
		t.Error("privatekey should be redacted in audit log")
	}
	if logged.Parameters["to"] != "0xabc" {
		t.Error("to should not be redacted")
	}
}

func TestLoggerLogAuth(t *testing.T) {
	dir := t.TempDir()
	l, err := NewLogger(dir, 90, "json")
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	defer l.Close()

	if err := l.LogAuth(context.Background(), "bearer", "minima_send", "10.0.0.1", true); err != nil {
		t.Fatalf("LogAuth success: %v", err)
	}
	if err := l.LogAuth(context.Background(), "anonymous", "minima_send", "10.0.0.1", false); err != nil {
		t.Fatalf("LogAuth denied: %v", err)
	}

	day := time.Now().UTC().Format("2006-01-02")
	data, err := os.ReadFile(filepath.Join(dir, "audit-"+day+".log"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	lines := splitLines(data)
	if len(lines) < 2 {
		t.Fatalf("expected at least 2 lines, got %d", len(lines))
	}

	var successEvt, deniedEvt Event
	json.Unmarshal([]byte(lines[0]), &successEvt)
	json.Unmarshal([]byte(lines[1]), &deniedEvt)

	if successEvt.Result != "success" {
		t.Error("auth success event should have result=success")
	}
	if deniedEvt.Result != "denied" {
		t.Error("auth denied event should have result=denied")
	}
}

func TestNoopLogger(t *testing.T) {
	l := &Logger{}

	err := l.Log(context.Background(), Event{
		EventType: EventToolCall,
		Action:    "test",
	})
	if err != nil {
		t.Errorf("noop Log should not error: %v", err)
	}
}

func TestPurgeOldLogs(t *testing.T) {
	dir := t.TempDir()

	l, err := NewLogger(dir, 90, "json")
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	l.Close()

	oldFile := filepath.Join(dir, "audit-2020-01-01.log")
	if err := os.WriteFile(oldFile, []byte("{}\n"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	oldTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatalf("Chtimes: %v", err)
	}

	l2, err := NewLogger(dir, 1, "json")
	if err != nil {
		t.Fatalf("NewLogger: %v", err)
	}
	defer l2.Close()

	if err := l2.PurgeOldLogs(); err != nil {
		t.Fatalf("PurgeOldLogs: %v", err)
	}

	if _, e := os.Stat(oldFile); !os.IsNotExist(e) {
		t.Error("old log file should have been purged")
	}
}

func splitLines(data []byte) []string {
	var lines []string
	start := 0
	for i, b := range data {
		if b == '\n' {
			line := string(data[start:i])
			if line != "" {
				lines = append(lines, line)
			}
			start = i + 1
		}
	}
	if start < len(data) {
		lines = append(lines, string(data[start:]))
	}
	return lines
}

func TestSanitizeParams(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name:     "redacts_private_key",
			input:    map[string]interface{}{"privatekey": "secret123", "name": "test"},
			expected: map[string]interface{}{"privatekey": "[REDACTED]", "name": "test"},
		},
		{
			name:     "redacts_password",
			input:    map[string]interface{}{"password": "mypass", "action": "send"},
			expected: map[string]interface{}{"password": "[REDACTED]", "action": "send"},
		},
		{
			name:     "preserves_normal_fields",
			input:    map[string]interface{}{"to": "0xabc", "amount": "10"},
			expected: map[string]interface{}{"to": "0xabc", "amount": "10"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizeParams(tt.input)
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("sanitizeParams()[%q] = %v, want %v", k, result[k], v)
				}
			}
		})
	}
}
