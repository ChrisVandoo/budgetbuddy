package db

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/ChrisVandoo/budgetbuddy/internal/types"
)

func (d *DB) CreateCategory(name, description string, budgetCents int64) (int64, error) {
	res, err := d.db.Exec(
		"INSERT INTO categories (name, description, monthly_budget_cents) VALUES (?, ?, ?)",
		name, description, budgetCents,
	)
	if err != nil {
		return 0, fmt.Errorf("create category: %w", err)
	}
	return res.LastInsertId()
}

func (d *DB) GetCategory(id int64) (*types.Category, error) {
	row := d.db.QueryRow("SELECT id, name, description, monthly_budget_cents, created_at FROM categories WHERE id = ?", id)
	return scanCategory(row)
}

func (d *DB) GetCategoryByName(name string) (*types.Category, error) {
	row := d.db.QueryRow("SELECT id, name, description, monthly_budget_cents, created_at FROM categories WHERE name = ?", name)
	return scanCategory(row)
}

func scanCategory(scanner interface {
	Scan(dest ...interface{}) error
}) (*types.Category, error) {
	var cat types.Category
	var createdAt string
	err := scanner.Scan(&cat.ID, &cat.Name, &cat.Description, &cat.MonthlyBudgetCents, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("scan category: %w", err)
	}
	return &cat, nil
}

func (d *DB) ListCategories() ([]types.Category, error) {
	rows, err := d.db.Query("SELECT id, name, description, monthly_budget_cents, created_at FROM categories ORDER BY name")
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	var cats []types.Category
	for rows.Next() {
		var cat types.Category
		var createdAt string
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Description, &cat.MonthlyBudgetCents, &createdAt); err != nil {
			return nil, fmt.Errorf("scan category row: %w", err)
		}
		cats = append(cats, cat)
	}
	return cats, rows.Err()
}

func (d *DB) UpdateCategory(id int64, name, description string, budgetCents int64) error {
	res, err := d.db.Exec(
		"UPDATE categories SET name = ?, description = ?, monthly_budget_cents = ? WHERE id = ?",
		name, description, budgetCents, id,
	)
	if err != nil {
		return fmt.Errorf("update category: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("category not found")
	}
	return nil
}

func (d *DB) DeleteCategory(id int64) error {
	res, err := d.db.Exec("DELETE FROM categories WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("category not found")
	}
	return nil
}

func (d *DB) CreateRule(categoryID int64, pattern string) (int64, error) {
	res, err := d.db.Exec(
		"INSERT INTO categorization_rules (category_id, pattern) VALUES (?, ?)",
		categoryID, pattern,
	)
	if err != nil {
		return 0, fmt.Errorf("create rule: %w", err)
	}
	return res.LastInsertId()
}

type RuleWithCategory struct {
	types.CategorizationRule
	CategoryName string
}

func (d *DB) ListRules() ([]RuleWithCategory, error) {
	rows, err := d.db.Query(`
		SELECT r.id, r.category_id, r.pattern, r.created_at, c.name
		FROM categorization_rules r
		JOIN categories c ON c.id = r.category_id
		ORDER BY r.id
	`)
	if err != nil {
		return nil, fmt.Errorf("list rules: %w", err)
	}
	defer rows.Close()

	var rules []RuleWithCategory
	for rows.Next() {
		var r RuleWithCategory
		var createdAt string
		if err := rows.Scan(&r.ID, &r.CategoryID, &r.Pattern, &createdAt, &r.CategoryName); err != nil {
			return nil, fmt.Errorf("scan rule row: %w", err)
		}
		rules = append(rules, r)
	}
	return rules, rows.Err()
}

func (d *DB) DeleteRule(id int64) error {
	res, err := d.db.Exec("DELETE FROM categorization_rules WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete rule: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("rule not found")
	}
	return nil
}

func (d *DB) InsertTransaction(source, date, description string, amountCents int64, categoryID *int64) (int64, error) {
	res, err := d.db.Exec(
		"INSERT INTO transactions (source, date, description, amount_cents, category_id) VALUES (?, ?, ?, ?, ?)",
		source, date, description, amountCents, categoryID,
	)
	if err != nil {
		return 0, fmt.Errorf("insert transaction: %w", err)
	}
	return res.LastInsertId()
}

func scanTransaction(scanner interface {
	Scan(dest ...interface{}) error
}) (*types.Transaction, error) {
	var txn types.Transaction
	var createdAt string
	err := scanner.Scan(&txn.ID, &txn.Source, &txn.Date, &txn.Description, &txn.AmountCents, &txn.CategoryID, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("transaction not found")
		}
		return nil, fmt.Errorf("scan transaction: %w", err)
	}
	return &txn, nil
}

