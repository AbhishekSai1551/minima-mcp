package minima

import (
	"context"
	"fmt"
)

func (c *Client) GetStatus(ctx context.Context) (*NodeStatus, error) {
	result, err := c.DataRequest(ctx, "status", nil)
	if err != nil {
		return nil, fmt.Errorf("get status: %w", err)
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected status response type: %T", result)
	}

	status := &NodeStatus{}
	if v, ok := data["status"]; ok {
		status.Status = toString(v)
	}
	if v, ok := data["synced"]; ok {
		status.Synced = toBool(v)
	}
	if v, ok := data["block"]; ok {
		status.BlockHeight = toInt64(v)
	}
	if v, ok := data["peers"]; ok {
		status.Peers = toInt(toInt64(v))
	}
	if v, ok := data["uptime"]; ok {
		status.Uptime = toInt64(v)
	}
	if v, ok := data["role"]; ok {
		status.Role = toString(v)
	}
	if v, ok := data["version"]; ok {
		status.Version = toString(v)
	}
	if v, ok := data["chain_id"]; ok {
		status.ChainID = toString(v)
	}

	return status, nil
}

func (c *Client) GetBlock(ctx context.Context, height int64) (interface{}, error) {
	params := map[string]interface{}{
		"height": height,
	}
	return c.DataRequest(ctx, "block", params)
}

func (c *Client) GetTransaction(ctx context.Context, txID string) (interface{}, error) {
	if err := ValidateTxID(txID); err != nil {
		return nil, fmt.Errorf("invalid transaction id: %w", err)
	}
	params := map[string]interface{}{
		"txid": txID,
	}
	return c.DataRequest(ctx, "transaction", params)
}

func (c *Client) GetNetworkInfo(ctx context.Context) (interface{}, error) {
	return c.DataRequest(ctx, "network", nil)
}
