package cli

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/JayDubyaEey/yeet/internal/ui"
	"github.com/JayDubyaEey/yeet/internal/updater"
	"github.com/JayDubyaEey/yeet/pkg/version"
)

var (
	updateVersion string
	forceUpdate   bool
)

func newUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update yeet to the latest version",
		Long: `Update yeet to the latest version automatically.

This command will:
  - Check for the latest release on GitHub
  - Download the appropriate binary for your platform
  - Verify the checksum for security
  - Replace the current binary with the new version

Note: If yeet is installed in a system directory (like /usr/local/bin),
you may need to run this command with elevated permissions (sudo).`,
		Example: `  # Update to the latest version
  yeet update

  # Update to a specific version
  yeet update --version v1.2.3

  # Force reinstall current version
  yeet update --force`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(cmd.Context())
		},
	}

	cmd.Flags().StringVar(&updateVersion, "version", "", "Specific version to update to (e.g., v1.2.3)")
	cmd.Flags().BoolVar(&forceUpdate, "force", false, "Force reinstall even if already on latest version")

	return cmd
}

func runUpdate(ctx context.Context) error {
	// Check if built from source
	if isBuiltFromSource() {
		ui.Warn("It appears you built yeet from source.")
		ui.Info("To update, run: go install github.com/JayDubyaEey/yeet/cmd@latest")
		ui.Info("Or rebuild from the latest source code.")
		return fmt.Errorf("cannot self-update when built from source")
	}

	u := updater.New()

	ui.Info("Checking for updates...")
	ui.Info("Current version: %s (%s/%s)", u.CurrentVersion, runtime.GOOS, runtime.GOARCH)

	// Check for update
	release, needsUpdate, err := u.CheckForUpdate(ctx, updateVersion)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !needsUpdate && !forceUpdate {
		ui.Success("You are already on the latest version: %s", release.TagName)
		return nil
	}

	if forceUpdate && !needsUpdate {
		ui.Info("Force reinstalling version: %s", release.TagName)
	} else {
		ui.Info("Latest version: %s", release.TagName)
	}

	// Download and install update
	ui.Info("Downloading %s...", release.TagName)

	if err := u.Update(ctx, release); err != nil {
		return err
	}

	ui.Success("✅ Successfully updated to %s", release.TagName)
	ui.Info("")
	ui.Info("Run 'yeet --version' to verify the update.")

	return nil
}

// isBuiltFromSource detects if the binary was built from source
func isBuiltFromSource() bool {
	// If version is "dev" or contains git info without a proper tag, likely built from source
	v := version.Version

	// Check if it's a dev build
	if v == "dev" {
		return true
	}

	// Check if it doesn't start with 'v' (release versions do)
	if !strings.HasPrefix(v, "v") {
		return true
	}

	// If it has git commit info but no valid semver, it's from source
	// Example: "v0.1.2-1-g2a79a10" is OK, "g2a79a10" is not
	if strings.HasPrefix(v, "g") {
		return true
	}

	return false
}
