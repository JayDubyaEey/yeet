package cli

import (
	"context"
	"fmt"
	"sort"

	"github.com/JayDubyaEey/yeet/internal/config"
	"github.com/JayDubyaEey/yeet/internal/provider/azcli"
	"github.com/JayDubyaEey/yeet/internal/ui"
	"github.com/spf13/cobra"
)

func newValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate config and check secrets exist in Key Vault",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidation(cmd.Context())
		},
	}
	return cmd
}

func runValidation(ctx context.Context) error {
	cfg, vault, prov, err := setupValidation()
	if err != nil {
		return err
	}

	secretsToCheck := collectSecretsToValidate(cfg)
	missing, err := checkSecretsExistence(ctx, prov, vault, secretsToCheck)
	if err != nil {
		return err
	}

	return reportValidationResults(missing, vault)
}

func setupValidation() (*config.Config, string, *azcli.Provider, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, "", nil, err
	}

	vault := cfg.KeyVaultName
	if vaultOverride != "" {
		vault = vaultOverride
	}

	prov := azcli.NewDefault()
	if err := prov.EnsureLoggedIn(context.Background()); err != nil {
		return nil, "", nil, fmt.Errorf("not logged in to Azure CLI: %w (run: yeet login)", err)
	}

	return cfg, vault, prov, nil
}

func collectSecretsToValidate(cfg *config.Config) map[string][]string {
	secretsToCheck := make(map[string][]string) // secret name -> list of env vars that use it

	for envKey, mapping := range cfg.Mappings {
		// Check local environment
		if localSpec := mapping.GetValueSpec(config.EnvLocal); localSpec != nil && localSpec.IsKeyvaultSecret() {
			secretsToCheck[localSpec.Value] = append(secretsToCheck[localSpec.Value], envKey+"(local)")
		}
		// Check docker environment
		if dockerSpec := mapping.GetValueSpec(config.EnvDocker); dockerSpec != nil && dockerSpec.IsKeyvaultSecret() {
			secretsToCheck[dockerSpec.Value] = append(secretsToCheck[dockerSpec.Value], envKey+"(docker)")
		}
	}

	return secretsToCheck
}

func checkSecretsExistence(ctx context.Context, prov *azcli.Provider, vault string, secretsToCheck map[string][]string) ([]string, error) {
	var missing []string

	for secretName, envVars := range secretsToCheck {
		exists, err := prov.SecretExists(ctx, vault, secretName)
		if err != nil {
			return nil, err
		}
		if !exists {
			for _, envVar := range envVars {
				missing = append(missing, fmt.Sprintf("%s -> %s", envVar, secretName))
			}
		}
	}

	return missing, nil
}

func reportValidationResults(missing []string, vault string) error {
	if len(missing) > 0 {
		sort.Strings(missing)
		ui.Error("missing %d secrets in vault %s:", len(missing), vault)
		for _, m := range missing {
			ui.Error("  - %s", m)
		}
		return fmt.Errorf("missing %d secrets", len(missing))
	}

	ui.Success("validation passed: all secrets exist in %s", vault)
	return nil
}
