package prompt_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ChrisVandoo/budgetbuddy/internal/db"
	"github.com/ChrisVandoo/budgetbuddy/internal/prompt"
	"github.com/ChrisVandoo/budgetbuddy/internal/types"
)

func TestCategorySelectorCreated(t *testing.T) {
	dir, err := os.MkdirTemp("", "budgetbuddy-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	dbPath := filepath.Join(dir, "test.db")
	database, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer database.Close()

	txn := types.Transaction{
		ID:          1,
		Description: "AMAZON.CA",
		AmountCents: -5000,
	}

	selector := prompt.NewCategorySelector(database, txn)
	if selector == nil {
		t.Fatal("expected non-nil CategorySelector")
	}
}

func TestSourceWizardCreated(t *testing.T) {
	headers := []string{"Date", "Description", "Amount"}
	wizard := prompt.NewSourceWizard(headers, "")
	if wizard == nil {
		t.Fatal("expected non-nil SourceWizard")
	}
}
