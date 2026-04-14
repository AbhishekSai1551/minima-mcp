package minima

import (
	"context"
	"fmt"
)

func (c *Client) ListMiniDAPPs(ctx context.Context) (interface{}, error) {
	return c.MDSRequest(ctx, "list", nil)
}

func (c *Client) InstallMiniDAPP(ctx context.Context, appUID, version string) (interface{}, error) {
	params := map[string]string{
		"uid":     appUID,
		"version": version,
	}
	return c.MDSRequest(ctx, "install", params)
}

func (c *Client) UninstallMiniDAPP(ctx context.Context, appUID string) (interface{}, error) {
	params := map[string]string{
		"uid": appUID,
	}
	return c.MDSRequest(ctx, "uninstall", params)
}

func (c *Client) GetMiniDAPPInfo(ctx context.Context, appUID string) (interface{}, error) {
	if err := ValidateMiniDAPPID(appUID); err != nil {
		return nil, fmt.Errorf("invalid minidapp id: %w", err)
	}
	params := map[string]string{
		"uid": appUID,
	}
	return c.MDSRequest(ctx, "info", params)
}

func (c *Client) RunMiniDAPP(ctx context.Context, appUID, command string, args map[string]string) (interface{}, error) {
	if err := ValidateMiniDAPPID(appUID); err != nil {
		return nil, fmt.Errorf("invalid minidapp id: %w", err)
	}
	params := map[string]string{
		"uid":     appUID,
		"command": command,
	}
	for k, v := range args {
		params[k] = v
	}
	return c.MDSRequest(ctx, "run", params)
}
