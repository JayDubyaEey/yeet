package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"github.com/JayDubyaEey/yeet/internal/config"
	"github.com/JayDubyaEey/yeet/internal/ui"
)

type initOptions struct {
	envExamplePath string
	outputPath     string
	keyVaultName   string
	force          bool
}

type envVariable struct {
	name  string
	value string
}

var envVarPattern = regexp.MustCompile(`^([A-Z_][A-Z0-9_]*)\s*=\s*(.*)$`)

func newInitCmd() *cobra.Command {
	opts := &initOptions{}
	cmd := &cobra.Command{
		Use:   "init [key-vault-name]",
		Short: "Initialize yeet configuration from .env.example file",
		Long:  getInitLongDescription(),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.keyVaultName = args[0]
			}
			return runInit(cmd.Context(), opts)
		},
		Example: getInitExamples(),
	}

	setupInitFlags(cmd, opts)
	return cmd
}

func getInitLongDescription() string {
	return `Initialize yeet configuration by parsing a .env.example file and generating
an env.config.json with mappings for Azure Key Vault secrets.

The command will:
1. Parse your .env.example file
2. Generate secret names based on environment variable names
3. Create an env.config.json file with the mappings
4. Optionally use docker-specific overrides for local values`
}

func getInitExamples() string {
	return `  # Initialize with default key vault name
  yeet init

  # Initialize with specific key vault name
  yeet init my-key-vault

  # Use custom .env.example path
  yeet init my-key-vault --env-example .env.sample

  # Force overwrite existing config
  yeet init --force`
}

func setupInitFlags(cmd *cobra.Command, opts *initOptions) {
	cmd.Flags().StringVar(&opts.envExamplePath, "env-example", ".env.example", "Path to .env.example file")
	cmd.Flags().StringVar(&opts.outputPath, "output", "env.config.json", "Path for generated config file")
	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Force overwrite existing config file")
}

func runInit(ctx context.Context, opts *initOptions) error {
	if err := validateInitOptions(opts); err != nil {
		return err
	}

	mappings, err := generateMappings(ctx, opts)
	if err != nil {
		return err
	}

	if err := saveConfig(ctx, opts, mappings); err != nil {
		return err
	}

	printNextSteps(opts)
	return nil
}

func validateInitOptions(opts *initOptions) error {
	if opts.keyVaultName == "" {
		opts.keyVaultName = "my-key-vault"
		ui.Warn("No key vault name provided, using placeholder: %s", opts.keyVaultName)
		ui.Info("You can update this later in the config file")
	}

	if err := checkFileConflict(opts.outputPath, opts.force); err != nil {
		return err
	}

	if _, err := os.Stat(opts.envExamplePath); err != nil {
		return fmt.Errorf("env example file not found: %s", opts.envExamplePath)
	}

	return nil
}

func checkFileConflict(path string, force bool) error {
	if _, err := os.Stat(path); err == nil && !force {
		return fmt.Errorf("config file %s already exists (use --force to overwrite)", path)
	}
	return nil
}

func generateMappings(ctx context.Context, opts *initOptions) (map[string]config.Mapping, error) {
	ui.Info("Parsing %s...", opts.envExamplePath)

	mappings, err := parseEnvExample(opts.envExamplePath)
	if err != nil {
		return nil, err
	}

	if len(mappings) == 0 {
		return nil, fmt.Errorf("no environment variables found in %s", opts.envExamplePath)
	}

	ui.Success("Found %d environment variables", len(mappings))
	return mappings, nil
}

func saveConfig(ctx context.Context, opts *initOptions, mappings map[string]config.Mapping) error {
	cfg := &config.Config{
		KeyVaultName: opts.keyVaultName,
		Mappings:     mappings,
	}

	if err := writeConfig(cfg, opts.outputPath); err != nil {
		return err
	}

	ui.Success("Created %s with %d mappings", opts.outputPath, len(mappings))
	return nil
}

func printNextSteps(opts *initOptions) {
	ui.Info("\nNext steps:")
	ui.Info("1. Review and update %s with your actual Key Vault name", opts.outputPath)
	ui.Info("2. Update secret names to match your Key Vault secrets")
	ui.Info("3. Add docker-specific overrides where needed")
	ui.Info("4. Run 'yeet validate' to check your configuration")
	ui.Info("5. Run 'yeet fetch' to generate .env files")
}

func parseEnvExample(path string) (map[string]config.Mapping, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return scanEnvFile(file)
}

func scanEnvFile(file *os.File) (map[string]config.Mapping, error) {
	mappings := make(map[string]config.Mapping)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		if envVar := parseEnvLine(scanner.Text()); envVar != nil {
			mappings[envVar.name] = createMapping(envVar)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return mappings, nil
}

func parseEnvLine(line string) *envVariable {
	line = strings.TrimSpace(line)
	if shouldSkipLine(line) {
		return nil
	}

	matches := envVarPattern.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	return &envVariable{
		name:  matches[1],
		value: matches[2],
	}
}

func shouldSkipLine(line string) bool {
	return line == "" || strings.HasPrefix(line, "#")
}

func createMapping(envVar *envVariable) config.Mapping {
	mapping := config.Mapping{
		Secret: generateSecretName(envVar.name),
	}

	if looksLikeLocalValue(envVar.value) {
		mapping.Docker = &envVar.value
	}

	return mapping
}

func generateSecretName(envVar string) string {
	// Convert environment variable name to a Key Vault secret name
	// Key Vault secret names must be 1-127 characters, alphanumeric and hyphens only

	// Convert to lowercase and replace underscores with hyphens
	secretName := strings.ToLower(envVar)
	secretName = strings.ReplaceAll(secretName, "_", "-")

	// Remove any trailing hyphens
	secretName = strings.TrimSuffix(secretName, "-")

	return secretName
}

var (
	localPatterns = []string{
		"localhost", "127.0.0.1", "0.0.0.0", "host.docker.internal",
		"test", "development", "debug", "true", "false",
	}
	portPattern = regexp.MustCompile(`:\d+`)
)

func looksLikeLocalValue(value string) bool {
	lowerValue := strings.ToLower(value)

	return containsLocalPattern(lowerValue) ||
		containsPort(value) ||
		isTestValue(value)
}

func containsLocalPattern(value string) bool {
	for _, pattern := range localPatterns {
		if strings.Contains(value, pattern) {
			return true
		}
	}
	return false
}

func containsPort(value string) bool {
	return portPattern.MatchString(value)
}

func isTestValue(value string) bool {
	return strings.HasPrefix(value, "your-") || strings.HasPrefix(value, "sk_test_")
}

func writeConfig(cfg *config.Config, path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal config with pretty printing
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file with secure permissions
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
