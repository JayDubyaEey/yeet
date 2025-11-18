package updater

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/JayDubyaEey/yeet/pkg/version"
)

const (
	githubAPIURL = "https://api.github.com/repos/JayDubyaEey/yeet/releases"
	owner        = "JayDubyaEey"
	repo         = "yeet"
)

// Release represents a GitHub release
type Release struct {
	TagName    string  `json:"tag_name"`
	Name       string  `json:"name"`
	Draft      bool    `json:"draft"`
	Prerelease bool    `json:"prerelease"`
	Assets     []Asset `json:"assets"`
}

// Asset represents a release asset
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// Updater handles self-updates
type Updater struct {
	CurrentVersion string
	httpClient     *http.Client
}

// New creates a new Updater
func New() *Updater {
	return &Updater{
		CurrentVersion: version.Version,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CheckForUpdate checks if a newer version is available
func (u *Updater) CheckForUpdate(ctx context.Context, targetVersion string) (*Release, bool, error) {
	var release *Release
	var err error

	if targetVersion != "" {
		// Fetch specific version
		release, err = u.fetchRelease(ctx, targetVersion)
	} else {
		// Fetch latest release
		release, err = u.fetchLatestRelease(ctx)
	}

	if err != nil {
		return nil, false, err
	}

	if targetVersion != "" {
		// For specific version, always consider it an update
		return release, true, nil
	}

	// Compare versions
	needsUpdate := u.needsUpdate(release.TagName)
	return release, needsUpdate, nil
}

// fetchLatestRelease fetches the latest GitHub release
func (u *Updater) fetchLatestRelease(ctx context.Context) (*Release, error) {
	url := fmt.Sprintf("%s/latest", githubAPIURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add GitHub token if available
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("GitHub API rate limit exceeded. Please try again later or set GITHUB_TOKEN environment variable")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &release, nil
}

// fetchRelease fetches a specific release by tag
func (u *Updater) fetchRelease(ctx context.Context, tag string) (*Release, error) {
	// Ensure tag starts with 'v'
	if !strings.HasPrefix(tag, "v") {
		tag = "v" + tag
	}

	url := fmt.Sprintf("%s/tags/%s", githubAPIURL, tag)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add GitHub token if available
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("release %s not found", tag)
	}

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("GitHub API rate limit exceeded. Please try again later or set GITHUB_TOKEN environment variable")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &release, nil
}

// needsUpdate compares versions to determine if an update is needed
func (u *Updater) needsUpdate(latestVersion string) bool {
	current := normalizeVersion(u.CurrentVersion)
	latest := normalizeVersion(latestVersion)

	// If current version is "dev", it needs update
	if current == "dev" {
		return true
	}

	// Simple string comparison (works for semver if properly formatted)
	return current != latest && latest > current
}

// normalizeVersion removes 'v' prefix and git suffix from version strings
func normalizeVersion(v string) string {
	v = strings.TrimPrefix(v, "v")
	// Remove git commit suffix like "-1-g2a79a10"
	if idx := strings.Index(v, "-"); idx != -1 {
		v = v[:idx]
	}
	return v
}

// Update performs the update
func (u *Updater) Update(ctx context.Context, release *Release) error {
	// Find the appropriate asset for current platform
	asset := u.findAsset(release)
	if asset == nil {
		return fmt.Errorf("no suitable asset found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// Download the asset
	tmpFile, err := u.downloadAsset(ctx, asset)
	if err != nil {
		return fmt.Errorf("failed to download asset: %w", err)
	}
	defer os.Remove(tmpFile)

	// Verify checksum
	if err := u.verifyChecksum(ctx, release, asset.Name, tmpFile); err != nil {
		return fmt.Errorf("checksum verification failed: %w", err)
	}

	// Extract the binary
	binaryPath, err := u.extractBinary(tmpFile, asset.Name)
	if err != nil {
		return fmt.Errorf("failed to extract binary: %w", err)
	}
	defer os.Remove(binaryPath)

	// Replace current binary
	if err := u.replaceBinary(binaryPath); err != nil {
		return fmt.Errorf("failed to replace binary: %w", err)
	}

	return nil
}

// findAsset finds the appropriate asset for the current platform
func (u *Updater) findAsset(release *Release) *Asset {
	osName := runtime.GOOS
	archName := runtime.GOARCH

	// Map Go arch names to release names
	archMap := map[string]string{
		"amd64": "x86_64",
		"386":   "i386",
		"arm64": "arm64",
	}

	releaseArch, ok := archMap[archName]
	if !ok {
		releaseArch = archName
	}

	// Determine file extension
	var ext string
	if osName == "windows" {
		ext = ".zip"
	} else {
		ext = ".tar.gz"
	}

	// Find matching asset
	pattern := fmt.Sprintf("yeet_%s_%s_%s%s", strings.TrimPrefix(release.TagName, "v"), osName, releaseArch, ext)

	for i := range release.Assets {
		if release.Assets[i].Name == pattern {
			return &release.Assets[i]
		}
	}

	return nil
}

// downloadAsset downloads the asset to a temporary file
func (u *Updater) downloadAsset(ctx context.Context, asset *Asset) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", asset.BrowserDownloadURL, nil)
	if err != nil {
		return "", err
	}

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "yeet-update-*")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	// Copy content
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	return tmpFile.Name(), nil
}

// extractBinary extracts the binary from the archive
func (u *Updater) extractBinary(archivePath, assetName string) (string, error) {
	if strings.HasSuffix(assetName, ".zip") {
		return u.extractFromZip(archivePath)
	}
	return u.extractFromTarGz(archivePath)
}

// extractFromZip extracts binary from zip archive
func (u *Updater) extractFromZip(zipPath string) (string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	binaryName := "yeet.exe"
	for _, f := range r.File {
		if f.Name == binaryName || filepath.Base(f.Name) == binaryName {
			return u.extractZipFile(f)
		}
	}

	return "", fmt.Errorf("binary not found in archive")
}

// extractZipFile extracts a single file from zip
func (u *Updater) extractZipFile(f *zip.File) (string, error) {
	rc, err := f.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	tmpFile, err := os.CreateTemp("", "yeet-binary-*")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, rc); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	// Make executable
	if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	return tmpFile.Name(), nil
}

// extractFromTarGz extracts binary from tar.gz archive
func (u *Updater) extractFromTarGz(tarGzPath string) (string, error) {
	f, err := os.Open(tarGzPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	binaryName := "yeet"

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		if header.Name == binaryName || filepath.Base(header.Name) == binaryName {
			tmpFile, err := os.CreateTemp("", "yeet-binary-*")
			if err != nil {
				return "", err
			}
			defer tmpFile.Close()

			if _, err := io.Copy(tmpFile, tr); err != nil {
				os.Remove(tmpFile.Name())
				return "", err
			}

			// Make executable
			if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
				os.Remove(tmpFile.Name())
				return "", err
			}

			return tmpFile.Name(), nil
		}
	}

	return "", fmt.Errorf("binary not found in archive")
}

