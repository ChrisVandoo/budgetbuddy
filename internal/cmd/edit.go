package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

func editCmd() *cobra.Command {
	var categoryName string
	var id int64

	c := &cobra.Command{
		Use:   "edit [id]",
		Short: "Change a transaction's category",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				parsed, err := strconv.ParseInt(args[0], 10, 64)
				if err != nil {
					return fmt.Errorf("invalid transaction ID: %s", args[0])
				}
				id = parsed
			}
			if id == 0 {
				return fmt.Errorf("transaction ID is required (use --id or pass as argument)")
			}
			if categoryName == "" {
				return fmt.Errorf("--category is required")
			}

			cat, err := database.GetCategoryByName(categoryName)
			if err != nil {
				return fmt.Errorf("category %q not found: %w", categoryName, err)
			}

			if err := database.UpdateTransactionCategory(id, cat.ID); err != nil {
				return fmt.Errorf("edit transaction: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Updated transaction %d to category %q\n", id, categoryName)
			return nil
		},
	}

	c.Flags().Int64Var(&id, "id", 0, "Transaction ID")
	c.Flags().StringVar(&categoryName, "category", "", "Category name")
	return c
}
