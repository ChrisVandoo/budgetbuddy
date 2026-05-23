package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ChrisVandoo/budgetbuddy/internal/parse"
)

func configSourceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "source",
		Short: "Manage bank sources",
	}

	cmd.AddCommand(configSourceListCmd())
	cmd.AddCommand(configSourceDeleteCmd())

	return cmd
}

func configSourceListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all configured sources",
		RunE: func(cmd *cobra.Command, args []string) error {
			sources, err := parse.LoadSources(sourcesPath)
			if err != nil {
				return fmt.Errorf("load sources: %w", err)
			}
			if len(sources.Sources) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No sources configured.")
				return nil
			}
			for key, src := range sources.Sources {
				fmt.Fprintf(cmd.OutOrStdout(), "Name: %s\n", src.Name)
				fmt.Fprintf(cmd.OutOrStdout(), "Headers: %s\n", key)
				fmt.Fprintf(cmd.OutOrStdout(), "Date column: %s (format: %s)\n",
					src.Mapping.Date.Header, src.Mapping.Date.Format)
				fmt.Fprintf(cmd.OutOrStdout(), "Description column: %s\n", src.Mapping.Description.Header)
				if src.Mapping.Amount.SingleColumn {
					fmt.Fprintf(cmd.OutOrStdout(), "Amount: single column (%s)\n", src.Mapping.Amount.HeaderOut)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "Amount: dual column (in: %s, out: %s)\n",
						src.Mapping.Amount.HeaderIn, src.Mapping.Amount.HeaderOut)
				}
				fmt.Fprintln(cmd.OutOrStdout())
			}
			return nil
		},
	}
}

func configSourceDeleteCmd() *cobra.Command {
	var name string

	c := &cobra.Command{
		Use:   "delete",
		Short: "Delete a source by name",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			sources, err := parse.LoadSources(sourcesPath)
			if err != nil {
				return fmt.Errorf("load sources: %w", err)
			}

			key := ""
			for k, src := range sources.Sources {
				if src.Name == name {
					key = k
					break
				}
			}
			if key == "" {
				return fmt.Errorf("source %q not found", name)
			}

			delete(sources.Sources, key)
			if err := parse.SaveSources(sourcesPath, sources); err != nil {
				return fmt.Errorf("save sources: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Deleted source %q\n", name)
			return nil
		},
	}

	c.Flags().StringVar(&name, "name", "", "Source name")
	return c
}
