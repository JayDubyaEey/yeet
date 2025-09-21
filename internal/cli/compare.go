package cli

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/JayDubyaEey/yeet/internal/config"
	"github.com/JayDubyaEey/yeet/internal/ui"
)

var (
	deploymentPath string
)

func newCompareCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compare",
		Short: "Compare configuration environment variables with Kubernetes deployment",
		Long: `Compare the environment variables defined in your configuration 
with those used in your Kubernetes deployment file.

This helps identify:
- Variables in config but not used in deployment
- Variables in deployment but not defined in config
- Potential mismatches or unused configurations`,
		Example: `  yeet compare
  yeet compare --deployment deploy/prod/deployment.yml
  yeet compare -d k8s/deployment.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCompare()
		},
	}

	cmd.Flags().StringVarP(&deploymentPath, "deployment", "d", "deploy/manifests/base/deployment.yaml",
		"Path to Kubernetes deployment YAML file")

	return cmd
}

// KubernetesDeployment represents the structure we care about in a K8s deployment
type KubernetesDeployment struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Spec       struct {
		Template struct {
			Spec struct {
				Containers []Container `yaml:"containers"`
			} `yaml:"spec"`
		} `yaml:"template"`
	} `yaml:"spec"`
}

// Container represents a container in a K8s deployment
type Container struct {
	Name string   `yaml:"name"`
	Env  []EnvVar `yaml:"env,omitempty"`
}

// EnvVar represents an environment variable in a K8s container
type EnvVar struct {
	Name      string             `yaml:"name"`
	Value     string             `yaml:"value,omitempty"`
	ValueFrom *EnvVarValueSource `yaml:"valueFrom,omitempty"`
}

// EnvVarValueSource represents the source of an environment variable value
type EnvVarValueSource struct {
	SecretKeyRef    *SecretKeyRef    `yaml:"secretKeyRef,omitempty"`
	ConfigMapKeyRef *ConfigMapKeyRef `yaml:"configMapKeyRef,omitempty"`
}

// SecretKeyRef represents a reference to a secret key
type SecretKeyRef struct {
	Name string `yaml:"name"`
	Key  string `yaml:"key"`
}

// ConfigMapKeyRef represents a reference to a config map key
type ConfigMapKeyRef struct {
	Name string `yaml:"name"`
	Key  string `yaml:"key"`
}

// ComparisonResult holds the result of comparing config vs deployment
type ComparisonResult struct {
	ConfigVars       []string // Variables defined in config
	DeploymentVars   []string // Variables defined in deployment
	InConfigOnly     []string // Variables in config but not in deployment
	InDeploymentOnly []string // Variables in deployment but not in config
	Matching         []string // Variables in both config and deployment
}

func runCompare() error {
	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Check if deployment file exists
	if _, err := os.Stat(deploymentPath); os.IsNotExist(err) {
		return fmt.Errorf("deployment file not found: %s", deploymentPath)
	}

	// Parse deployment file
	deploymentVars, err := extractEnvVarsFromDeployment(deploymentPath)
	if err != nil {
		return fmt.Errorf("failed to parse deployment file: %w", err)
	}

	// Extract config variables
	configVars := extractConfigVars(cfg)

	// Compare and generate result
	result := compareVars(configVars, deploymentVars)

	// Display results
	displayComparisonResult(result, deploymentPath)

	return nil
}

func extractEnvVarsFromDeployment(filePath string) ([]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read deployment file: %w", err)
	}

	var envVars []string
	envVarSet := make(map[string]bool) // Use set to avoid duplicates

	// Split YAML documents (in case there are multiple documents in one file)
	documents := strings.Split(string(data), "---")

	for _, doc := range documents {
		doc = strings.TrimSpace(doc)
		if doc == "" {
			continue
		}

		var deployment KubernetesDeployment
		if err := yaml.Unmarshal([]byte(doc), &deployment); err != nil {
			// Skip non-deployment documents or malformed YAML
			continue
		}

		// Only process Deployment resources
		if deployment.Kind != "Deployment" {
			continue
		}

		// Extract environment variables from all containers
		for _, container := range deployment.Spec.Template.Spec.Containers {
			for _, env := range container.Env {
				if !envVarSet[env.Name] {
					envVars = append(envVars, env.Name)
					envVarSet[env.Name] = true
				}
			}
		}
	}

	sort.Strings(envVars)
	return envVars, nil
}

func extractConfigVars(cfg *config.Config) []string {
	var vars []string
	for varName := range cfg.Mappings {
		vars = append(vars, varName)
	}
	sort.Strings(vars)
	return vars
}

func compareVars(configVars, deploymentVars []string) ComparisonResult {
	configSet := make(map[string]bool)
	deploymentSet := make(map[string]bool)

	// Create sets for faster lookup
	for _, v := range configVars {
		configSet[v] = true
	}
	for _, v := range deploymentVars {
		deploymentSet[v] = true
	}

	var inConfigOnly, inDeploymentOnly, matching []string

	// Find variables in config but not in deployment
	for _, v := range configVars {
		if !deploymentSet[v] {
			inConfigOnly = append(inConfigOnly, v)
		} else {
			matching = append(matching, v)
		}
	}

	// Find variables in deployment but not in config
	for _, v := range deploymentVars {
		if !configSet[v] {
			inDeploymentOnly = append(inDeploymentOnly, v)
		}
	}

	return ComparisonResult{
		ConfigVars:       configVars,
		DeploymentVars:   deploymentVars,
		InConfigOnly:     inConfigOnly,
		InDeploymentOnly: inDeploymentOnly,
		Matching:         matching,
	}
}

func displayComparisonResult(result ComparisonResult, deploymentFile string) {
	ui.Info("🔍 Comparing configuration with deployment: %s", deploymentFile)
	fmt.Println()

	displaySummary(result)
	displayMatchingVariables(result.Matching)
	displayConfigOnlyVariables(result.InConfigOnly)
	displayDeploymentOnlyVariables(result.InDeploymentOnly)
	displayOverallStatus(result)
}

func displaySummary(result ComparisonResult) {
	ui.Info("📊 Summary:")
	fmt.Printf("  • Configuration variables: %d\n", len(result.ConfigVars))
	fmt.Printf("  • Deployment variables:    %d\n", len(result.DeploymentVars))
	fmt.Printf("  • Matching variables:      %d\n", len(result.Matching))
	fmt.Println()
}

func displayMatchingVariables(matching []string) {
	if len(matching) > 0 {
		ui.Success("✅ Variables present in both config and deployment (%d):", len(matching))
		for _, v := range matching {
			fmt.Printf("  ✓ %s\n", v)
		}
		fmt.Println()
	}
}

func displayConfigOnlyVariables(configOnly []string) {
	if len(configOnly) > 0 {
		ui.Warn("⚠️  Variables in configuration but NOT used in deployment (%d):", len(configOnly))
		for _, v := range configOnly {
			fmt.Printf("  ⚠ %s\n", v)
		}
		ui.Warn("These variables are configured but not used in your Kubernetes deployment.")
		ui.Warn("Consider removing them from config or adding them to the deployment.")
		fmt.Println()
	}
}

func displayDeploymentOnlyVariables(deploymentOnly []string) {
	if len(deploymentOnly) > 0 {
		ui.Warn("⚠️  Variables in deployment but NOT defined in configuration (%d):", len(deploymentOnly))
		for _, v := range deploymentOnly {
			fmt.Printf("  ⚠ %s\n", v)
		}
		ui.Warn("These variables are used in deployment but not managed by yeet.")
		ui.Warn("Consider adding them to your env.config.json if they should be managed.")
		fmt.Println()
	}
}

func displayOverallStatus(result ComparisonResult) {
	if len(result.InConfigOnly) == 0 && len(result.InDeploymentOnly) == 0 {
		ui.Success("🎉 Perfect match! All variables are consistent between config and deployment.")
	} else {
		ui.Info("💡 Recommendations:")
		if len(result.InConfigOnly) > 0 {
			fmt.Println("  • Review unused config variables and remove if not needed")
		}
		if len(result.InDeploymentOnly) > 0 {
			fmt.Println("  • Add missing variables to env.config.json for centralized management")
		}
	}
}
