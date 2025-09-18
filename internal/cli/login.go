package cli

import (
	"context"
	"github.com/JayDubyaEey/yeet/internal/provider/azcli"
	"github.com/JayDubyaEey/yeet/internal/ui"
	"github.com/spf13/cobra"
)

func newLoginCmd() *cobra.Command {
	var tenant string
	var subscription string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Azure via Azure CLI",
		RunE: func(cmd *cobra.Command, args []string) error {
			prov := azcli.NewDefault()
			if err := prov.Login(context.Background(), tenant, subscription); err != nil {
				return err
			}
			ui.Success("logged in to Azure CLI")
			return nil
		},
	}
	cmd.Flags().StringVar(&tenant, "tenant", "", "Tenant ID or domain to log into")
	cmd.Flags().StringVar(&subscription, "subscription", "", "Subscription ID or name to set after login")
	return cmd
}

func newLogoutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout from Azure CLI",
		RunE: func(cmd *cobra.Command, args []string) error {
			prov := azcli.NewDefault()
			if err := prov.Logout(context.Background()); err != nil {
				return err
			}
			ui.Success("logged out of Azure CLI")
			return nil
		},
	}
	return cmd
}
