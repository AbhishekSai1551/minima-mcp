package minima

import (
	"context"
	"fmt"
)

func (c *Client) GetBalance(ctx context.Context, address string) (*Balance, error) {
	params := map[string]interface{}{}
	if address != "" {
		params["address"] = address
	}
	result, err := c.DataRequest(ctx, "balance", params)
	if err != nil {
		return nil, fmt.Errorf("get balance: %w", err)
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected balance response type: %T", result)
	}

	balance := &Balance{
		Asset: "Minima",
	}
	if v, ok := data["confirmed"]; ok {
		balance.Confirmed = toString(v)
	}
	if v, ok := data["unconfirmed"]; ok {
		balance.Unconfirmed = toString(v)
	}
	if v, ok := data["tokenid"]; ok {
		balance.TokenID = toString(v)
	}

	return balance, nil
}

func (c *Client) SendTransaction(ctx context.Context, to, amount, asset string) (interface{}, error) {
	if err := ValidateAddress(to); err != nil {
		return nil, fmt.Errorf("invalid recipient address: %w", err)
	}
	if err := ValidateAmount(amount); err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	params := map[string]interface{}{
		"to":     to,
		"amount": amount,
	}
	if asset != "" {
		params["asset"] = asset
	}

	return c.DataRequest(ctx, "send", params)
}

func (c *Client) GetTransactions(ctx context.Context, offset, limit int) (interface{}, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	params := map[string]interface{}{
		"offset": offset,
		"limit":  limit,
	}
	return c.DataRequest(ctx, "transactions", params)
}

func (c *Client) GetCoins(ctx context.Context) (interface{}, error) {
	return c.DataRequest(ctx, "coins", nil)
}
