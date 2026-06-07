package cmd

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/ChrisVandoo/budgetbuddy/internal/categorize"
	"github.com/ChrisVandoo/budgetbuddy/internal/parse"
	"github.com/ChrisVandoo/budgetbuddy/internal/prompt"
)

func parseCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "parse [paths...]",
		Short: "Import CSV files",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, paths []string) error {
			for _, path := range paths {
				info, err := os.Stat(path)
				if err != nil {
					return fmt.Errorf("access %s: %w", path, err)
				}

				if info.IsDir() {
					entries, err := os.ReadDir(path)
					if err != nil {
						return fmt.Errorf("read dir %s: %w", path, err)
					}
					for _, entry := range entries {
						if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".csv") {
							filename := fmt.Sprintf("%s/%s", path, entry.Name())
							if err := processCSVFile(cmd, filename); err != nil {
								return err
							}
						}
					}
				} else {
					if err := processCSVFile(cmd, path); err != nil {
						return err
					}
				}
			}
			return nil
		},
	}
	return c
}

func processCSVFile(cmd *cobra.Command, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	parser := parse.NewParser()
	if err := parser.ReadCSVFile(f); err != nil {
		return fmt.Errorf("failed to read %s: %w", path, err)
	}

	headers, err := parser.GetHeaderRecord()
	if err != nil {
		return fmt.Errorf("failed to get headers for %s: %w", path, err)
	}

	sources, err := parse.LoadSources(sourcesPath)
	if err != nil {
		return fmt.Errorf("load sources: %w", err)
	}

	headerKey, config, found := parse.DetectSource(headers, sources)
	if !found {
		wizard := prompt.NewSourceWizard(headers, path)
		prog := tea.NewProgram(wizard)
		model, err := prog.Run()
		if err != nil {
			return fmt.Errorf("source wizard: %w", err)
		}
		wiz := model.(*prompt.SourceWizard)
		if wiz.Cancelled() {
			return fmt.Errorf("source creation cancelled")
		}
		newConfig := wiz.Config()
		headerKey = strings.Join(headers, ",")
		sources.Sources[headerKey] = newConfig
		if err := parse.SaveSources(sourcesPath, sources); err != nil {
			return fmt.Errorf("save new source: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Created new source '%s' for %s\n", newConfig.Name, headerKey)
		config = &newConfig
	}

	transactions, err := parser.ParseRecords(config.Mapping, config.Name)
	if err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}

	imported := 0
	skipped := 0
	for _, txn := range transactions {
		_, err := database.InsertTransaction(txn.Source, txn.Date, txn.Description, txn.AmountCents, nil)
		if err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint") {
				skipped++
				continue
			}
			return fmt.Errorf("insert transaction: %w", err)
		}
		imported++
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Imported %d transactions from %s", imported, path)
	if skipped > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), " (%d duplicates skipped)", skipped)
	}
	fmt.Fprintln(cmd.OutOrStdout())

	engine := categorize.NewEngine(database)
	categorized, err := engine.AutoCategorizeTransactions()
	if err != nil {
		if mmErr, ok := err.(*categorize.MultiMatchError); ok {
			fmt.Fprintf(cmd.OutOrStdout(), "Warning: %d transactions matched multiple rules\n", len(mmErr.Warnings))
		} else {
			return fmt.Errorf("auto-categorize: %w", err)
		}
	}

	if len(categorized) > 0 {
		fmt.Fprintf(cmd.OutOrStdout(), "Auto-categorized %d transactions\n", len(categorized))
	}

	return nil
}
