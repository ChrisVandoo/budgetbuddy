package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/ChrisVandoo/budgetbuddy/internal/db"
)

var (
	database    *db.DB
	sourcesPath string
)

func SetDB(d *db.DB) {
	database = d
}

func SetSourcesPath(path string) {
	sourcesPath = path
}

func RootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "budgetbuddy",
		Short: "A friendly budgeting utility",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Name() == "help" || cmd.Name() == "completion" {
				return nil
			}
			return initDB()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	root.AddCommand(parseCmd())
	root.AddCommand(reportCmd())
	root.AddCommand(editCmd())
	root.AddCommand(configCmd())

	return root
}

func Execute() {
	root := RootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initDB() error {
	if database != nil {
		return nil
	}

	configDir, err := db.DefaultConfigDir()
	if err != nil {
		return fmt.Errorf("config dir: %w", err)
	}

	if sourcesPath == "" {
		sourcesPath = configDir + "/sources.yaml"
	}

	dbPath, err := db.DefaultDBPath()
	if err != nil {
		return fmt.Errorf("db path: %w", err)
	}

	d, err := db.Open(dbPath)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	database = d
	return nil
}
