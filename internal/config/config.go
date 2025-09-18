package config

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)

// Mapping represents a single env var mapping
type Mapping struct {
	Secret string  `json:"secret"`
	Docker *string `json:"docker,omitempty"`
}

// Config represents the env.config.json structure
type Config struct {
	KeyVaultName string             `json:"keyVaultName"`
	Mappings     map[string]Mapping `json:"mappings"`
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

	// Try simple string first
	var secretName string
	if err := json.Unmarshal(rawVal, &secretName); err == nil {
		mapping.Secret = secretName
		return mapping, nil
	}

	// Try complex object
	if err := json.Unmarshal(rawVal, &mapping); err != nil {
		return mapping, fmt.Errorf("invalid mapping for %s: must be string or {secret, docker}", key)
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
	if !envVarRegex.MatchString(key) {
		return fmt.Errorf("invalid env var name: %s (must match ^[A-Z_][A-Z0-9_]*$)", key)
	}
	if mapping.Secret == "" {
		return fmt.Errorf("secret name cannot be empty for %s", key)
	}
	return nil
}
