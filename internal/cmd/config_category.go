package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func configCategoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "category",
		Short: "Manage categories",
	}

	cmd.AddCommand(configCategoryAddCmd())
	cmd.AddCommand(configCategoryListCmd())
	cmd.AddCommand(configCategoryEditCmd())
	cmd.AddCommand(configCategoryDeleteCmd())

	return cmd
}

func configCategoryAddCmd() *cobra.Command {
	var name, desc string
	var budget float64

	c := &cobra.Command{
		Use:   "add",
		Short: "Add a new category",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			budgetCents := int64(budget * 100)
			id, err := database.CreateCategory(name, desc, budgetCents)
			if err != nil {
				return fmt.Errorf("add category: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Added category '%s' (ID: %d) with budget $%.2f\n", name, id, budget)
			return nil
		},
	}

	c.Flags().StringVar(&name, "name", "", "Category name")
	c.Flags().StringVar(&desc, "desc", "", "Category description")
	c.Flags().Float64Var(&budget, "budget", 0, "Monthly budget in dollars")
	return c
}

func configCategoryListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all categories",
		RunE: func(cmd *cobra.Command, args []string) error {
			cats, err := database.ListCategories()
			if err != nil {
				return fmt.Errorf("list categories: %w", err)
			}
			if len(cats) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No categories found.")
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "%-3s %-20s %-30s %s\n", "ID", "Name", "Description", "Budget")
			fmt.Fprintln(cmd.OutOrStdout(), "--------------------------------------------------------------")
			for _, cat := range cats {
				budget := float64(cat.MonthlyBudgetCents) / 100
				fmt.Fprintf(cmd.OutOrStdout(), "%-3d %-20s %-30s $%.2f\n", cat.ID, cat.Name, cat.Description, budget)
			}
			return nil
		},
	}
}

func configCategoryEditCmd() *cobra.Command {
	var id int64
	var name, desc string
	var budget float64

	c := &cobra.Command{
		Use:   "edit",
		Short: "Edit a category",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == 0 {
				return fmt.Errorf("--id is required")
			}
			if name == "" {
				return fmt.Errorf("--name is required")
			}
			budgetCents := int64(budget * 100)
			if err := database.UpdateCategory(id, name, desc, budgetCents); err != nil {
				return fmt.Errorf("edit category: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Updated category ID %d\n", id)
			return nil
		},
	}

	c.Flags().Int64Var(&id, "id", 0, "Category ID")
	c.Flags().StringVar(&name, "name", "", "Category name")
	c.Flags().StringVar(&desc, "desc", "", "Category description")
	c.Flags().Float64Var(&budget, "budget", 0, "Monthly budget in dollars")
	return c
}

func configCategoryDeleteCmd() *cobra.Command {
	var id int64

	c := &cobra.Command{
		Use:   "delete",
		Short: "Delete a category",
		RunE: func(cmd *cobra.Command, args []string) error {
			if id == 0 {
				return fmt.Errorf("--id is required")
			}
			if err := database.DeleteCategory(id); err != nil {
				return fmt.Errorf("delete category: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Deleted category ID %d\n", id)
			return nil
		},
	}

	c.Flags().Int64Var(&id, "id", 0, "Category ID")
	return c
}
