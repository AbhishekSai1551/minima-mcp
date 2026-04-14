package mcp

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/minima-global/minima-mcp/internal/minima"
)

func object(props map[string]interface{}, required []string) map[string]interface{} {
	schema := map[string]interface{}{
		"type":       "object",
		"properties": props,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func strProp(desc string) map[string]interface{} {
	return map[string]interface{}{"type": "string", "description": desc}
}

func intProp(desc string) map[string]interface{} {
	return map[string]interface{}{"type": "integer", "description": desc}
}

func strArrayProp(desc string) map[string]interface{} {
	return map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": desc}
}

func RegisterNodeTools(s *MCPServer, c *minima.Client, l *slog.Logger) {
	s.RegisterTool(&Tool{
		Name:        "minima_get_status",
		Description: "Get the current status of the Minima node including sync state, block height, peer count, and version",
		InputSchema: object(map[string]interface{}{}, nil),
		AuthLevel:   "public",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return c.GetStatus(ctx)
		},
	})

	s.RegisterTool(&Tool{
		Name:        "minima_get_block",
		Description: "Get block information by height from the Minima blockchain",
		InputSchema: object(map[string]interface{}{
			"height": intProp("Block height number"),
		}, []string{"height"}),
		AuthLevel: "public",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			height, ok := args["height"].(float64)
			if !ok {
				return nil, fmt.Errorf("height must be a number")
			}
			return c.GetBlock(ctx, int64(height))
		},
	})

	s.RegisterTool(&Tool{
		Name:        "minima_get_transaction",
		Description: "Get transaction details by transaction ID",
		InputSchema: object(map[string]interface{}{
			"txid": strProp("Transaction ID (hash)"),
		}, []string{"txid"}),
		AuthLevel: "public",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			txid, _ := args["txid"].(string)
			return c.GetTransaction(ctx, txid)
		},
	})

	s.RegisterTool(&Tool{
		Name:        "minima_get_network_info",
		Description: "Get network information including peers and connection status",
		InputSchema: object(map[string]interface{}{}, nil),
		AuthLevel:   "public",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return c.GetNetworkInfo(ctx)
		},
	})
}

func RegisterWalletTools(s *MCPServer, c *minima.Client, l *slog.Logger) {
	s.RegisterTool(&Tool{
		Name:        "minima_get_balance",
		Description: "Get the balance of a Minima wallet. Returns confirmed and unconfirmed balances.",
		InputSchema: object(map[string]interface{}{
			"address": strProp("Optional address to check balance for. Defaults to current node wallet."),
		}, nil),
		AuthLevel: "public",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			addr, _ := args["address"].(string)
			return c.GetBalance(ctx, addr)
		},
	})

	s.RegisterTool(&Tool{
		Name:        "minima_send_transaction",
		Description: "Send a transaction on the Minima blockchain. Requires recipient address and amount.",
		InputSchema: object(map[string]interface{}{
			"to":     strProp("Recipient address (Mx... or 0x...)"),
			"amount": strProp("Amount to send (e.g. \"10\")"),
			"asset":  strProp("Optional asset/token ID to send"),
		}, []string{"to", "amount"}),
		AuthLevel: "authenticated",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			to, _ := args["to"].(string)
			amount, _ := args["amount"].(string)
			asset, _ := args["asset"].(string)
			return c.SendTransaction(ctx, to, amount, asset)
		},
	})

	s.RegisterTool(&Tool{
		Name:        "minima_get_transactions",
		Description: "Get transaction history for the current wallet",
		InputSchema: object(map[string]interface{}{
			"offset": intProp("Offset for pagination (default 0)"),
			"limit":  intProp("Maximum number of transactions to return (1-100, default 50)"),
		}, nil),
		AuthLevel: "authenticated",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			offset := 0
			limit := 50
			if v, ok := args["offset"].(float64); ok {
				offset = int(v)
			}
			if v, ok := args["limit"].(float64); ok {
				limit = int(v)
			}
			return c.GetTransactions(ctx, offset, limit)
		},
	})

	s.RegisterTool(&Tool{
		Name:        "minima_get_coins",
		Description: "List all coins/tokens held by the current wallet",
		InputSchema: object(map[string]interface{}{}, nil),
		AuthLevel:   "authenticated",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return c.GetCoins(ctx)
		},
	})
}

