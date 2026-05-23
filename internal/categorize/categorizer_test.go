package categorize_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ChrisVandoo/budgetbuddy/internal/categorize"
	"github.com/ChrisVandoo/budgetbuddy/internal/db"
)

func setupTest(t *testing.T) (*categorize.Engine, *db.DB, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "budgetbuddy-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	dbPath := filepath.Join(dir, "test.db")
	database, err := db.Open(dbPath)
	if err != nil {
		os.RemoveAll(dir)
		t.Fatalf("failed to open db: %v", err)
	}

	engine := categorize.NewEngine(database)
	cleanup := func() {
		database.Close()
		os.RemoveAll(dir)
	}
	return engine, database, cleanup
}

func TestExactMatch(t *testing.T) {
	engine, database, cleanup := setupTest(t)
	defer cleanup()

	catID, _ := database.CreateCategory("Food", "Food items", 50000)
	database.CreateRule(catID, "AMAZON.CA")

	matches, err := engine.Categorize("AMAZON.CA")
	if err != nil {
		t.Fatalf("Categorize failed: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	if *matches[0].CategoryID != catID {
		t.Errorf("expected category %d, got %d", catID, *matches[0].CategoryID)
	}
}

func TestGlobMatch(t *testing.T) {
	engine, database, cleanup := setupTest(t)
	defer cleanup()

	catID, _ := database.CreateCategory("Food", "Food items", 50000)
	database.CreateRule(catID, "*Amazon*")

	matches, err := engine.Categorize("AMAZON.CA")
	if err != nil {
		t.Fatalf("Categorize failed: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("expected 0 matches (case-sensitive), got %d", len(matches))
	}
}

func TestGlobMatchCaseInsensitive(t *testing.T) {
	engine, database, cleanup := setupTest(t)
	defer cleanup()

	catID, _ := database.CreateCategory("Food", "Food items", 50000)
	database.CreateRule(catID, "*Amazon*")

	matches, err := engine.Categorize("Amazon.com")
	if err != nil {
		t.Fatalf("Categorize failed: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
}

func TestWildcardMatch(t *testing.T) {
	engine, database, cleanup := setupTest(t)
	defer cleanup()

	catID, _ := database.CreateCategory("Food", "Food items", 50000)
	database.CreateRule(catID, "*")

	matches, err := engine.Categorize("ANYTHING")
	if err != nil {
		t.Fatalf("Categorize failed: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
}

func TestNoMatch(t *testing.T) {
	engine, database, cleanup := setupTest(t)
	defer cleanup()

	catID, _ := database.CreateCategory("Food", "Food items", 50000)
	database.CreateRule(catID, "*Amazon*")

	matches, err := engine.Categorize("WALMART")
	if err != nil {
		t.Fatalf("Categorize failed: %v", err)
	}
	if len(matches) != 0 {
		t.Fatalf("expected 0 matches, got %d", len(matches))
	}
}

func TestMultipleMatches(t *testing.T) {
	engine, database, cleanup := setupTest(t)
	defer cleanup()

	catID1, _ := database.CreateCategory("Food", "Food items", 50000)
	catID2, _ := database.CreateCategory("Shopping", "Shopping", 30000)
	database.CreateRule(catID1, "*Amazon*")
	database.CreateRule(catID2, "*Amazon*")

	matches, err := engine.Categorize("Amazon.com")
	if err != nil {
		t.Fatalf("Categorize failed: %v", err)
	}
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
}

func TestAutoCategorize(t *testing.T) {
	engine, database, cleanup := setupTest(t)
	defer cleanup()

	catID, _ := database.CreateCategory("Food", "Food items", 50000)
	database.CreateRule(catID, "*Amazon*")

	database.InsertTransaction("BMO", "2026-01-15", "Amazon.com", -5000, nil)
	database.InsertTransaction("BMO", "2026-01-16", "WALMART", -3000, nil)

	categorized, err := engine.AutoCategorizeTransactions()
	if err != nil {
		t.Fatalf("AutoCategorizeTransactions failed: %v", err)
	}
	if len(categorized) != 1 {
		t.Fatalf("expected 1 categorized transaction, got %d", len(categorized))
	}
	if categorized[0].Description != "Amazon.com" {
		t.Errorf("expected Amazon.com, got %s", categorized[0].Description)
	}
}

func TestAutoCategorizeNoRules(t *testing.T) {
	engine, database, cleanup := setupTest(t)
	defer cleanup()

	database.InsertTransaction("BMO", "2026-01-15", "AMAZON.CA", -5000, nil)

	categorized, err := engine.AutoCategorizeTransactions()
	if err != nil {
		t.Fatalf("AutoCategorizeTransactions failed: %v", err)
	}
	if len(categorized) != 0 {
		t.Fatalf("expected 0 categorized, got %d", len(categorized))
	}
}

func TestMultiMatchWarning(t *testing.T) {
	engine, database, cleanup := setupTest(t)
	defer cleanup()

	catID1, _ := database.CreateCategory("Food", "Food", 50000)
	catID2, _ := database.CreateCategory("Shopping", "Shopping", 30000)
	database.CreateRule(catID1, "*Amazon*")
	database.CreateRule(catID2, "*Amazon*")

	id, _ := database.InsertTransaction("BMO", "2026-01-15", "Amazon.com", -5000, nil)

	_, err := engine.AutoCategorizeTransactions()
	if err == nil {
		t.Fatal("expected MultiMatchError for multiple matches")
	}

	mmErr, ok := err.(*categorize.MultiMatchError)
	if !ok {
		t.Fatalf("expected MultiMatchError, got %T", err)
	}
	if !mmErr.HasWarnings() {
		t.Error("expected HasWarnings to be true")
	}
	if _, exists := mmErr.Warnings[id]; !exists {
		t.Errorf("expected transaction %d in warnings", id)
	}
}
