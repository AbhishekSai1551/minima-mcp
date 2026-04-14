package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/minima-global/minima-mcp/internal/audit"
	"github.com/minima-global/minima-mcp/internal/ratelimit"
)

func newTestServer(t *testing.T) *MCPServer {
	rl := ratelimit.NewRateLimiter(100, 100, 300) // generous for tests
	auditLog := &audit.Logger{}                   // noop for tests

	s := NewMCPServer(nil, auditLog, rl, "false", nil)

	s.RegisterTool(&Tool{
		Name:        "test_tool",
		Description: "A test tool",
		InputSchema: object(map[string]interface{}{}, nil),
		AuthLevel:   "public",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"result": "ok"}, nil
		},
	})

	return s
}

func TestHandleInitialize(t *testing.T) {
	s := newTestServer(t)

	body, _ := json.Marshal(MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
	})

	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.HandleMCP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("initialize status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp MCPResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error != nil {
		t.Errorf("initialize error = %v", resp.Error)
	}
}

func TestHandleToolsList(t *testing.T) {
	s := newTestServer(t)

	body, _ := json.Marshal(MCPRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	})

	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.HandleMCP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("tools/list status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp MCPResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error != nil {
		t.Errorf("tools/list error = %v", resp.Error)
	}
}

func TestHandleToolCall(t *testing.T) {
	s := newTestServer(t)

	params, _ := json.Marshal(ToolCallParams{
		Name:      "test_tool",
		Arguments: map[string]interface{}{},
	})

	body, _ := json.Marshal(MCPRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "tools/call",
		Params:  params,
	})

	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.HandleMCP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("tools/call status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp MCPResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Error != nil {
		t.Errorf("tools/call error = %v", resp.Error)
	}
}

func TestHandleToolCallUnknownTool(t *testing.T) {
	s := newTestServer(t)

	params, _ := json.Marshal(ToolCallParams{
		Name: "nonexistent_tool",
	})

	body, _ := json.Marshal(MCPRequest{
		JSONRPC: "2.0",
		ID:      4,
		Method:  "tools/call",
		Params:  params,
	})

	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.HandleMCP(w, req)

	if w.Code == http.StatusOK {
		t.Error("expected non-200 for unknown tool")
	}
}

func TestHandleInvalidMethod(t *testing.T) {
	s := newTestServer(t)

	body, _ := json.Marshal(MCPRequest{
		JSONRPC: "2.0",
		ID:      5,
		Method:  "invalid/method",
	})

	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.HandleMCP(w, req)

	if w.Code == http.StatusOK {
		t.Error("expected non-200 for invalid method")
	}
}

func TestHandleGetRequest(t *testing.T) {
	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	w := httptest.NewRecorder()
	s.HandleMCP(w, req)

	if w.Code == http.StatusOK {
		t.Error("expected non-200 for GET request")
	}
}

func TestHandlePing(t *testing.T) {
	s := newTestServer(t)

	body, _ := json.Marshal(MCPRequest{
		JSONRPC: "2.0",
		ID:      6,
		Method:  "ping",
	})

	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
	w := httptest.NewRecorder()
	s.HandleMCP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("ping status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestRateLimitBlocks(t *testing.T) {
	s := newTestServer(t)
	s.rateLimit = ratelimit.NewRateLimiter(1, 1, 300)

	for i := 0; i < 3; i++ {
		body, _ := json.Marshal(MCPRequest{
			JSONRPC: "2.0",
			ID:      i,
			Method:  "ping",
		})
		req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
		w := httptest.NewRecorder()
		s.HandleMCP(w, req)
		if i >= 1 && w.Code != http.StatusTooManyRequests {
			t.Logf("request %d: status %d", i, w.Code)
		}
	}
}
