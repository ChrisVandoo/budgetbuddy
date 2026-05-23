package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/ChrisVandoo/budgetbuddy/internal/categorize"
	"github.com/ChrisVandoo/budgetbuddy/internal/parse"
	"github.com/ChrisVandoo/budgetbuddy/internal/types"
)

func parseCmd() *cobra.Command {
	var sourceName string

	c := &cobra.Command{
		Use:   "parse [paths...]",
		Short: "Import CSV files",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				info, err := os.Stat(arg)
				if err != nil {
					return fmt.Errorf("access %s: %w", arg, err)
				}

				if info.IsDir() {
					entries, err := os.ReadDir(arg)
					if err != nil {
						return fmt.Errorf("read dir %s: %w", arg, err)
					}
					for _, entry := range entries {
						if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".csv") {
							if err := processCSVFile(cmd, arg+"/"+entry.Name(), sourceName); err != nil {
								return err
							}
						}
					}
				} else {
					if err := processCSVFile(cmd, arg, sourceName); err != nil {
						return err
					}
				}
			}
			return nil
		},
	}

	c.Flags().StringVar(&sourceName, "source", "", "Source name (skip auto-detection)")
	return c
}

func processCSVFile(cmd *cobra.Command, path, sourceName string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}

	headers, err := parse.ReadCSVHeaders(f)
	f.Close()
	if err != nil {
		return fmt.Errorf("read headers from %s: %w", path, err)
	}

	sources, err := parse.LoadSources(sourcesPath)
	if err != nil {
		return fmt.Errorf("load sources: %w", err)
	}

	srcName := sourceName
	var mapping types.SourceMapping

	if sourceName != "" {
		found := false
		for _, src := range sources.Sources {
			if src.Name == sourceName {
				mapping = src.Mapping
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("source %q not found in config", sourceName)
		}
	} else {
		_, config, found := parse.DetectSource(headers, sources)
		if !found {
			if len(sources.Sources) == 0 {
				return fmt.Errorf("no sources configured")
			}
			return fmt.Errorf("unknown headers in %s", path)
		}
		srcName = config.Name
		mapping = config.Mapping
	}

	f, err = os.Open(path)
	if err != nil {
		return fmt.Errorf("re-open %s: %w", path, err)
	}
	defer f.Close()

	parser := parse.NewParser(srcName, mapping)
	txns, err := parser.Parse(f)
	if err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}

	imported := 0
	skipped := 0
	for _, txn := range txns {
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
