package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/minima-global/minima-mcp/internal/audit"
	"github.com/minima-global/minima-mcp/internal/minima"
	"github.com/minima-global/minima-mcp/internal/ratelimit"
)

type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
	Handler     ToolHandler
	AuthLevel   string
}

type ToolHandler func(ctx context.Context, args map[string]interface{}) (interface{}, error)

type MCPServer struct {
	mu        sync.RWMutex
	tools     map[string]*Tool
	client    *minima.Client
	auditLog  *audit.Logger
	rateLimit *ratelimit.RateLimiter
	server    *http.Server
	logger    *slog.Logger
	authMode  string
}

func NewMCPServer(client *minima.Client, auditLog *audit.Logger, rl *ratelimit.RateLimiter, authMode string, logger *slog.Logger) *MCPServer {
	return &MCPServer{
		tools:     make(map[string]*Tool),
		client:    client,
		auditLog:  auditLog,
		rateLimit: rl,
		authMode:  authMode,
		logger:    logger,
	}
}

func (s *MCPServer) RegisterTool(tool *Tool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tools[tool.Name] = tool
}

func (s *MCPServer) GetTool(name string) (*Tool, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tools[name]
	return t, ok
}

func (s *MCPServer) ListTools() []*Tool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	tools := make([]*Tool, 0, len(s.tools))
	for _, t := range s.tools {
		tools = append(tools, t)
	}
	return tools
}

type MCPRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type ToolListResult struct {
	Tools []ToolDef `json:"tools"`
}

type ToolDef struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type ToolCallResult struct {
	Content []ContentBlock `json:"content"`
	IsError bool           `json:"isError,omitempty"`
}

type ContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (s *MCPServer) HandleMCP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	sourceIP := extractIP(r)

	if r.Method != http.MethodPost {
		s.writeError(w, nil, -32600, "invalid request: method not allowed")
		return
	}

	if s.rateLimit != nil {
		result := s.rateLimit.Allow(sourceIP)
		if !result.Allowed {
			s.auditLog.LogRateLimit(r.Context(), sourceIP, "mcp_request")
			w.Header().Set("Retry-After", fmt.Sprintf("%.0f", result.RetryAfter.Seconds()))
			s.writeError(w, nil, -32429, "rate limit exceeded")
			return
		}
	}

	var req MCPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, nil, -32700, "parse error: invalid JSON")
		return
	}

	if req.JSONRPC != "2.0" {
		s.writeError(w, req.ID, -32600, "invalid request: unsupported JSON-RPC version")
		return
	}

	switch req.Method {
	case "tools/list":
		s.handleToolsList(w, req)
	case "tools/call":
		s.handleToolCall(w, r, req, start)
	case "initialize":
		s.handleInitialize(w, req)
	case "ping":
		s.writeResult(w, req.ID, map[string]interface{}{"status": "ok"})
	default:
		s.writeError(w, req.ID, -32601, fmt.Sprintf("method not found: %s", req.Method))
	}
}

func (s *MCPServer) handleToolsList(w http.ResponseWriter, req MCPRequest) {
	tools := s.ListTools()
	defs := make([]ToolDef, 0, len(tools))
	for _, t := range tools {
		defs = append(defs, ToolDef{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: t.InputSchema,
		})
	}
	s.writeResult(w, req.ID, ToolListResult{Tools: defs})
}

func (s *MCPServer) handleToolCall(w http.ResponseWriter, r *http.Request, req MCPRequest, start time.Time) {
	var params ToolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		s.writeError(w, req.ID, -32602, "invalid params: "+err.Error())
		return
	}

	tool, ok := s.GetTool(params.Name)
	if !ok {
		s.writeError(w, req.ID, -32602, fmt.Sprintf("unknown tool: %s", params.Name))
		return
	}

	if tool.AuthLevel != "public" && s.authMode != "false" {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			s.auditLog.LogAuth(r.Context(), "anonymous", tool.Name, extractIP(r), false)
			s.writeError(w, req.ID, -32600, "authentication required")
			return
		}
		s.auditLog.LogAuth(r.Context(), "bearer", tool.Name, extractIP(r), true)
	}

	requestID := generateRequestID()
	sourceIP := extractIP(r)

	s.auditLog.LogToolCall(r.Context(), sourceIP, tool.Name, params.Arguments, requestID, sourceIP)

	result, err := tool.Handler(r.Context(), params.Arguments)
	duration := time.Since(start).Nanoseconds()

	if err != nil {
		s.auditLog.LogToolResult(r.Context(), sourceIP, tool.Name, fmt.Sprintf("error: %s", err.Error()), requestID, duration)
		s.writeResult(w, req.ID, ToolCallResult{
			Content: []ContentBlock{
				{Type: "text", Text: fmt.Sprintf("Error: %s", err.Error())},
			},
			IsError: true,
		})
		return
	}

	resultJSON, _ := json.Marshal(result)
	s.auditLog.LogToolResult(r.Context(), sourceIP, tool.Name, string(resultJSON), requestID, duration)

	s.writeResult(w, req.ID, ToolCallResult{
		Content: []ContentBlock{
			{Type: "text", Text: string(resultJSON)},
		},
	})
}

func (s *MCPServer) handleInitialize(w http.ResponseWriter, req MCPRequest) {
	s.writeResult(w, req.ID, map[string]interface{}{
		"protocolVersion": "2025-03-26",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{
				"listChanged": true,
			},
		},
		"serverInfo": map[string]interface{}{
			"name":    "minima-mcp",
			"version": "1.0.0",
		},
	})
}

func (s *MCPServer) writeResult(w http.ResponseWriter, id interface{}, result interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	})
}

func (s *MCPServer) writeError(w http.ResponseWriter, id interface{}, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	statusCode := http.StatusInternalServerError
	switch code {
	case -32700, -32600, -32602:
		statusCode = http.StatusBadRequest
	case -32601:
		statusCode = http.StatusNotFound
	case -32429:
		statusCode = http.StatusTooManyRequests
	}
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(MCPResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &MCPError{Code: code, Message: message},
	})
}

func extractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}

func generateRequestID() string {
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}