func RegisterContractTools(s *MCPServer, c *minima.Client, l *slog.Logger) {
	s.RegisterTool(&Tool{
		Name:        "minima_deploy_contract",
		Description: "Deploy a new smart contract on the Minima blockchain. The script is written in Minima's Kotlin-based smart contract language.",
		InputSchema: object(map[string]interface{}{
			"script": strProp("Smart contract script (Kotlin/Minima DSL)"),
		}, []string{"script"}),
		AuthLevel: "authenticated",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			script, _ := args["script"].(string)
			stateParams := make(map[string]string)
			if sp, ok := args["state_params"].(map[string]interface{}); ok {
				for k, v := range sp {
					stateParams[k] = fmt.Sprintf("%v", v)
				}
			}
			return c.DeployContract(ctx, script, stateParams)
		},
	})

	s.RegisterTool(&Tool{
		Name:        "minima_execute_contract",
		Description: "Execute a function on a deployed Minima smart contract",
		InputSchema: object(map[string]interface{}{
			"contract_id":   strProp("The contract ID to execute"),
			"function_name": strProp("Name of the function to call"),
		}, []string{"contract_id", "function_name"}),
		AuthLevel: "authenticated",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			contractID, _ := args["contract_id"].(string)
			functionName, _ := args["function_name"].(string)
			funcArgs := make(map[string]string)
			if a, ok := args["args"].(map[string]interface{}); ok {
				for k, v := range a {
					funcArgs[k] = fmt.Sprintf("%v", v)
				}
			}
			return c.ExecuteContract(ctx, contractID, functionName, funcArgs)
		},
	})

	s.RegisterTool(&Tool{
		Name:        "minima_list_contracts",
		Description: "List all deployed smart contracts on the node",
		InputSchema: object(map[string]interface{}{}, nil),
		AuthLevel:   "authenticated",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return c.ListContracts(ctx)
		},
	})

	s.RegisterTool(&Tool{
		Name:        "minima_get_contract",
		Description: "Get the status and state of a deployed smart contract",
		InputSchema: object(map[string]interface{}{
			"contract_id": strProp("The contract ID to query"),
		}, []string{"contract_id"}),
		AuthLevel: "authenticated",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			contractID, _ := args["contract_id"].(string)
			return c.GetContract(ctx, contractID)
		},
	})
}

func RegisterTokenTools(s *MCPServer, c *minima.Client, l *slog.Logger) {
	s.RegisterTool(&Tool{
		Name:        "minima_create_token",
		Description: "Create a new custom token on the Minima blockchain",
		InputSchema: object(map[string]interface{}{
			"name":     strProp("Token name"),
			"amount":   strProp("Initial supply amount"),
			"decimals": intProp("Number of decimal places (default 0)"),
		}, []string{"name", "amount"}),
		AuthLevel: "authenticated",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			name, _ := args["name"].(string)
			amount, _ := args["amount"].(string)
			decimals := 0
			if v, ok := args["decimals"].(float64); ok {
				decimals = int(v)
			}
			return c.CreateToken(ctx, name, amount, decimals)
		},
	})

	s.RegisterTool(&Tool{
		Name:        "minima_list_tokens",
		Description: "List all tokens on the Minima node",
		InputSchema: object(map[string]interface{}{}, nil),
		AuthLevel:   "public",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return c.ListTokens(ctx)
		},
	})

	s.RegisterTool(&Tool{
		Name:        "minima_transfer_token",
		Description: "Transfer a custom token to another address",
		InputSchema: object(map[string]interface{}{
			"token_id": strProp("Token ID to transfer"),
			"to":       strProp("Recipient address"),
			"amount":   strProp("Amount to transfer"),
		}, []string{"token_id", "to", "amount"}),
		AuthLevel: "authenticated",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			tokenID, _ := args["token_id"].(string)
			to, _ := args["to"].(string)
			amount, _ := args["amount"].(string)
			return c.TransferToken(ctx, tokenID, to, amount)
		},
	})

	s.RegisterTool(&Tool{
		Name:        "minima_get_token_info",
		Description: "Get information about a specific token",
		InputSchema: object(map[string]interface{}{
			"token_id": strProp("Token ID to look up"),
		}, []string{"token_id"}),
		AuthLevel: "public",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			tokenID, _ := args["token_id"].(string)
			return c.GetTokenInfo(ctx, tokenID)
		},
	})
}

