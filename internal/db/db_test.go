package db_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ChrisVandoo/budgetbuddy/internal/db"
)

func setupTestDB(t *testing.T) (*db.DB, func()) {
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
	cleanup := func() {
		database.Close()
		os.RemoveAll(dir)
	}
	return database, cleanup
}

func TestOpenAndClose(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()
	if database == nil {
		t.Fatal("expected non-nil db")
	}
}

func TestCreateCategory(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	id, err := database.CreateCategory("Groceries", "Food items", 50000)
	if err != nil {
		t.Fatalf("CreateCategory failed: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero category ID")
	}
}

func TestCreateCategoryDuplicateName(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := database.CreateCategory("Groceries", "Food", 50000)
	if err != nil {
		t.Fatalf("first CreateCategory failed: %v", err)
	}
	_, err = database.CreateCategory("Groceries", "Food", 50000)
	if err == nil {
		t.Fatal("expected error for duplicate category name")
	}
}

func TestGetCategory(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	id, err := database.CreateCategory("Groceries", "Food items", 50000)
	if err != nil {
		t.Fatalf("CreateCategory failed: %v", err)
	}

	cat, err := database.GetCategory(id)
	if err != nil {
		t.Fatalf("GetCategory failed: %v", err)
	}
	if cat.Name != "Groceries" {
		t.Errorf("expected Groceries, got %s", cat.Name)
	}
	if cat.MonthlyBudgetCents != 50000 {
		t.Errorf("expected 50000, got %d", cat.MonthlyBudgetCents)
	}
}

func TestGetCategoryByName(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := database.CreateCategory("Groceries", "Food items", 50000)
	if err != nil {
		t.Fatalf("CreateCategory failed: %v", err)
	}

	cat, err := database.GetCategoryByName("Groceries")
	if err != nil {
		t.Fatalf("GetCategoryByName failed: %v", err)
	}
	if cat.Name != "Groceries" {
		t.Errorf("expected Groceries, got %s", cat.Name)
	}
}

func TestGetCategoryByNameNotFound(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	_, err := database.GetCategoryByName("NonExistent")
	if err == nil {
		t.Fatal("expected error for non-existent category")
	}
}

func TestListCategories(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	database.CreateCategory("Food", "Food items", 50000)
	database.CreateCategory("Transport", "Gas etc", 20000)

	cats, err := database.ListCategories()
	if err != nil {
		t.Fatalf("ListCategories failed: %v", err)
	}
	if len(cats) != 2 {
		t.Errorf("expected 2 categories, got %d", len(cats))
	}
}

func TestUpdateCategory(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	id, err := database.CreateCategory("Groceries", "Food", 50000)
	if err != nil {
		t.Fatalf("CreateCategory failed: %v", err)
	}

	err = database.UpdateCategory(id, "Groceries Updated", "Food items", 60000)
	if err != nil {
		t.Fatalf("UpdateCategory failed: %v", err)
	}

	cat, err := database.GetCategory(id)
	if err != nil {
		t.Fatalf("GetCategory failed: %v", err)
	}
	if cat.Name != "Groceries Updated" {
		t.Errorf("expected 'Groceries Updated', got '%s'", cat.Name)
	}
	if cat.MonthlyBudgetCents != 60000 {
		t.Errorf("expected 60000, got %d", cat.MonthlyBudgetCents)
	}
}

func TestDeleteCategory(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	id, err := database.CreateCategory("Groceries", "Food", 50000)
	if err != nil {
		t.Fatalf("CreateCategory failed: %v", err)
	}

	err = database.DeleteCategory(id)
	if err != nil {
		t.Fatalf("DeleteCategory failed: %v", err)
	}

	_, err = database.GetCategory(id)
	if err == nil {
		t.Fatal("expected error after deleting category")
	}
}

func TestCreateRule(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	catID, err := database.CreateCategory("Groceries", "Food", 50000)
	if err != nil {
		t.Fatalf("CreateCategory failed: %v", err)
	}

	ruleID, err := database.CreateRule(catID, "*Amazon*")
	if err != nil {
		t.Fatalf("CreateRule failed: %v", err)
	}
	if ruleID == 0 {
		t.Fatal("expected non-zero rule ID")
	}
}

func TestListRules(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	catID, _ := database.CreateCategory("Groceries", "Food", 50000)
	database.CreateRule(catID, "*Amazon*")
	database.CreateRule(catID, "*Walmart*")

	rules, err := database.ListRules()
	if err != nil {
		t.Fatalf("ListRules failed: %v", err)
	}
	if len(rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(rules))
	}
}

func TestDeleteRule(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	catID, _ := database.CreateCategory("Groceries", "Food", 50000)
	ruleID, _ := database.CreateRule(catID, "*Amazon*")

	err := database.DeleteRule(ruleID)
	if err != nil {
		t.Fatalf("DeleteRule failed: %v", err)
	}

	rules, _ := database.ListRules()
	if len(rules) != 0 {
		t.Errorf("expected 0 rules after delete, got %d", len(rules))
	}
}

