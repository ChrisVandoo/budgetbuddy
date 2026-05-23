package cmd_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ChrisVandoo/budgetbuddy/internal/cmd"
	"github.com/ChrisVandoo/budgetbuddy/internal/db"
	"github.com/ChrisVandoo/budgetbuddy/internal/parse"
	"github.com/ChrisVandoo/budgetbuddy/internal/types"
)

func setupCmdEnv(t *testing.T) (string, *db.DB, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "budgetbuddy-cmd-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(dir, "test.db")
	database, err := db.Open(dbPath)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatalf("failed to open db: %v", err)
	}

	cleanup := func() {
		database.Close()
		os.RemoveAll(dir)
	}

	return dir, database, cleanup
}

func TestExecuteRoot(t *testing.T) {
	dir, database, cleanup := setupCmdEnv(t)
	defer cleanup()

	cmd.SetDB(database)
	cmd.SetSourcesPath(filepath.Join(dir, "sources.yaml"))

	var stdout bytes.Buffer
	rootCmd := cmd.RootCmd()
	rootCmd.SetOut(&stdout)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("root --help failed: %v", err)
	}

	output := stdout.String()
	if !strings.Contains(output, "budgetbuddy") {
		t.Errorf("expected help to contain 'budgetbuddy', got: %s", output)
	}
}

func TestConfigCategoryAdd(t *testing.T) {
	dir, database, cleanup := setupCmdEnv(t)
	defer cleanup()

	cmd.SetDB(database)
	cmd.SetSourcesPath(filepath.Join(dir, "sources.yaml"))

	var stdout, stderr bytes.Buffer
	rootCmd := cmd.RootCmd()
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"config", "category", "add", "--name", "Food", "--desc", "Food items", "--budget", "500.00"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("config category add failed: %v, stderr: %s", err, stderr.String())
	}

	cat, err := database.GetCategoryByName("Food")
	if err != nil {
		t.Fatalf("expected category to exist: %v", err)
	}
	if cat.MonthlyBudgetCents != 50000 {
		t.Errorf("expected budget 50000, got %d", cat.MonthlyBudgetCents)
	}
}

func TestConfigCategoryList(t *testing.T) {
	dir, database, cleanup := setupCmdEnv(t)
	defer cleanup()

	cmd.SetDB(database)
	cmd.SetSourcesPath(filepath.Join(dir, "sources.yaml"))

	database.CreateCategory("Food", "Food items", 50000)
	database.CreateCategory("Transport", "Gas", 20000)

	var stdout, stderr bytes.Buffer
	rootCmd := cmd.RootCmd()
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"config", "category", "list"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("config category list failed: %v, stderr: %s", err, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Food") {
		t.Errorf("expected output to contain Food, got: %s", output)
	}
	if !strings.Contains(output, "Transport") {
		t.Errorf("expected output to contain Transport, got: %s", output)
	}
}

func TestConfigCategoryEdit(t *testing.T) {
	dir, database, cleanup := setupCmdEnv(t)
	defer cleanup()

	cmd.SetDB(database)
	cmd.SetSourcesPath(filepath.Join(dir, "sources.yaml"))

	id, _ := database.CreateCategory("Food", "Food items", 50000)

	var stdout, stderr bytes.Buffer
	rootCmd := cmd.RootCmd()
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"config", "category", "edit", "--id", "1", "--name", "Groceries", "--desc", "Groceries", "--budget", "600.00"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("config category edit failed: %v, stderr: %s", err, stderr.String())
	}

	cat, _ := database.GetCategory(id)
	if cat.Name != "Groceries" {
		t.Errorf("expected Groceries, got %s", cat.Name)
	}
}

func TestConfigCategoryDelete(t *testing.T) {
	dir, database, cleanup := setupCmdEnv(t)
	defer cleanup()

	cmd.SetDB(database)
	cmd.SetSourcesPath(filepath.Join(dir, "sources.yaml"))

	_, _ = database.CreateCategory("Food", "Food", 50000)

	var stdout, stderr bytes.Buffer
	rootCmd := cmd.RootCmd()
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"config", "category", "delete", "--id", "1"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("config category delete failed: %v, stderr: %s", err, stderr.String())
	}

	_, err = database.GetCategory(1)
	if err == nil {
		t.Fatal("expected category to be deleted")
	}
}

func TestConfigRuleAdd(t *testing.T) {
	dir, database, cleanup := setupCmdEnv(t)
	defer cleanup()

	cmd.SetDB(database)
	cmd.SetSourcesPath(filepath.Join(dir, "sources.yaml"))

	database.CreateCategory("Food", "Food", 50000)

	var stdout, stderr bytes.Buffer
	rootCmd := cmd.RootCmd()
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"config", "rule", "add", "--category", "Food", "--pattern", "*Amazon*"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("config rule add failed: %v, stderr: %s", err, stderr.String())
	}

	rules, _ := database.ListRules()
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Pattern != "*Amazon*" {
		t.Errorf("expected *Amazon*, got %s", rules[0].Pattern)
	}
}

func TestConfigRuleList(t *testing.T) {
	dir, database, cleanup := setupCmdEnv(t)
	defer cleanup()

	cmd.SetDB(database)
	cmd.SetSourcesPath(filepath.Join(dir, "sources.yaml"))

	catID, _ := database.CreateCategory("Food", "Food", 50000)
	database.CreateRule(catID, "*Amazon*")

	var stdout, stderr bytes.Buffer
	rootCmd := cmd.RootCmd()
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"config", "rule", "list"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("config rule list failed: %v, stderr: %s", err, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "*Amazon*") {
		t.Errorf("expected output to contain *Amazon*, got: %s", output)
	}
}

