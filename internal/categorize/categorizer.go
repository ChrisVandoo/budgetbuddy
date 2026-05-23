package categorize

import (
	"fmt"
	"path"
	"sort"

	"github.com/ChrisVandoo/budgetbuddy/internal/db"
	"github.com/ChrisVandoo/budgetbuddy/internal/types"
)

type Engine struct {
	db *db.DB
}

func NewEngine(database *db.DB) *Engine {
	return &Engine{db: database}
}

type MatchResult struct {
	TransactionID int64
	Description   string
	CategoryID    *int64
	CategoryName  string
	Pattern       string
}

func (e *Engine) Categorize(description string) ([]MatchResult, error) {
	rules, err := e.db.ListRules()
	if err != nil {
		return nil, fmt.Errorf("list rules: %w", err)
	}

	var matches []MatchResult
	for _, rule := range rules {
		matched, err := path.Match(rule.Pattern, description)
		if err != nil {
			continue
		}
		if matched {
			matches = append(matches, MatchResult{
				Description:  description,
				CategoryID:   &rule.CategoryID,
				CategoryName: rule.CategoryName,
				Pattern:      rule.Pattern,
			})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return len(matches[i].Pattern) > len(matches[j].Pattern)
	})

	return matches, nil
}

func (e *Engine) AutoCategorizeTransactions() ([]types.Transaction, error) {
	txns, err := e.db.GetUncategorizedTransactions()
	if err != nil {
		return nil, fmt.Errorf("get uncategorized: %w", err)
	}

	var categorized []types.Transaction
	multiWarnings := make(map[int64][]MatchResult)

	for _, txn := range txns {
		matches, err := e.Categorize(txn.Description)
		if err != nil {
			continue
		}

		if len(matches) == 1 {
			err := e.db.UpdateTransactionCategory(txn.ID, *matches[0].CategoryID)
			if err != nil {
				continue
			}
			txn.CategoryID = matches[0].CategoryID
			categorized = append(categorized, txn)
		} else if len(matches) > 1 {
			multiWarnings[txn.ID] = matches
		}
	}

	if len(multiWarnings) > 0 {
		return categorized, &MultiMatchError{Warnings: multiWarnings}
	}

	return categorized, nil
}

type MultiMatchError struct {
	Warnings map[int64][]MatchResult
}

func (e *MultiMatchError) Error() string {
	return fmt.Sprintf("%d transactions matched multiple rules", len(e.Warnings))
}

func (e *MultiMatchError) HasWarnings() bool {
	return len(e.Warnings) > 0
}
