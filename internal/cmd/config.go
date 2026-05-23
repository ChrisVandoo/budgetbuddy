package cmd

import (
	"github.com/spf13/cobra"
)

func configCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
	}

	cmd.AddCommand(configCategoryCmd())
	cmd.AddCommand(configRuleCmd())
	cmd.AddCommand(configSourceCmd())

	return cmd
}
