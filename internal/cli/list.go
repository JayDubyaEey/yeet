package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/spf13/cobra"
	"github.com/JayDubyaEey/yeet/internal/config"
	"github.com/JayDubyaEey/yeet/internal/provider/azcli"
	"github.com/JayDubyaEey/yeet/internal/ui"
)

func newListCmd() *cobra.Command {
	var existsOnly bool
	var missingOnly bool
	var raw bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List env var mappings and existence status in Key Vault",
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
			type row struct { Env string `json:"env"`; Secret string `json:"secret"`; Exists bool `json:"exists"` }
			rows := make([]row, 0, len(cfg.Mappings))
			for envKey, mapping := range cfg.Mappings {
				exists, err := prov.SecretExists(cmd.Context(), vault, mapping.Secret)
				if err != nil { return err }
				rows = append(rows, row{Env: envKey, Secret: mapping.Secret, Exists: exists})
			}
			sort.Slice(rows, func(i, j int) bool { return rows[i].Env < rows[j].Env })
			if raw {
				b, _ := json.MarshalIndent(rows, "", "  ")
				fmt.Println(string(b))
				return nil
			}
			for _, r := range rows {
				if existsOnly && !r.Exists { continue }
				if missingOnly && r.Exists { continue }
				status := "missing"
				if r.Exists { status = "exists" }
				if r.Exists { ui.Success("%s -> %s [%s]", r.Env, r.Secret, status) } else { ui.Warn("%s -> %s [%s]", r.Env, r.Secret, status) }
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&existsOnly, "exists-only", false, "Show only secrets that exist")
	cmd.Flags().BoolVar(&missingOnly, "missing-only", false, "Show only secrets that are missing")
	cmd.Flags().BoolVar(&raw, "raw", false, "Output JSON for scripting")
	return cmd
}

