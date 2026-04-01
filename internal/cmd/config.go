package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tradekit-dev/tradekit-cli/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage CLI configuration",
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a config value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, value := args[0], args[1]
		if err := config.Set(key, value); err != nil {
			return fmt.Errorf("failed to set config: %w", err)
		}
		fmt.Printf("Set %s = %s\n", key, value)
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a config value",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		value := config.Get(args[0])
		if value == "" {
			fmt.Printf("%s is not set\n", args[0])
		} else {
			fmt.Println(value)
		}
		return nil
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show all configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		settings := config.AllSettings()
		return printResult(cmd, settings)
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configListCmd)
}
