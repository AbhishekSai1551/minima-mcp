package minima

import (
	"context"
	"fmt"
)

func (c *Client) GenerateKey(ctx context.Context) (*KeyPair, error) {
	result, err := c.DataRequest(ctx, "keys", map[string]interface{}{
		"action": "generate",
	})
	if err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected key response type: %T", result)
	}

	kp := &KeyPair{}
	if v, ok := data["publickey"]; ok {
		kp.PublicKey = toString(v)
	}
	if v, ok := data["privatekey"]; ok {
		kp.PrivateKey = toString(v)
	}
	if v, ok := data["address"]; ok {
		kp.Address = toString(v)
	}

	return kp, nil
}

func (c *Client) ListKeys(ctx context.Context) (interface{}, error) {
	return c.DataRequest(ctx, "keys", map[string]interface{}{
		"action": "list",
	})
}

func (c *Client) SignMessage(ctx context.Context, privateKey, message string) (interface{}, error) {
	if err := ValidatePrivateKey(privateKey); err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}
	if err := ValidateMessage(message); err != nil {
		return nil, fmt.Errorf("invalid message: %w", err)
	}

	params := map[string]interface{}{
		"action":     "sign",
		"privatekey": privateKey,
		"message":    message,
	}

	return c.DataRequest(ctx, "keys", params)
}

func (c *Client) VerifySignature(ctx context.Context, publicKey, message, signature string) (interface{}, error) {
	if err := ValidatePublicKey(publicKey); err != nil {
		return nil, fmt.Errorf("invalid public key: %w", err)
	}

	params := map[string]interface{}{
		"action":    "verify",
		"publickey": publicKey,
		"message":   message,
		"signature": signature,
	}

	return c.DataRequest(ctx, "keys", params)
}

func (c *Client) GetVault(ctx context.Context) (interface{}, error) {
	return c.DataRequest(ctx, "vault", nil)
}
