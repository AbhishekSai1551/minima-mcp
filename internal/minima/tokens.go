package minima

import (
	"context"
	"fmt"
)

func (c *Client) CreateToken(ctx context.Context, name, amount string, decimals int) (interface{}, error) {
	if err := ValidateTokenName(name); err != nil {
		return nil, fmt.Errorf("invalid token name: %w", err)
	}
	if err := ValidateAmount(amount); err != nil {
		return nil, fmt.Errorf("invalid token amount: %w", err)
	}

	params := map[string]interface{}{
		"action":   "create",
		"name":     name,
		"amount":   amount,
		"decimals": decimals,
	}

	return c.DataRequest(ctx, "token", params)
}

func (c *Client) ListTokens(ctx context.Context) (interface{}, error) {
	params := map[string]interface{}{
		"action": "list",
	}
	return c.DataRequest(ctx, "token", params)
}

func (c *Client) TransferToken(ctx context.Context, tokenID, to, amount string) (interface{}, error) {
	if err := ValidateTokenID(tokenID); err != nil {
		return nil, fmt.Errorf("invalid token id: %w", err)
	}
	if err := ValidateAddress(to); err != nil {
		return nil, fmt.Errorf("invalid recipient: %w", err)
	}
	if err := ValidateAmount(amount); err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	params := map[string]interface{}{
		"action":  "transfer",
		"tokenid": tokenID,
		"to":      to,
		"amount":  amount,
	}

	return c.DataRequest(ctx, "token", params)
}

func (c *Client) GetTokenInfo(ctx context.Context, tokenID string) (interface{}, error) {
	if err := ValidateTokenID(tokenID); err != nil {
		return nil, fmt.Errorf("invalid token id: %w", err)
	}

	params := map[string]interface{}{
		"action":  "info",
		"tokenid": tokenID,
	}
	return c.DataRequest(ctx, "token", params)
}
