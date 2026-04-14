package main

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type TenzroCLI struct {
	binPath string
	rpcURL  string
}

func NewTenzroCLI(binPath, rpcURL string) *TenzroCLI {
	if binPath == "" {
		binPath = "tenzro-cli"
	}
	return &TenzroCLI{binPath: binPath, rpcURL: rpcURL}
}

func (t *TenzroCLI) run(ctx context.Context, args ...string) (string, error) {
	if t.rpcURL != "" {
		args = append(args, "--rpc", t.rpcURL)
	}
	cmd := exec.CommandContext(ctx, t.binPath, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("tenzro-cli %s: %w\noutput: %s", strings.Join(args, " "), err, string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

func (t *TenzroCLI) Join(ctx context.Context, name string) (string, error) {
	return t.run(ctx, "join", "--name", name)
}

func (t *TenzroCLI) WalletCreate(ctx context.Context, walletName string) (string, error) {
	return t.run(ctx, "wallet", "create", "--name", walletName)
}

func (t *TenzroCLI) WalletBalance(ctx context.Context, address string) (string, error) {
	args := []string{"wallet", "balance"}
	if address != "" {
		args = append(args, "--address", address)
	}
	return t.run(ctx, args...)
}

func (t *TenzroCLI) IdentityRegister(ctx context.Context, name, identityType string) (string, error) {
	return t.run(ctx, "identity", "register", "--name", name, "--identity-type", identityType)
}

func (t *TenzroCLI) IdentityResolve(ctx context.Context, did string) (string, error) {
	return t.run(ctx, "identity", "resolve", did)
}

func (t *TenzroCLI) SetUsername(ctx context.Context, username string) (string, error) {
	return t.run(ctx, "set-username", username)
}

func (t *TenzroCLI) Faucet(ctx context.Context, address string) (string, error) {
	return t.run(ctx, "faucet", address)
}

func (t *TenzroCLI) Info(ctx context.Context) (string, error) {
	return t.run(ctx, "info")
}

func (t *TenzroCLI) ModelList(ctx context.Context) (string, error) {
	return t.run(ctx, "model", "list", "--serving")
}

func (t *TenzroCLI) Stake(ctx context.Context, amount, providerType string) (string, error) {
	return t.run(ctx, "stake", "deposit", amount, "--provider-type", providerType)
}

func (t *TenzroCLI) BridgeQuote(ctx context.Context, fromChain, toChain, token string, amount string) (string, error) {
	return t.run(ctx, "bridge", "quote",
		"--from-chain", fromChain,
		"--to-chain", toChain,
		"--token", token,
		"--amount", amount)
}

func (t *TenzroCLI) BridgeExecute(ctx context.Context, fromChain, toChain, token, amount, sender, recipient string) (string, error) {
	return t.run(ctx, "bridge", "execute",
		"--from-chain", fromChain,
		"--to-chain", toChain,
		"--token", token,
		"--amount", amount,
		"--sender", sender,
		"--recipient", recipient)
}

func (t *TenzroCLI) RegisterTool(ctx context.Context, name, description, endpoint, toolType, category, version, creatorDID string) (string, error) {
	return t.run(ctx, "tool", "register",
		"--name", name,
		"--description", description,
		"--endpoint", endpoint,
		"--tool-type", toolType,
		"--category", category,
		"--version", version,
		"--creator-did", creatorDID)
}
