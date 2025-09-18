package azcli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Provider implements secret operations using Azure CLI
type Provider struct {
	timeout time.Duration
}

// NewDefault creates a new Azure CLI provider with default settings
func NewDefault() *Provider {
	return &Provider{
		timeout: 30 * time.Second,
	}
}

// EnsureLoggedIn checks if the user is logged into Azure CLI
func (p *Provider) EnsureLoggedIn(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "az", "account", "show", "-o", "none")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("not logged in to Azure CLI")
	}
	return nil
}

// Login authenticates with Azure CLI
func (p *Provider) Login(ctx context.Context, tenant, subscription string) error {
	args := []string{"login", "-o", "none"}
	if tenant != "" {
		args = append(args, "--tenant", tenant)
	}
	
	cmd := exec.CommandContext(ctx, "az", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("az login failed: %w (stderr: %s)", err, stderr.String())
	}
	
	// Set subscription if provided
	if subscription != "" {
		cmd := exec.CommandContext(ctx, "az", "account", "set", "--subscription", subscription, "-o", "none")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to set subscription %s: %w", subscription, err)
		}
	}
	
	return nil
}

// Logout logs out from Azure CLI
func (p *Provider) Logout(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "az", "logout", "-o", "none")
	return cmd.Run()
}

// GetSecret retrieves a secret value from Key Vault
func (p *Provider) GetSecret(ctx context.Context, vault, name string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, p.timeout)
	defer cancel()
	
	cmd := exec.CommandContext(ctx, "az", "keyvault", "secret", "show",
		"--vault-name", vault,
		"--name", name,
		"-o", "json")
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		if strings.Contains(stderr.String(), "SecretNotFound") ||
			strings.Contains(stderr.String(), "(404)") {
			return "", &NotFoundError{Secret: name, Vault: vault}
		}
		return "", fmt.Errorf("failed to get secret: %w (stderr: %s)", err, stderr.String())
	}
	
	var result struct {
		Value string `json:"value"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		return "", fmt.Errorf("failed to parse secret response: %w", err)
	}
	
	return result.Value, nil
}

// SecretExists checks if a secret exists in Key Vault
func (p *Provider) SecretExists(ctx context.Context, vault, name string) (bool, error) {
	_, err := p.GetSecret(ctx, vault, name)
	if err != nil {
		if IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// WarmToken attempts to refresh the access token
func (p *Provider) WarmToken(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "az", "account", "get-access-token",
		"--resource", "https://vault.azure.net",
		"-o", "none")
	return cmd.Run()
}

// NotFoundError indicates a secret was not found
type NotFoundError struct {
	Secret string
	Vault  string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("secret %s not found in vault %s", e.Secret, e.Vault)
}

// IsNotFound checks if the error is a not found error
func IsNotFound(err error) bool {
	_, ok := err.(*NotFoundError)
	return ok
}
