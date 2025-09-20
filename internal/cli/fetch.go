package cli

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/JayDubyaEey/yeet/internal/config"
	"github.com/JayDubyaEey/yeet/internal/envwriter"
	"github.com/JayDubyaEey/yeet/internal/provider/azcli"
	"github.com/JayDubyaEey/yeet/internal/ui"
)

func newFetchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch",
		Short: "Fetch secrets and write .env and docker.env",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFetch(cmd.Context())
		},
	}
	return cmd
}

func newRefreshCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "refresh",
		Short: "Refresh .env and docker.env from Key Vault",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Same as fetch; optionally warm token
			if err := azcli.NewDefault().WarmToken(cmd.Context()); err != nil {
				ui.Warn("could not warm token: %v", err)
			}
			return runFetch(cmd.Context())
		},
	}
	return cmd
}

type fetchContext struct {
	cfg   *config.Config
	vault string
	prov  *azcli.Provider
}

type secretResult struct {
	key         string
	value       string
	environment config.Environment
}

func runFetch(ctx context.Context) error {
	fctx, err := prepareFetch()
	if err != nil {
		return err
	}

	if err := fctx.prov.EnsureLoggedIn(ctx); err != nil {
		return fmt.Errorf("not logged in to Azure CLI: %w (run: yeet login)", err)
	}

	results, missing, err := fetchSecrets(ctx, fctx)
	if err != nil {
		return err
	}

	if len(missing) > 0 {
		return reportMissingSecrets(missing, fctx.vault)
	}

	envMap, dockerMap := buildEnvMaps(results, fctx.cfg)
	return writeEnvFiles(envMap, dockerMap, fctx)
}

func prepareFetch() (*fetchContext, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, err
	}

	vault := cfg.KeyVaultName
	if vaultOverride != "" {
		vault = vaultOverride
	}

	return &fetchContext{
		cfg:   cfg,
		vault: vault,
		prov:  azcli.NewDefault(),
	}, nil
}

func fetchSecrets(ctx context.Context, fctx *fetchContext) ([]secretResult, []string, error) {
	// We need to fetch secrets for both environments
	localSecrets := make(map[string]string) // secret name -> value cache
	results := make([]secretResult, 0)
	missing := make([]string, 0)

	g, gctx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, 6)
	var mu sync.Mutex

	// Collect all unique Key Vault secrets we need to fetch
	secretsToFetch := make(map[string]bool)
	for _, mapping := range fctx.cfg.Mappings {
		// Check local environment
		if localSpec := mapping.GetValueSpec(config.EnvLocal); localSpec != nil && localSpec.IsKeyvaultSecret() {
			secretsToFetch[localSpec.Value] = true
		}
		// Check docker environment
		if dockerSpec := mapping.GetValueSpec(config.EnvDocker); dockerSpec != nil && dockerSpec.IsKeyvaultSecret() {
			secretsToFetch[dockerSpec.Value] = true
		}
	}

	// Fetch all required secrets
	for secretName := range secretsToFetch {
		secretName := secretName
		sem <- struct{}{}
		g.Go(func() error {
			defer func() { <-sem }()
			return fetchKeyVaultSecret(gctx, fctx, secretName, localSecrets, &missing, &mu)
		})
	}

	if err := g.Wait(); err != nil {
		return nil, nil, err
	}

	// Now build results for each environment variable
	for envKey, mapping := range fctx.cfg.Mappings {
		// Process local environment
		if localSpec := mapping.GetValueSpec(config.EnvLocal); localSpec != nil {
			result := secretResult{
				key:         envKey,
				environment: config.EnvLocal,
			}
			if localSpec.IsKeyvaultSecret() {
				if val, exists := localSecrets[localSpec.Value]; exists {
					result.value = val
				} else {
					missing = append(missing, fmt.Sprintf("%s (local) -> %s", envKey, localSpec.Value))
					continue
				}
			} else {
				result.value = localSpec.Value
			}
			results = append(results, result)
		}

		// Process docker environment
		if dockerSpec := mapping.GetValueSpec(config.EnvDocker); dockerSpec != nil {
			result := secretResult{
				key:         envKey,
				environment: config.EnvDocker,
			}
			if dockerSpec.IsKeyvaultSecret() {
				if val, exists := localSecrets[dockerSpec.Value]; exists {
					result.value = val
				} else {
					missing = append(missing, fmt.Sprintf("%s (docker) -> %s", envKey, dockerSpec.Value))
					continue
				}
			} else {
				result.value = dockerSpec.Value
			}
			results = append(results, result)
		}
	}

	return results, missing, nil
}