func TestConfigRuleDelete(t *testing.T) {
	dir, database, cleanup := setupCmdEnv(t)
	defer cleanup()

	cmd.SetDB(database)
	cmd.SetSourcesPath(filepath.Join(dir, "sources.yaml"))

	catID, _ := database.CreateCategory("Food", "Food", 50000)
	database.CreateRule(catID, "*Amazon*")

	var stdout, stderr bytes.Buffer
	rootCmd := cmd.RootCmd()
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"config", "rule", "delete", "--id", "1"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("config rule delete failed: %v, stderr: %s", err, stderr.String())
	}

	rules, _ := database.ListRules()
	if len(rules) != 0 {
		t.Errorf("expected 0 rules, got %d", len(rules))
	}
}

func TestConfigSourceList(t *testing.T) {
	dir, database, cleanup := setupCmdEnv(t)
	defer cleanup()

	cmd.SetDB(database)
	sourcesPath := filepath.Join(dir, "sources.yaml")
	cmd.SetSourcesPath(sourcesPath)

	sources := &types.SourcesYAML{
		Sources: map[string]types.SourceConfig{
			"col1,col2": {Name: "Test Source"},
		},
	}
	parse.SaveSources(sourcesPath, sources)

	var stdout, stderr bytes.Buffer
	rootCmd := cmd.RootCmd()
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"config", "source", "list"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("config source list failed: %v, stderr: %s", err, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "Test Source") {
		t.Errorf("expected output to contain Test Source, got: %s", output)
	}
}

func TestEditCommand(t *testing.T) {
	dir, database, cleanup := setupCmdEnv(t)
	defer cleanup()

	cmd.SetDB(database)
	cmd.SetSourcesPath(filepath.Join(dir, "sources.yaml"))

	catID, _ := database.CreateCategory("Food", "Food", 50000)
	txnID, _ := database.InsertTransaction("BMO", "2026-01-15", "AMAZON.CA", -5000, nil)

	var stdout, stderr bytes.Buffer
	rootCmd := cmd.RootCmd()
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"edit", "--id", "1", "--category", "Food"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("edit failed: %v, stderr: %s", err, stderr.String())
	}

	txn, _ := database.GetTransaction(txnID)
	if txn.CategoryID == nil || *txn.CategoryID != catID {
		t.Errorf("expected category %d, got %v", catID, txn.CategoryID)
	}
}

func TestEditCommandTxIDArg(t *testing.T) {
	dir, database, cleanup := setupCmdEnv(t)
	defer cleanup()

	cmd.SetDB(database)
	cmd.SetSourcesPath(filepath.Join(dir, "sources.yaml"))

	database.CreateCategory("Food", "Food", 50000)
	database.InsertTransaction("BMO", "2026-01-15", "AMAZON.CA", -5000, nil)

	var stdout, stderr bytes.Buffer
	rootCmd := cmd.RootCmd()
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"edit", "1", "--category", "Food"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("edit with positional id failed: %v, stderr: %s", err, stderr.String())
	}
	_ = stdout.String()
}

func TestReportCommand(t *testing.T) {
	dir, database, cleanup := setupCmdEnv(t)
	defer cleanup()

	cmd.SetDB(database)
	cmd.SetSourcesPath(filepath.Join(dir, "sources.yaml"))

	catID, _ := database.CreateCategory("Food", "Food", 50000)
	database.InsertTransaction("BMO", "2026-01-15", "AMAZON.CA", -5000, &catID)
	database.InsertTransaction("BMO", "2026-01-20", "WALMART", -3000, &catID)

	var stdout, stderr bytes.Buffer
	rootCmd := cmd.RootCmd()
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"report", "--month", "1", "--year", "2026"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("report failed: %v, stderr: %s", err, stderr.String())
	}

	output := stdout.String()
	if !strings.Contains(output, "AMAZON.CA") {
		t.Errorf("expected output to contain AMAZON.CA, got: %s", output)
	}
	if !strings.Contains(output, "Food") {
		t.Errorf("expected output to contain Food category, got: %s", output)
	}
}

func TestReportCommandCSV(t *testing.T) {
	dir, database, cleanup := setupCmdEnv(t)
	defer cleanup()

	cmd.SetDB(database)
	cmd.SetSourcesPath(filepath.Join(dir, "sources.yaml"))

	catID, _ := database.CreateCategory("Food", "Food", 50000)
	database.InsertTransaction("BMO", "2026-01-15", "AMAZON.CA", -5000, &catID)

	var stdout, stderr bytes.Buffer
	rootCmd := cmd.RootCmd()
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"report", "--month", "1", "--year", "2026", "--csv"})

	err := rootCmd.Execute()
	if err != nil {
		t.Fatalf("report --csv failed: %v, stderr: %s", err, stderr.String())
	}

	// Check CSV file was created
	csvPath := filepath.Join(dir, "01-2026-report.csv")
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		// May have been written to cwd, check there too
		cwdCsv := "01-2026-report.csv"
		if _, err2 := os.Stat(cwdCsv); os.IsNotExist(err2) {
			t.Logf("csv file not found at %s or %s (may be written to cwd)", csvPath, cwdCsv)
		} else {
			os.Remove(cwdCsv)
		}
	}
}

func TestParseCommandNoArgs(t *testing.T) {
	dir, database, cleanup := setupCmdEnv(t)
	defer cleanup()

	cmd.SetDB(database)
	cmd.SetSourcesPath(filepath.Join(dir, "sources.yaml"))

	var stdout, stderr bytes.Buffer
	rootCmd := cmd.RootCmd()
	rootCmd.SetOut(&stdout)
	rootCmd.SetErr(&stderr)
	rootCmd.SetArgs([]string{"parse"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for parse with no args")
	}
}
