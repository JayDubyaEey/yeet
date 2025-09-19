package cli

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/JayDubyaEey/yeet/internal/ui"
	"github.com/JayDubyaEey/yeet/pkg/version"
)

var (
	configPath    string
	vaultOverride string
	noColor       bool
	verbose       bool
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "yeet",
		Short:         "Yeet envs from Azure Key Vault into .env and docker.env",
		Long:          "Yeet pulls secrets from Azure Key Vault and generates .env and docker.env for local dev and docker.",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			ui.Setup(noColor, verbose)
		},
	}

	cmd.PersistentFlags().StringVar(&configPath, "config", "env.config.json", "Path to env configuration file")
	cmd.PersistentFlags().StringVar(&vaultOverride, "vault", "", "Override Key Vault name from config")
	cmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")

	cmd.Version = version.Version + fmt.Sprintf(" (%s/%s)", runtime.GOOS, runtime.GOARCH)

	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newFetchCmd())
	cmd.AddCommand(newRefreshCmd())
	cmd.AddCommand(newRunCmd())
	cmd.AddCommand(newLoginCmd())
	cmd.AddCommand(newLogoutCmd())
	cmd.AddCommand(newValidateCmd())
	cmd.AddCommand(newListCmd())

	return cmd
}

// Execute runs the CLI
func Execute() {
	if err := newRootCmd().Execute(); err != nil {
		ui.Error("%s", err.Error())
		os.Exit(1)
	}
}
