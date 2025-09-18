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
			for envKey, mapping := range cfg.Mappings {
				exists, err := prov.SecretExists(cmd.Context(), vault, mapping.Secret)
				if err != nil {
					return err
				}
				if !exists {
					missing = append(missing, fmt.Sprintf("%s -> %s", envKey, mapping.Secret))
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
