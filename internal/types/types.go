package types

import "time"

type Category struct {
	ID                 int64     `json:"id" yaml:"id"`
	Name               string    `json:"name" yaml:"name"`
	Description        string    `json:"description" yaml:"description"`
	MonthlyBudgetCents int64     `json:"monthly_budget_cents" yaml:"monthly_budget_cents"`
	CreatedAt          time.Time `json:"created_at" yaml:"created_at"`
}

type CategorizationRule struct {
	ID         int64     `json:"id" yaml:"id"`
	CategoryID int64     `json:"category_id" yaml:"category_id"`
	Pattern    string    `json:"pattern" yaml:"pattern"`
	CreatedAt  time.Time `json:"created_at" yaml:"created_at"`
}

type Transaction struct {
	ID          int64     `json:"id" yaml:"id"`
	Source      string    `json:"source" yaml:"source"`
	Date        string    `json:"date" yaml:"date"`
	Description string    `json:"description" yaml:"description"`
	AmountCents int64     `json:"amount_cents" yaml:"amount_cents"`
	CategoryID  *int64    `json:"category_id" yaml:"category_id"`
	CreatedAt   time.Time `json:"created_at" yaml:"created_at"`
}

type DateMapping struct {
	Header string `yaml:"header"`
	Format string `yaml:"format"`
}

type DescriptionMapping struct {
	Header string `yaml:"header"`
}

type AmountMapping struct {
	SingleColumn      bool   `yaml:"single_column"`
	IsPositiveMoneyIn bool   `yaml:"is_positive_money_in"`
	HeaderOut         string `yaml:"header_out"`
	HeaderIn          string `yaml:"header_in"`
}

type SourceMapping struct {
	Date        DateMapping        `yaml:"date"`
	Description DescriptionMapping `yaml:"description"`
	Amount      AmountMapping      `yaml:"amount"`
}

type SourceConfig struct {
	Name    string        `yaml:"name"`
	Mapping SourceMapping `yaml:"mapping"`
}

type SourcesYAML struct {
	Sources map[string]SourceConfig `yaml:"sources"`
}

type LineItem struct {
	ID          int64
	Date        string
	Description string
	AmountCents int64
	Category    string
}

type CategorySummary struct {
	Category         string
	BudgetCents      int64
	ActualCents      int64
	RemainingCents   int64
}

const ConfigDirName = "budgetbuddy"
const DBName = "transactions.db"
const SourcesName = "sources.yaml"
