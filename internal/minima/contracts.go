package minima

import (
	"context"
	"fmt"
)

func (c *Client) DeployContract(ctx context.Context, script string, stateParams map[string]string) (*ContractResult, error) {
	if err := ValidateScript(script); err != nil {
		return nil, fmt.Errorf("invalid contract script: %w", err)
	}

	params := map[string]interface{}{
		"script": script,
	}
	for k, v := range stateParams {
		params[k] = v
	}

	result, err := c.DataRequest(ctx, "contract", params)
	if err != nil {
		return nil, fmt.Errorf("deploy contract: %w", err)
	}

	data, ok := result.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected contract response type: %T", result)
	}

	contractResult := &ContractResult{}
	if v, ok := data["contractid"]; ok {
		contractResult.ContractID = toString(v)
	}
	if v, ok := data["status"]; ok {
		contractResult.Status = toString(v)
	}
	if v, ok := data["message"]; ok {
		contractResult.Message = toString(v)
	}
	return contractResult, nil
}

func (c *Client) ExecuteContract(ctx context.Context, contractID, functionName string, args map[string]string) (interface{}, error) {
	if err := ValidateContractID(contractID); err != nil {
		return nil, fmt.Errorf("invalid contract id: %w", err)
	}
	if err := ValidateFunctionName(functionName); err != nil {
		return nil, fmt.Errorf("invalid function name: %w", err)
	}

	params := map[string]interface{}{
		"contractid":   contractID,
		"functionname": functionName,
	}
	for k, v := range args {
		params[k] = v
	}

	return c.DataRequest(ctx, "contract_run", params)
}

func (c *Client) ListContracts(ctx context.Context) (interface{}, error) {
	return c.DataRequest(ctx, "contracts", nil)
}

func (c *Client) GetContract(ctx context.Context, contractID string) (interface{}, error) {
	if err := ValidateContractID(contractID); err != nil {
		return nil, fmt.Errorf("invalid contract id: %w", err)
	}

	params := map[string]interface{}{
		"contractid": contractID,
	}
	return c.DataRequest(ctx, "contract_status", params)
}
