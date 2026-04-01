package cmd

import (
	"github.com/spf13/cobra"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Show trading dashboard summary",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := getClient(cmd)
		dashboard, err := c.GetDashboard(cmd.Context())
		if err != nil {
			return err
		}
		return printResult(cmd, dashboard)
	},
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
}
