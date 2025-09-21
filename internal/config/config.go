package config

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)

// ValueType represents the type of a configuration value
type ValueType string

const (
	ValueTypeKeyvault ValueType = "keyvault"
	ValueTypeLiteral  ValueType = "literal"
)

// ValueSpec represents a value specification with type and value
type ValueSpec struct {
	Type  ValueType `json:"type"`
	Value string    `json:"value"`
}

// Mapping represents a single env var mapping with support for environments
type Mapping struct {
	// New enhanced format
	Local  *ValueSpec `json:"local,omitempty"`
	Docker *ValueSpec `json:"docker,omitempty"`

	// Global fallback (when not environment-specific)
	Type  ValueType `json:"type,omitempty"`
	Value string    `json:"value,omitempty"`
}

// Environment represents the target environment
type Environment string

const (
	EnvLocal  Environment = "local"
	EnvDocker Environment = "docker"
)

// Config represents the env.config.json structure
type Config struct {
	KeyVaultName string             `json:"keyVaultName"`
	Mappings     map[string]Mapping `json:"mappings"`
}

// GetValueSpec returns the appropriate ValueSpec for the given environment
func (m *Mapping) GetValueSpec(env Environment) *ValueSpec {
	switch env {
	case EnvLocal:
		if m.Local != nil {
			return m.Local
		}
	case EnvDocker:
		if m.Docker != nil {
			return m.Docker
		}
	}

	// Fallback to global value if present
	if m.Type != "" && m.Value != "" {
		return &ValueSpec{
			Type:  m.Type,
			Value: m.Value,
		}
	}

	return nil
}

// IsKeyvaultSecret returns true if the value should be fetched from Key Vault
func (v *ValueSpec) IsKeyvaultSecret() bool {
	return v != nil && v.Type == ValueTypeKeyvault
}

// IsLiteral returns true if the value is a literal value
func (v *ValueSpec) IsLiteral() bool {
	return v != nil && v.Type == ValueTypeLiteral
}

// rawMapping helps parse JSON where value can be string or object
type rawMapping struct {
	KeyVaultName string                     `json:"keyVaultName"`
	Mappings     map[string]json.RawMessage `json:"mappings"`
}

var envVarRegex = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)

// Load reads and validates env.config.json
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	cfg, err := parseConfig(data, path)
	if err != nil {
		return nil, err
	}

	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// parseConfig parses the JSON data into a Config struct
func parseConfig(data []byte, path string) (*Config, error) {
	var raw rawMapping
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON in %s: %w", path, err)
	}

	cfg := &Config{
		KeyVaultName: raw.KeyVaultName,
		Mappings:     make(map[string]Mapping),
	}

	for key, rawVal := range raw.Mappings {
		mapping, err := parseMapping(key, rawVal)
		if err != nil {
			return nil, err
		}
		cfg.Mappings[key] = mapping
	}

	return cfg, nil
}

// parseMapping parses a single mapping from JSON
func parseMapping(key string, rawVal json.RawMessage) (Mapping, error) {
	var mapping Mapping

	// Try simple string first - convert to global keyvault type
	var secretName string
	if err := json.Unmarshal(rawVal, &secretName); err == nil {
		mapping.Type = ValueTypeKeyvault
		mapping.Value = secretName
		return mapping, nil
	}

	// Try complex object
	if err := json.Unmarshal(rawVal, &mapping); err != nil {
		return mapping, fmt.Errorf("invalid mapping for %s: %w", key, err)
	}

	return mapping, nil
}

// validateConfig validates the configuration
func validateConfig(cfg *Config) error {
	if cfg.KeyVaultName == "" {
		return fmt.Errorf("keyVaultName is required")
	}
	if len(cfg.Mappings) == 0 {
		return fmt.Errorf("at least one mapping is required")
	}
	for key, mapping := range cfg.Mappings {
		if err := validateMapping(key, mapping); err != nil {
			return err
		}
	}
	return nil
}

// validateMapping validates a single mapping
func validateMapping(key string, mapping Mapping) error {
	if err := validateEnvironmentVarName(key); err != nil {
		return err
	}

	if err := validateMappingHasValues(key, mapping); err != nil {
		return err
	}

	return validateIndividualSpecs(key, mapping)
}

func validateEnvironmentVarName(key string) error {
	if !envVarRegex.MatchString(key) {
		return fmt.Errorf("invalid env var name: %s (must match ^[A-Z_][A-Z0-9_]*$)", key)
	}
	return nil
}

func validateMappingHasValues(key string, mapping Mapping) error {
	hasLocal := mapping.Local != nil
	hasDocker := mapping.Docker != nil
	hasGlobal := mapping.Type != "" && mapping.Value != ""

	if !hasLocal && !hasDocker && !hasGlobal {
		return fmt.Errorf("mapping for %s must have at least one value specification", key)
	}
	return nil
}

func validateIndividualSpecs(key string, mapping Mapping) error {
	if mapping.Local != nil {
		if err := validateValueSpec(key, "local", mapping.Local); err != nil {
			return err
		}
	}
	if mapping.Docker != nil {
		if err := validateValueSpec(key, "docker", mapping.Docker); err != nil {
			return err
		}
	}
	if mapping.Type != "" && mapping.Value != "" {
		globalSpec := &ValueSpec{Type: mapping.Type, Value: mapping.Value}
		if err := validateValueSpec(key, "global", globalSpec); err != nil {
			return err
		}
	}
	return nil
}

// validateValueSpec validates a single value specification
func validateValueSpec(key, context string, spec *ValueSpec) error {
	if spec == nil {
		return fmt.Errorf("value spec cannot be nil for %s (%s)", key, context)
	}
	if spec.Value == "" {
		return fmt.Errorf("value cannot be empty for %s (%s)", key, context)
	}
	if spec.Type != ValueTypeKeyvault && spec.Type != ValueTypeLiteral {
		return fmt.Errorf("invalid type %q for %s (%s): must be %q or %q",
			spec.Type, key, context, ValueTypeKeyvault, ValueTypeLiteral)
	}
	return nil
}
