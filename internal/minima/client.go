package minima

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/minima-global/minima-mcp/internal/config"
)

type Client struct {
	dataClient  *http.Client
	mdsClient   *http.Client
	cfg         *config.MinimaConfig
	dataBaseURL string
	mdsBaseURL  string
}

func NewClient(cfg *config.MinimaConfig) *Client {
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	}

	return &Client{
		dataClient: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: transport,
		},
		mdsClient: &http.Client{
			Timeout:   cfg.Timeout,
			Transport: transport,
		},
		cfg:         cfg,
		dataBaseURL: cfg.DataURL(),
		mdsBaseURL:  cfg.MiniDAPPURL(),
	}
}

func (c *Client) DataRequest(ctx context.Context, method string, params interface{}) (interface{}, error) {
	reqBody := RPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  method,
		Params:  params,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	var resp *http.Response
	for attempt := 0; attempt <= c.cfg.RetryAttempts; attempt++ {
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.dataBaseURL, bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("create request: %w", err)
		}
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err = c.dataClient.Do(httpReq)
		if err != nil {
			if attempt < c.cfg.RetryAttempts {
				time.Sleep(c.cfg.RetryDelay)
				continue
			}
			return nil, fmt.Errorf("data request after %d retries: %w", c.cfg.RetryAttempts, err)
		}
		break
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("data request failed status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var rpcResp RPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("rpc error code=%d msg=%s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	return rpcResp.Result, nil
}

func (c *Client) MDSRequest(ctx context.Context, action string, params map[string]string) (interface{}, error) {
	url := fmt.Sprintf("%s/mds?action=%s", c.mdsBaseURL, action)
	for k, v := range params {
		url += fmt.Sprintf("&%s=%s", k, v)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create mds request: %w", err)
	}

	resp, err := c.mdsClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("mds request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read mds response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mds request failed status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var result interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return string(respBody), nil
	}

	return result, nil
}