func (d *DB) GetTransaction(id int64) (*types.Transaction, error) {
	row := d.db.QueryRow(
		"SELECT id, source, date, description, amount_cents, category_id, created_at FROM transactions WHERE id = ?",
		id,
	)
	return scanTransaction(row)
}

func (d *DB) ListTransactions() ([]types.Transaction, error) {
	rows, err := d.db.Query("SELECT id, source, date, description, amount_cents, category_id, created_at FROM transactions ORDER BY date")
	if err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}
	defer rows.Close()

	var txns []types.Transaction
	for rows.Next() {
		var txn types.Transaction
		var createdAt string
		if err := rows.Scan(&txn.ID, &txn.Source, &txn.Date, &txn.Description, &txn.AmountCents, &txn.CategoryID, &createdAt); err != nil {
			return nil, fmt.Errorf("scan transaction row: %w", err)
		}
		txns = append(txns, txn)
	}
	return txns, rows.Err()
}

func (d *DB) UpdateTransactionCategory(id, categoryID int64) error {
	res, err := d.db.Exec("UPDATE transactions SET category_id = ? WHERE id = ?", categoryID, id)
	if err != nil {
		return fmt.Errorf("update transaction category: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("transaction not found")
	}
	return nil
}

func (d *DB) GetTransactionsByMonth(month, year int) ([]types.Transaction, error) {
	prefix := fmt.Sprintf("%04d-%02d", year, month)
	rows, err := d.db.Query(
		"SELECT id, source, date, description, amount_cents, category_id, created_at FROM transactions WHERE date LIKE ? ORDER BY date",
		prefix+"%",
	)
	if err != nil {
		return nil, fmt.Errorf("get transactions by month: %w", err)
	}
	defer rows.Close()

	var txns []types.Transaction
	for rows.Next() {
		var txn types.Transaction
		var createdAt string
		if err := rows.Scan(&txn.ID, &txn.Source, &txn.Date, &txn.Description, &txn.AmountCents, &txn.CategoryID, &createdAt); err != nil {
			return nil, fmt.Errorf("scan transaction row: %w", err)
		}
		txns = append(txns, txn)
	}
	return txns, rows.Err()
}

func (d *DB) GetReportData(month, year int) ([]types.LineItem, []types.CategorySummary, error) {
	txns, err := d.GetTransactionsByMonth(month, year)
	if err != nil {
		return nil, nil, err
	}

	var items []types.LineItem
	summaryMap := make(map[string]*types.CategorySummary)

	for _, txn := range txns {
		catName := ""
		if txn.CategoryID != nil {
			cat, err := d.GetCategory(*txn.CategoryID)
			if err == nil {
				catName = cat.Name
			}
		}

		items = append(items, types.LineItem{
			ID:          txn.ID,
			Date:        txn.Date,
			Description: txn.Description,
			AmountCents: txn.AmountCents,
			Category:    catName,
		})

		if catName != "" {
			if _, ok := summaryMap[catName]; !ok {
				cat, err := d.GetCategoryByName(catName)
				budget := int64(0)
				if err == nil {
					budget = cat.MonthlyBudgetCents
				}
				summaryMap[catName] = &types.CategorySummary{
					Category:    catName,
					BudgetCents: budget,
				}
			}
			summaryMap[catName].ActualCents += txn.AmountCents
		}
	}

	var summaries []types.CategorySummary
	for _, name := range sortedKeys(summaryMap) {
		s := summaryMap[name]
		s.RemainingCents = s.BudgetCents + s.ActualCents
		summaries = append(summaries, *s)
	}

	return items, summaries, nil
}

func sortedKeys(m map[string]*types.CategorySummary) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sortStrings(keys)
	return keys
}

func sortStrings(s []string) {
	for i := 0; i < len(s); i++ {
		for j := i + 1; j < len(s); j++ {
			if strings.ToLower(s[i]) > strings.ToLower(s[j]) {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

func (d *DB) GetUncategorizedTransactions() ([]types.Transaction, error) {
	rows, err := d.db.Query(
		"SELECT id, source, date, description, amount_cents, category_id, created_at FROM transactions WHERE category_id IS NULL ORDER BY date",
	)
	if err != nil {
		return nil, fmt.Errorf("get uncategorized: %w", err)
	}
	defer rows.Close()

	var txns []types.Transaction
	for rows.Next() {
		var txn types.Transaction
		var createdAt string
		if err := rows.Scan(&txn.ID, &txn.Source, &txn.Date, &txn.Description, &txn.AmountCents, &txn.CategoryID, &createdAt); err != nil {
			return nil, fmt.Errorf("scan transaction row: %w", err)
		}
		txns = append(txns, txn)
	}
	return txns, rows.Err()
}