func TestInsertTransaction(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	catID, _ := database.CreateCategory("Groceries", "Food", 50000)
	id, err := database.InsertTransaction("BMO", "2026-01-15", "AMAZON.CA", -5000, &catID)
	if err != nil {
		t.Fatalf("InsertTransaction failed: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero transaction ID")
	}
}

func TestInsertTransactionDuplicate(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	id1, err := database.InsertTransaction("BMO", "2026-01-15", "AMAZON.CA", -5000, nil)
	if err != nil {
		t.Fatalf("first InsertTransaction failed: %v", err)
	}
	if id1 == 0 {
		t.Fatal("expected non-zero id")
	}

	_, err = database.InsertTransaction("BMO", "2026-01-15", "AMAZON.CA", -5000, nil)
	if err == nil {
		t.Fatal("expected error for duplicate transaction")
	}
}

func TestListTransactions(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	database.InsertTransaction("BMO", "2026-01-15", "AMAZON.CA", -5000, nil)
	database.InsertTransaction("BMO", "2026-01-16", "WALMART", -3000, nil)

	txns, err := database.ListTransactions()
	if err != nil {
		t.Fatalf("ListTransactions failed: %v", err)
	}
	if len(txns) != 2 {
		t.Errorf("expected 2 transactions, got %d", len(txns))
	}
}

func TestGetTransaction(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	catID, _ := database.CreateCategory("Food", "Food", 50000)
	id, _ := database.InsertTransaction("BMO", "2026-01-15", "AMAZON.CA", -5000, &catID)

	txn, err := database.GetTransaction(id)
	if err != nil {
		t.Fatalf("GetTransaction failed: %v", err)
	}
	if txn.Description != "AMAZON.CA" {
		t.Errorf("expected AMAZON.CA, got %s", txn.Description)
	}
	if txn.CategoryID == nil || *txn.CategoryID != catID {
		t.Errorf("expected category %d", catID)
	}
}

func TestUpdateTransactionCategory(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	catID1, _ := database.CreateCategory("Food", "Food", 50000)
	catID2, _ := database.CreateCategory("Transport", "Transport", 20000)
	id, _ := database.InsertTransaction("BMO", "2026-01-15", "AMAZON.CA", -5000, &catID1)

	err := database.UpdateTransactionCategory(id, catID2)
	if err != nil {
		t.Fatalf("UpdateTransactionCategory failed: %v", err)
	}

	txn, _ := database.GetTransaction(id)
	if *txn.CategoryID != catID2 {
		t.Errorf("expected category %d, got %d", catID2, *txn.CategoryID)
	}
}

func TestGetTransactionsByMonth(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	catID, _ := database.CreateCategory("Food", "Food", 50000)
	database.InsertTransaction("BMO", "2026-01-15", "AMAZON.CA", -5000, &catID)
	database.InsertTransaction("BMO", "2026-01-20", "WALMART", -3000, &catID)
	database.InsertTransaction("BMO", "2026-02-10", "COSTCO", -10000, &catID)

	txns, err := database.GetTransactionsByMonth(1, 2026)
	if err != nil {
		t.Fatalf("GetTransactionsByMonth failed: %v", err)
	}
	if len(txns) != 2 {
		t.Errorf("expected 2 transactions in January, got %d", len(txns))
	}
}

func TestGetReportData(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	catID, _ := database.CreateCategory("Food", "Food items", 50000)
	database.InsertTransaction("BMO", "2026-01-15", "AMAZON.CA", -5000, &catID)
	database.InsertTransaction("BMO", "2026-01-20", "WALMART", -3000, &catID)

	items, summaries, err := database.GetReportData(1, 2026)
	if err != nil {
		t.Fatalf("GetReportData failed: %v", err)
	}
	if len(items) != 2 {
		t.Errorf("expected 2 line items, got %d", len(items))
	}
	if len(summaries) != 1 {
		t.Errorf("expected 1 summary, got %d", len(summaries))
	}
	if summaries[0].Category != "Food" {
		t.Errorf("expected Food category in summary, got %s", summaries[0].Category)
	}
	if summaries[0].ActualCents != -8000 {
		t.Errorf("expected actual -8000, got %d", summaries[0].ActualCents)
	}
}

func TestGetUncategorizedTransactions(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	database.InsertTransaction("BMO", "2026-01-15", "AMAZON.CA", -5000, nil)
	catID, _ := database.CreateCategory("Food", "Food", 50000)
	database.InsertTransaction("BMO", "2026-01-16", "WALMART", -3000, &catID)

	txns, err := database.GetUncategorizedTransactions()
	if err != nil {
		t.Fatalf("GetUncategorizedTransactions failed: %v", err)
	}
	if len(txns) != 1 {
		t.Errorf("expected 1 uncategorized transaction, got %d", len(txns))
	}
}

func TestMigrationsApplied(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()

	// Verify the _migrations table was populated
	var count int
	err := database.DB().QueryRow("SELECT COUNT(*) FROM _migrations").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query _migrations: %v", err)
	}
	if count == 0 {
		t.Fatal("expected at least 1 migration to be applied")
	}
}
