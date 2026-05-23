package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ChrisVandoo/budgetbuddy/internal/types"
)

func reportCmd() *cobra.Command {
	var month, year int
	var csvOutput bool

	c := &cobra.Command{
		Use:   "report",
		Short: "Show monthly spending report",
		RunE: func(cmd *cobra.Command, args []string) error {
			if month < 1 || month > 12 {
				return fmt.Errorf("month must be 1-12")
			}
			if year < 2000 || year > 3000 {
				return fmt.Errorf("year must be valid")
			}

			items, summaries, err := database.GetReportData(month, year)
			if err != nil {
				return fmt.Errorf("get report: %w", err)
			}

			if csvOutput {
				return writeCSVReport(cmd.OutOrStdout(), month, year, items, summaries)
			}

			printReport(cmd.OutOrStdout(), month, year, items, summaries)
			return nil
		},
	}

	c.Flags().IntVarP(&month, "month", "m", 0, "Month (1-12)")
	c.Flags().IntVarP(&year, "year", "y", 0, "Year (e.g., 2026)")
	c.Flags().BoolVar(&csvOutput, "csv", false, "Write CSV report file")
	return c
}

func printReport(w io.Writer, month, year int, items []types.LineItem, summaries []types.CategorySummary) {
	monthName := monthNames[month]
	fmt.Fprintf(w, "=== %s %d Report ===\n\n", monthName, year)
	fmt.Fprintln(w, "--- Detail ---")
	fmt.Fprintf(w, "%-4s %-12s %-50s %10s %s\n", "ID", "Date", "Description", "Amount", "Category")
	fmt.Fprintln(w, strings.Repeat("-", 100))
	for _, item := range items {
		amount := float64(item.AmountCents) / 100
		fmt.Fprintf(w, "%-4d %-12s %-50s %9.2f %s\n", item.ID, item.Date, truncate(item.Description, 50), amount, item.Category)
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, "--- Summary ---")
	fmt.Fprintf(w, "%-25s %15s %15s %15s\n", "Category", "Budget", "Actual", "Diff")
	fmt.Fprintln(w, strings.Repeat("-", 75))
	for _, s := range summaries {
		budget := float64(s.BudgetCents) / 100
		actual := float64(s.ActualCents) / 100
		diff := float64(s.RemainingCents) / 100
		fmt.Fprintf(w, "%-25s %12.2f %12.2f %12.2f\n", s.Category, budget, actual, diff)
	}

	if len(items) == 0 {
		fmt.Fprintln(w, "No transactions found for this period.")
	}
}

func writeCSVReport(w io.Writer, month, year int, items []types.LineItem, summaries []types.CategorySummary) error {
	filename := fmt.Sprintf("%02d-%d-report.csv", month, year)
	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create csv: %w", err)
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	writer.Write([]string{"ID", "Date", "Description", "Amount", "Category"})
	for _, item := range items {
		amount := float64(item.AmountCents) / 100
		writer.Write([]string{
			strconv.FormatInt(item.ID, 10),
			item.Date,
			item.Description,
			fmt.Sprintf("%.2f", amount),
			item.Category,
		})
	}

	writer.Write(nil)

	writer.Write([]string{"Category", "Budget", "Actual", "Diff"})
	for _, s := range summaries {
		budget := float64(s.BudgetCents) / 100
		actual := float64(s.ActualCents) / 100
		diff := float64(s.RemainingCents) / 100
		writer.Write([]string{
			s.Category,
			fmt.Sprintf("%.2f", budget),
			fmt.Sprintf("%.2f", actual),
			fmt.Sprintf("%.2f", diff),
		})
	}

	fmt.Fprintf(w, "Report written to %s\n", filename)
	return nil
}

var monthNames = map[int]string{
	1: "January", 2: "February", 3: "March", 4: "April",
	5: "May", 6: "June", 7: "July", 8: "August",
	9: "September", 10: "October", 11: "November", 12: "December",
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
