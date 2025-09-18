package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/JayDubyaEey/yeet/internal/config"
	"github.com/JayDubyaEey/yeet/internal/provider/azcli"
	"github.com/JayDubyaEey/yeet/internal/ui"
	"github.com/spf13/cobra"
)

type listOptions struct {
	existsOnly  bool
	missingOnly bool
	raw         bool
}

type secretRow struct {
	Env    string `json:"env"`
	Secret string `json:"secret"`
	Exists bool   `json:"exists"`
}

func newListCmd() *cobra.Command {
	opts := &listOptions{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List env var mappings and existence status in Key Vault",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd.Context(), opts)
		},
	}
	cmd.Flags().BoolVar(&opts.existsOnly, "exists-only", false, "Show only secrets that exist")
	cmd.Flags().BoolVar(&opts.missingOnly, "missing-only", false, "Show only secrets that are missing")
	cmd.Flags().BoolVar(&opts.raw, "raw", false, "Output JSON for scripting")
	return cmd
}

func runList(ctx context.Context, opts *listOptions) error {
	cfg, vault, err := loadConfigAndVault()
	if err != nil {
		return err
	}

	prov := azcli.NewDefault()
	if err := prov.EnsureLoggedIn(context.Background()); err != nil {
		return fmt.Errorf("not logged in to Azure CLI: %w (run: yeet login)", err)
	}

	rows, err := fetchSecretStatuses(ctx, cfg, vault, prov)
	if err != nil {
		return err
	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].Env < rows[j].Env })

	if opts.raw {
		return outputJSON(rows)
	}

	return outputFormatted(rows, opts)
}

func loadConfigAndVault() (*config.Config, string, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, "", err
	}

	vault := cfg.KeyVaultName
	if vaultOverride != "" {
		vault = vaultOverride
	}

	return cfg, vault, nil
}

func fetchSecretStatuses(ctx context.Context, cfg *config.Config, vault string, prov *azcli.Provider) ([]secretRow, error) {
	rows := make([]secretRow, 0, len(cfg.Mappings))
	for envKey, mapping := range cfg.Mappings {
		exists, err := prov.SecretExists(ctx, vault, mapping.Secret)
		if err != nil {
			return nil, err
		}
		rows = append(rows, secretRow{Env: envKey, Secret: mapping.Secret, Exists: exists})
	}
	return rows, nil
}

func outputJSON(rows []secretRow) error {
	b, err := json.MarshalIndent(rows, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))
	return nil
}

func outputFormatted(rows []secretRow, opts *listOptions) error {
	for _, r := range rows {
		if shouldSkipRow(r, opts) {
			continue
		}
		printRow(r)
	}
	return nil
}

func shouldSkipRow(r secretRow, opts *listOptions) bool {
	return (opts.existsOnly && !r.Exists) || (opts.missingOnly && r.Exists)
}

func printRow(r secretRow) {
	status := "missing"
	if r.Exists {
		status = "exists"
	}

	if r.Exists {
		ui.Success("%s -> %s [%s]", r.Env, r.Secret, status)
	} else {
		ui.Warn("%s -> %s [%s]", r.Env, r.Secret, status)
	}
}