func fetchKeyVaultSecret(ctx context.Context, fctx *fetchContext, secretName string, cache map[string]string, missing *[]string, mu *sync.Mutex) error {
	val, err := fctx.prov.GetSecret(ctx, fctx.vault, secretName)
	if err != nil {
		if azcli.IsNotFound(err) {
			mu.Lock()
			*missing = append(*missing, fmt.Sprintf("secret: %s", secretName))
			mu.Unlock()
			return nil
		}
		return fmt.Errorf("failed to get secret %s: %w", secretName, err)
	}
	mu.Lock()
	cache[secretName] = val
	mu.Unlock()
	return nil
}

func reportMissingSecrets(missing []string, vault string) error {
	ui.Error("missing %d secrets in vault %s:", len(missing), vault)
	sort.Strings(missing)
	for _, m := range missing {
		ui.Error("  - %s", m)
	}
	return errors.New("one or more secrets are missing")
}

func buildEnvMaps(results []secretResult, cfg *config.Config) (map[string]string, map[string]string) {
	envMap := make(map[string]string)    // local environment
	dockerMap := make(map[string]string) // docker environment

	// Group results by environment
	for _, r := range results {
		switch r.environment {
		case config.EnvLocal:
			envMap[r.key] = r.value
		case config.EnvDocker:
			dockerMap[r.key] = r.value
		}
	}

	// Handle mappings that don't have environment-specific values
	// (fallback to global values or legacy format)
	for envKey, mapping := range cfg.Mappings {
		// For local environment
		if _, exists := envMap[envKey]; !exists {
			if spec := mapping.GetValueSpec(config.EnvLocal); spec != nil {
				if spec.IsLiteral() {
					envMap[envKey] = spec.Value
				}
				// Keyvault values should already be in results
			}
		}

		// For docker environment
		if _, exists := dockerMap[envKey]; !exists {
			if spec := mapping.GetValueSpec(config.EnvDocker); spec != nil {
				if spec.IsLiteral() {
					dockerMap[envKey] = spec.Value
				}
				// Keyvault values should already be in results
			} else {
				// If no docker-specific value, fall back to local
				if localVal, hasLocal := envMap[envKey]; hasLocal {
					dockerMap[envKey] = localVal
				}
			}
		}
	}

	return envMap, dockerMap
}

func writeEnvFiles(envMap, dockerMap map[string]string, fctx *fetchContext) error {
	existingEnv, _ := envwriter.ReadKeyValues(".env")
	existingDocker, _ := envwriter.ReadKeyValues("docker.env")

	finalEnv := envwriter.MergeRetainUnknowns(envMap, existingEnv, fctx.cfg.Mappings)
	finalDocker := envwriter.MergeRetainUnknowns(dockerMap, existingDocker, fctx.cfg.Mappings)

	warnUnmappedKeys(existingEnv, existingDocker, fctx.cfg.Mappings)

	header := fmt.Sprintf("# Generated by yeet\n# Source: %s\n# Vault: %s\n# Generated: %s\n",
		configPath, fctx.vault, time.Now().Format(time.RFC3339))

	if err := envwriter.WriteEnvFile(".env", finalEnv, header); err != nil {
		return err
	}
	if err := envwriter.WriteEnvFile("docker.env", finalDocker, header); err != nil {
		return err
	}

	ui.Success("wrote .env and docker.env (%d keys)", len(finalEnv))
	return nil
}

func warnUnmappedKeys(existingEnv, existingDocker map[string]string, mappings map[string]config.Mapping) {
	unmappedEnv := envwriter.UnmappedKeys(existingEnv, mappings)
	unmappedDocker := envwriter.UnmappedKeys(existingDocker, mappings)

	if len(unmappedEnv) > 0 || len(unmappedDocker) > 0 {
		ui.Warn("retaining %d keys not defined in env.config.json", len(unmappedEnv)+len(unmappedDocker))
		for _, k := range unmappedEnv {
			ui.Warn("  - .env: %s", k)
		}
		for _, k := range unmappedDocker {
			ui.Warn("  - docker.env: %s", k)
		}
	}
}
