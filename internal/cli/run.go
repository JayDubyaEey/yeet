package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/JayDubyaEey/yeet/internal/config"
	"github.com/JayDubyaEey/yeet/internal/provider/azcli"
	"github.com/JayDubyaEey/yeet/internal/ui"
)

var (
	loadEnvFile bool
	envFilePath string
)

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run [command...]",
		Short: "Run a command with secrets from Key Vault as environment variables",
		Long: `Run a command with secrets from Key Vault as environment variables.

Examples:
  yeet run make dev
  yeet run --vault my-vault npm start
  yeet run -v production-vault -- docker-compose up
  yeet run --load-env -- npm start  # Load .env file for overrides`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWithSecrets(cmd.Context(), args)
		},
	}

	cmd.Flags().BoolVarP(&loadEnvFile, "load-env", "e", false, "Load .env file for local overrides")
	cmd.Flags().StringVar(&envFilePath, "env-file", ".env", "Path to env file to load (only used with --load-env)")

	return cmd
}

func runWithSecrets(ctx context.Context, args []string) error {
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	vault := cfg.KeyVaultName
	if vaultOverride != "" {
		vault = vaultOverride
	}

	// Ensure logged in
	prov := azcli.NewDefault()
	if err := prov.EnsureLoggedIn(ctx); err != nil {
		return fmt.Errorf("not logged in to Azure CLI: %w (run: yeet login)", err)
	}

	ui.Info("fetching secrets from vault: %s", vault)

	// Fetch secrets concurrently
	envVars, err := fetchSecretsAsEnv(ctx, cfg, vault, prov)
	if err != nil {
		return err
	}

	ui.Success("loaded %d environment variables from Key Vault", len(envVars))

	// Load local .env overrides if requested
	if loadEnvFile {
		overrides, err := loadEnvOverrides(envFilePath)
		if err != nil {
			ui.Warn("could not load env file %s: %v", envFilePath, err)
		} else {
			// Apply overrides
			for key, value := range overrides {
				if _, exists := envVars[key]; exists {
					ui.Info("overriding %s from %s", key, envFilePath)
				}
				envVars[key] = value
			}
			ui.Success("loaded %d overrides from %s", len(overrides), envFilePath)
		}
	}

	// Prepare command
	cmdName := args[0]
	cmdArgs := args[1:]

	ui.Info("running: %s %s", cmdName, strings.Join(cmdArgs, " "))

	// Create command
	cmd := exec.CommandContext(ctx, cmdName, cmdArgs...)

	// Set up environment
	cmd.Env = os.Environ() // Start with current environment
	for key, value := range envVars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
	}

	// Connect stdio
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		// Try to get the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
				os.Exit(status.ExitStatus())
			}
		}
		return fmt.Errorf("command failed: %w", err)
	}

	return nil
}

func fetchSecretsAsEnv(ctx context.Context, cfg *config.Config, vault string, prov *azcli.Provider) (map[string]string, error) {
	envVars := make(map[string]string)
	missing := make([]string, 0)

	g, gctx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, 6) // concurrency limit
	var mu sync.Mutex

	for envKey, mapping := range cfg.Mappings {
		envKey := envKey
		mapping := mapping
		sem <- struct{}{}
		g.Go(func() error {
			defer func() { <-sem }()

			val, err := prov.GetSecret(gctx, vault, mapping.Secret)
			if err != nil {
				if azcli.IsNotFound(err) {
					mu.Lock()
					missing = append(missing, fmt.Sprintf("%s -> %s", envKey, mapping.Secret))
					mu.Unlock()
					return nil
				}
				return fmt.Errorf("failed to get secret %s for %s: %w", mapping.Secret, envKey, err)
			}

			mu.Lock()
			envVars[envKey] = val
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	if len(missing) > 0 {
		ui.Error("missing %d secrets in vault %s:", len(missing), vault)
		for _, m := range missing {
			ui.Error("  - %s", m)
		}
		return nil, fmt.Errorf("one or more secrets are missing")
	}

	return envVars, nil
}

func loadEnvOverrides(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	overrides := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
			// Unescape basic sequences
			value = strings.ReplaceAll(value, "\\\"", "\"")
			value = strings.ReplaceAll(value, "\\n", "\n")
			value = strings.ReplaceAll(value, "\\r", "\r")
			value = strings.ReplaceAll(value, "\\\\", "\\")
		}

		overrides[key] = value
	}

	return overrides, scanner.Err()
}
