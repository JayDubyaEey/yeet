package cli

import (
	"testing"

	"github.com/JayDubyaEey/yeet/pkg/version"
)

func TestIsBuiltFromSource(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		expected bool
	}{
		{"dev version", "dev", true},
		{"release version", "v1.2.3", false},
		{"release with git", "v1.2.3-1-g2a79a10", false},
		{"git hash only", "g2a79a10", true},
		{"no v prefix", "1.2.3", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Temporarily set version
			oldVersion := version.Version
			version.Version = tt.version
			defer func() { version.Version = oldVersion }()

			got := isBuiltFromSource()
			if got != tt.expected {
				t.Errorf("isBuiltFromSource() with version %q = %v, want %v", tt.version, got, tt.expected)
			}
		})
	}
}
