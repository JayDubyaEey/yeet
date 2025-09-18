package envwriter

import (
	"bufio"
	"fmt"
	"github.com/JayDubyaEey/yeet/internal/config"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var envLineRegex = regexp.MustCompile(`^([A-Z_][A-Z0-9_]*)=`)

// WriteEnvFile writes env vars to a file atomically
func WriteEnvFile(path string, vars map[string]string, header string) error {
	// Create temp file in same directory for atomic write
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".env-tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath) // Clean up on any error

	// Write header
	if header != "" {
		if _, err := tmp.WriteString(header + "\n"); err != nil {
			tmp.Close()
			return err
		}
	}

	// Sort keys for stable output
	keys := make([]string, 0, len(vars))
	for k := range vars {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Write each var
	for _, key := range keys {
		value := vars[key]
		line := fmt.Sprintf("%s=%s\n", key, quoteValue(value))
		if _, err := tmp.WriteString(line); err != nil {
			tmp.Close()
			return err
		}
	}

	// Sync to disk
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	tmp.Close()

	// Atomic rename
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// quoteValue adds quotes if the value contains special characters
func quoteValue(value string) string {
	// Check if quoting is needed
	needsQuotes := strings.ContainsAny(value, " \t\n\r#\"'") || strings.TrimSpace(value) != value

	if !needsQuotes {
		return value
	}

	// Escape quotes and newlines
	escaped := strings.ReplaceAll(value, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
	escaped = strings.ReplaceAll(escaped, "\n", "\\n")
	escaped = strings.ReplaceAll(escaped, "\r", "\\r")

	return fmt.Sprintf("\"%s\"", escaped)
}

// ReadKeyValues reads existing env file and returns key-value pairs
func ReadKeyValues(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]string), nil
		}
		return nil, err
	}
	defer file.Close()

	vars := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Extract key (we only care about keys for unmapped detection)
		matches := envLineRegex.FindStringSubmatch(line)
		if len(matches) > 1 {
			vars[matches[1]] = "" // Value doesn't matter for our use case
		}
	}

	return vars, scanner.Err()
}

// MergeRetainUnknowns merges new values with existing, retaining unmapped keys
func MergeRetainUnknowns(newVars, existingVars map[string]string, mappings map[string]config.Mapping) map[string]string {
	result := make(map[string]string)

	// Add all new vars
	for k, v := range newVars {
		result[k] = v
	}

	// Retain existing vars not in mappings
	for k := range existingVars {
		if _, inMappings := mappings[k]; !inMappings {
			if _, alreadySet := result[k]; !alreadySet {
				// This key is not mapped, retain it
				result[k] = existingVars[k]
			}
		}
	}

	return result
}

// UnmappedKeys returns keys present in existing but not in mappings
func UnmappedKeys(existingVars map[string]string, mappings map[string]config.Mapping) []string {
	var unmapped []string
	for k := range existingVars {
		if _, ok := mappings[k]; !ok {
			unmapped = append(unmapped, k)
		}
	}
	sort.Strings(unmapped)
	return unmapped
}
