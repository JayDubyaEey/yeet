package updater

import (
	"testing"
)

func TestNew(t *testing.T) {
	u := New()
	if u == nil {
		t.Fatal("New() returned nil")
	}
	if u.httpClient == nil {
		t.Error("httpClient is nil")
	}
}

func TestNormalizeVersion(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"v1.2.3", "1.2.3"},
		{"1.2.3", "1.2.3"},
		{"v1.2.3-1-g2a79a10", "1.2.3"},
		{"dev", "dev"},
		{"v0.1.2-dirty", "0.1.2"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeVersion(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeVersion(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNeedsUpdate(t *testing.T) {
	tests := []struct {
		name           string
		currentVersion string
		latestVersion  string
		want           bool
	}{
		{"same version", "v1.2.3", "v1.2.3", false},
		{"newer available", "v1.2.2", "v1.2.3", true},
		{"dev version", "dev", "v1.2.3", true},
		{"already latest", "v1.2.3", "v1.2.2", false},
		{"with git suffix current", "v1.2.3-1-g2a79a10", "v1.2.3", false},
		{"with git suffix both", "v1.2.2-1-g2a79a10", "v1.2.3-1-gabcdef0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &Updater{CurrentVersion: tt.currentVersion}
			got := u.needsUpdate(tt.latestVersion)
			if got != tt.want {
				t.Errorf("needsUpdate(%q, %q) = %v, want %v", tt.currentVersion, tt.latestVersion, got, tt.want)
			}
		})
	}
}

func TestFindAsset(t *testing.T) {
	release := &Release{
		TagName: "v1.2.3",
		Assets: []Asset{
			{Name: "yeet_1.2.3_linux_x86_64.tar.gz", BrowserDownloadURL: "https://example.com/linux"},
			{Name: "yeet_1.2.3_darwin_arm64.tar.gz", BrowserDownloadURL: "https://example.com/darwin"},
			{Name: "yeet_1.2.3_windows_x86_64.zip", BrowserDownloadURL: "https://example.com/windows"},
			{Name: "checksums.txt", BrowserDownloadURL: "https://example.com/checksums"},
		},
	}

	u := New()
	asset := u.findAsset(release)

	// Should find an asset (depends on runtime.GOOS/GOARCH)
	if asset == nil {
		t.Error("findAsset returned nil, expected to find a matching asset")
	}

	// Verify it's not the checksums file
	if asset != nil && asset.Name == "checksums.txt" {
		t.Error("findAsset returned checksums.txt instead of binary asset")
	}
}
