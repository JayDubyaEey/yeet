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
	KeyVaultName string                       `json:"keyVaultName"`
	Mappings     map[string]json.RawMessage `json:"mappings"`
}

var envVarRegex = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)

// Load reads and validates env.config.json
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var raw rawMapping
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("invalid JSON in %s: %w", path, err)
	}

	cfg := &Config{
		KeyVaultName: raw.KeyVaultName,
		Mappings:     make(map[string]Mapping),
	}

	// Normalize mappings: string or object -> Mapping
	for key, rawVal := range raw.Mappings {
		var mapping Mapping
		
		// Try simple string first
		var secretName string
		if err := json.Unmarshal(rawVal, &secretName); err == nil {
			mapping.Secret = secretName
		} else {
			// Try complex object
			if err := json.Unmarshal(rawVal, &mapping); err != nil {
				return nil, fmt.Errorf("invalid mapping for %s: must be string or {secret, docker}", key)
			}
		}
		
		cfg.Mappings[key] = mapping
	}

	// Validate
	if cfg.KeyVaultName == "" {
		return nil, fmt.Errorf("keyVaultName is required")
	}
	if len(cfg.Mappings) == 0 {
		return nil, fmt.Errorf("at least one mapping is required")
	}
	for key, mapping := range cfg.Mappings {
		if !envVarRegex.MatchString(key) {
			return nil, fmt.Errorf("invalid env var name: %s (must match ^[A-Z_][A-Z0-9_]*$)", key)
		}
		if mapping.Secret == "" {
			return nil, fmt.Errorf("secret name cannot be empty for %s", key)
		}
	}

	return cfg, nil
}