func RegisterKeyTools(s *MCPServer, c *minima.Client, l *slog.Logger) {
	s.RegisterTool(&Tool{
		Name:        "minima_generate_key",
		Description: "Generate a new key pair on the Minima node",
		InputSchema: object(map[string]interface{}{}, nil),
		AuthLevel:   "authenticated",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return c.GenerateKey(ctx)
		},
	})

	s.RegisterTool(&Tool{
		Name:        "minima_list_keys",
		Description: "List all public keys in the vault",
		InputSchema: object(map[string]interface{}{}, nil),
		AuthLevel:   "authenticated",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return c.ListKeys(ctx)
		},
	})

	s.RegisterTool(&Tool{
		Name:        "minima_sign_message",
		Description: "Sign a message with a private key from the vault",
		InputSchema: object(map[string]interface{}{
			"private_key": strProp("Private key to sign with"),
			"message":     strProp("Message to sign"),
		}, []string{"private_key", "message"}),
		AuthLevel: "authenticated",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			privateKey, _ := args["private_key"].(string)
			message, _ := args["message"].(string)
			return c.SignMessage(ctx, privateKey, message)
		},
	})

	s.RegisterTool(&Tool{
		Name:        "minima_verify_signature",
		Description: "Verify a cryptographic signature against a public key and message",
		InputSchema: object(map[string]interface{}{
			"public_key": strProp("Public key that produced the signature"),
			"message":    strProp("Original message that was signed"),
			"signature":  strProp("The signature to verify"),
		}, []string{"public_key", "message", "signature"}),
		AuthLevel: "public",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			publicKey, _ := args["public_key"].(string)
			message, _ := args["message"].(string)
			signature, _ := args["signature"].(string)
			return c.VerifySignature(ctx, publicKey, message, signature)
		},
	})

	s.RegisterTool(&Tool{
		Name:        "minima_get_vault",
		Description: "Get vault information from the Minima node",
		InputSchema: object(map[string]interface{}{}, nil),
		AuthLevel:   "authenticated",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return c.GetVault(ctx)
		},
	})
}

func RegisterMiniDAPPTools(s *MCPServer, c *minima.Client, l *slog.Logger) {
	s.RegisterTool(&Tool{
		Name:        "minima_list_minidapps",
		Description: "List all installed MiniDAPPs on the node",
		InputSchema: object(map[string]interface{}{}, nil),
		AuthLevel:   "authenticated",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return c.ListMiniDAPPs(ctx)
		},
	})

	s.RegisterTool(&Tool{
		Name:        "minima_install_minidapp",
		Description: "Install a MiniDAPP on the node",
		InputSchema: object(map[string]interface{}{
			"app_uid": strProp("MiniDAPP unique identifier"),
			"version": strProp("Version to install"),
		}, []string{"app_uid"}),
		AuthLevel: "authenticated",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			appUID, _ := args["app_uid"].(string)
			version, _ := args["version"].(string)
			return c.InstallMiniDAPP(ctx, appUID, version)
		},
	})

	s.RegisterTool(&Tool{
		Name:        "minima_uninstall_minidapp",
		Description: "Uninstall a MiniDAPP from the node",
		InputSchema: object(map[string]interface{}{
			"app_uid": strProp("MiniDAPP unique identifier"),
		}, []string{"app_uid"}),
		AuthLevel: "authenticated",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			appUID, _ := args["app_uid"].(string)
			return c.UninstallMiniDAPP(ctx, appUID)
		},
	})

	s.RegisterTool(&Tool{
		Name:        "minima_get_minidapp_info",
		Description: "Get information about a specific MiniDAPP",
		InputSchema: object(map[string]interface{}{
			"app_uid": strProp("MiniDAPP unique identifier"),
		}, []string{"app_uid"}),
		AuthLevel: "public",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			appUID, _ := args["app_uid"].(string)
			return c.GetMiniDAPPInfo(ctx, appUID)
		},
	})

	s.RegisterTool(&Tool{
		Name:        "minima_run_minidapp",
		Description: "Execute a command on a MiniDAPP",
		InputSchema: object(map[string]interface{}{
			"app_uid": strProp("MiniDAPP unique identifier"),
			"command": strProp("Command to execute"),
			"args":    strArrayProp("Command arguments"),
		}, []string{"app_uid", "command"}),
		AuthLevel: "authenticated",
		Handler: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			appUID, _ := args["app_uid"].(string)
			command, _ := args["command"].(string)
			runArgs := make(map[string]string)
			if a, ok := args["args"].([]interface{}); ok {
				for i, v := range a {
					runArgs[fmt.Sprintf("arg%d", i)] = fmt.Sprintf("%v", v)
				}
			}
			return c.RunMiniDAPP(ctx, appUID, command, runArgs)
		},
	})
}