// replaceBinary replaces the current binary with the new one
func (u *Updater) replaceBinary(newBinaryPath string) error {
	// Get current executable path
	currentPath, err := os.Executable()
	if err != nil {
		return err
	}

	// Resolve symlinks
	currentPath, err = filepath.EvalSymlinks(currentPath)
	if err != nil {
		return err
	}

	// Create backup
	backupPath := currentPath + ".backup"
	if err := copyFile(currentPath, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Try to replace the binary
	if err := copyFile(newBinaryPath, currentPath); err != nil {
		// Restore backup on failure
		_ = copyFile(backupPath, currentPath)
		os.Remove(backupPath)
		return fmt.Errorf("failed to replace binary: %w (you may need elevated permissions)", err)
	}

	// Remove backup on success
	os.Remove(backupPath)

	// Ensure executable permissions
	if err := os.Chmod(currentPath, 0755); err != nil {
		return fmt.Errorf("failed to set permissions: %w", err)
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return out.Close()
}

// verifyChecksum downloads and verifies the checksum file
func (u *Updater) verifyChecksum(ctx context.Context, release *Release, assetName string, filePath string) error {
	// Find checksums.txt asset
	var checksumAsset *Asset
	for i := range release.Assets {
		if release.Assets[i].Name == "checksums.txt" {
			checksumAsset = &release.Assets[i]
			break
		}
	}

	if checksumAsset == nil {
		// No checksums available, skip verification
		return nil
	}

	// Download checksums file
	req, err := http.NewRequestWithContext(ctx, "GET", checksumAsset.BrowserDownloadURL, nil)
	if err != nil {
		return err
	}

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download checksums: status %d", resp.StatusCode)
	}

	// Parse checksums file and find the matching one
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[1] == assetName {
			expectedChecksum := parts[0]

			// Compute actual checksum
			actualChecksum, err := computeSHA256(filePath)
			if err != nil {
				return fmt.Errorf("failed to compute checksum: %w", err)
			}

			if actualChecksum != expectedChecksum {
				return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
			}

			return nil
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return fmt.Errorf("checksum not found for %s", assetName)
}

// computeSHA256 computes SHA256 hash of a file
func computeSHA256(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
