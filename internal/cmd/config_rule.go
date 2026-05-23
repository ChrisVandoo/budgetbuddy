package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func configRuleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rule",
		Short: "Manage categorization rules",
	}

	cmd.AddCommand(configRuleAddCmd())
	cmd.AddCommand(configRuleListCmd())
	cmd.AddCommand(configRuleDeleteCmd())

	return cmd
}

func configRuleAddCmd() *cobra.Command {
	var categoryName, pattern string

	c := &cobra.Command{
		Use:   "add",
		Short: "Add a new categorization rule",
		RunE: func(cmd *cobra.Command, args []string) error {
			if categoryName == "" {
				return fmt.Errorf("--category is required")
			}
			if pattern == "" {
				return fmt.Errorf("--pattern is required")
			}

			cat, err := database.GetCategoryByName(categoryName)
			if err != nil {
				return fmt.Errorf("category %q not found: %w", categoryName, err)
			}

			id, err := database.CreateRule(cat.ID, pattern)
			if err != nil {
				return fmt.Errorf("add rule: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Added rule ID %d: pattern %q -> category %q\n", id, pattern, categoryName)
			return nil
		},
	}

	c.Flags().StringVar(&categoryName, "category", "", "Category name")
	c.Flags().StringVar(&pattern, "pattern", "", "Glob pattern to match")
	return c
}

func configRuleListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all categorization rules",
		RunE: func(cmd *cobra.Command, args []string) error {
			rules, err := database.ListRules()
			if err != nil {
				return fmt.Errorf("list rules: %w", err)
			}
			if len(rules) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No rules found.")
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%-3s %-20s %s\n", "ID", "Category", "Pattern")
			fmt.Fprintln(cmd.OutOrStdout(), "-----------------------------------------")
			for _, r := range rules {
				fmt.Fprintf(cmd.OutOrStdout(), "%-3d %-20s %s\n", r.ID, r.CategoryName, r.Pattern)
			}
			return nil
		},
	}
}

func configRuleDeleteCmd() *cobra.Command {
	var id int64

	c := &cobra.Command{
		Use:   "delete",
		Short: "Delete a categorization rule",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == 0 {
				return fmt.Errorf("--id is required")
			}
			if err := database.DeleteRule(id); err != nil {
				return fmt.Errorf("delete rule: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Deleted rule ID %d\n", id)
			return nil
		},
	}

	c.Flags().Int64Var(&id, "id", 0, "Rule ID")
	return c
}
