package types_test

import (
	"testing"
	"time"

	"github.com/ChrisVandoo/budgetbuddy/internal/types"
)

func TestCategoryStruct(t *testing.T) {
	now := time.Now()
	c := types.Category{
		ID:                 1,
		Name:               "Groceries",
		Description:        "Food and household items",
		MonthlyBudgetCents: 50000,
		CreatedAt:          now,
	}
	if c.Name != "Groceries" {
		t.Errorf("expected Name Groceries, got %s", c.Name)
	}
	if c.MonthlyBudgetCents != 50000 {
		t.Errorf("expected MonthlyBudgetCents 50000, got %d", c.MonthlyBudgetCents)
	}
}

func TestTransactionStruct(t *testing.T) {
	now := time.Now()
	catID := int64(2)
	txn := types.Transaction{
		ID:          1,
		Source:      "BMO Credit Card",
		Date:        "2026-01-15",
		Description: "AMAZON.CA",
		AmountCents: -5000,
		CategoryID:  &catID,
		CreatedAt:   now,
	}
	if txn.Description != "AMAZON.CA" {
		t.Errorf("expected AMAZON.CA, got %s", txn.Description)
	}
	if *txn.CategoryID != 2 {
		t.Errorf("expected category ID 2, got %d", *txn.CategoryID)
	}
	if txn.AmountCents != -5000 {
		t.Errorf("expected -5000, got %d", txn.AmountCents)
	}
}

func TestCategorizationRuleStruct(t *testing.T) {
	r := types.CategorizationRule{
		ID:         1,
		CategoryID: 2,
		Pattern:    "*Amazon*",
	}
	if r.Pattern != "*Amazon*" {
		t.Errorf("expected pattern *Amazon*, got %s", r.Pattern)
	}
}

func TestSourceConfig(t *testing.T) {
	sc := types.SourceConfig{
		Name: "BMO Credit Card",
		Mapping: types.SourceMapping{
			Date: types.DateMapping{
				Header: "Transaction Date",
				Format: "20060102",
			},
			Description: types.DescriptionMapping{
				Header: "Description",
			},
			Amount: types.AmountMapping{
				SingleColumn:      true,
				IsPositiveMoneyIn: false,
				HeaderOut:         "Transaction Amount",
				HeaderIn:          "Transaction Amount",
			},
		},
	}
	if sc.Name != "BMO Credit Card" {
		t.Errorf("expected BMO Credit Card, got %s", sc.Name)
	}
	if !sc.Mapping.Amount.SingleColumn {
		t.Error("expected single column amount")
	}
}

func TestSourcesYAML(t *testing.T) {
	sy := types.SourcesYAML{
		Sources: map[string]types.SourceConfig{
			"col1,col2": {Name: "Test Source"},
		},
	}
	if len(sy.Sources) != 1 {
		t.Errorf("expected 1 source, got %d", len(sy.Sources))
	}
}

func TestConstants(t *testing.T) {
	if types.ConfigDirName != "budgetbuddy" {
		t.Errorf("unexpected ConfigDirName: %s", types.ConfigDirName)
	}
	if types.DBName != "transactions.db" {
		t.Errorf("unexpected DBName: %s", types.DBName)
	}
	if types.SourcesName != "sources.yaml" {
		t.Errorf("unexpected SourcesName: %s", types.SourcesName)
	}
}

func TestLineItem(t *testing.T) {
	li := types.LineItem{
		ID:          1,
		Date:        "2026-01-15",
		Description: "Test",
		AmountCents: -1000,
		Category:    "Food",
	}
	if li.Category != "Food" {
		t.Errorf("expected Food, got %s", li.Category)
	}
}

func TestCategorySummary(t *testing.T) {
	cs := types.CategorySummary{
		Category:       "Food",
		BudgetCents:    50000,
		ActualCents:    30000,
		RemainingCents: 20000,
	}
	if cs.RemainingCents != 20000 {
		t.Errorf("expected 20000, got %d", cs.RemainingCents)
	}
}
