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
			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}
			vault := cfg.KeyVaultName
			if vaultOverride != "" {
				vault = vaultOverride
			}
			prov := azcli.NewDefault()
			if err := prov.EnsureLoggedIn(context.Background()); err != nil {
				return fmt.Errorf("not logged in to Azure CLI: %w (run: yeet login)", err)
			}
			missing := make([]string, 0)
			secretsToCheck := make(map[string][]string) // secret name -> list of env vars that use it

			// Collect all unique secrets from all environments
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

			// Check each unique secret
			for secretName, envVars := range secretsToCheck {
				exists, err := prov.SecretExists(cmd.Context(), vault, secretName)
				if err != nil {
					return err
				}
				if !exists {
					for _, envVar := range envVars {
						missing = append(missing, fmt.Sprintf("%s -> %s", envVar, secretName))
					}
				}
			}
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
		},
	}
	return cmd
}
