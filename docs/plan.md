# BudgetBuddy Plan

## Project Layout

```
cmd/budgetbuddy/main.go          # Entry point
internal/
  cmd/                           # Cobra command definitions
    root.go                      # Root command, shared flags
    parse.go                     # parse subcommand
    report.go                    # report subcommand
    edit.go                      # edit subcommand
    config_category.go           # config category add/list/edit/delete
    config_rule.go               # config rule add/list/delete
    config_source.go             # config source list/delete
  db/                            # Database layer
    db.go                        # Connection management
    migrations.go                # Embedded SQL migrations
    models.go                    # Struct definitions
  parse/                         # CSV parsing
    parser.go                    # Core parser
    source.go                    # Auto-detect source from headers
  categorize/                    # Categorization engine
    categorizer.go               # Glob matching, rule resolution
  prompt/                        # Bubble Tea UI components
    category_selector.go         # Type-to-filter category picker
    source_wizard.go             # Interactive source creation wizard
  types/                         # Shared types/enums
    types.go
```

## Subcommands

| Command | Description |
|---|---|
| `parse [paths...] [--source name]` | Import CSV files (files or directories). Auto-detect source by header match. If unknown headers, launch interactive Bubble Tea wizard to create a new source. Auto-categorize via rules; Bubble Tea prompt for uncategorized. Warn on multi-rule matches. |
| `report --month M --year Y [--csv]` | Show combined report: line-item detail (ID, Date, Description, Amount, Category) sorted by date, then aggregated summary (Category, Budget, Actual, Diff) sorted alphabetically. With `--csv`: write `{month}-{year}-report.csv` to cwd. |
| `edit <id> --category "Name"` | Change a transaction's category. |
| `config category add/list/edit/delete` | Manage categories (name, description, monthly budget in dollars). |
| `config rule add/list/delete` | Manage glob patterns → category mappings. |
| `config source list/delete` | View/remove bank source configs. (No add — created automatically during parse.) |

## Database Schema (`~/.config/budgetbuddy/transactions.db`)

```sql
CREATE TABLE _migrations (
  version INTEGER PRIMARY KEY,
  applied_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE categories (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL UNIQUE,
  description TEXT NOT NULL DEFAULT '',
  monthly_budget_cents INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE categorization_rules (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  category_id INTEGER NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
  pattern TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE transactions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  source TEXT NOT NULL,
  date TEXT NOT NULL,
  description TEXT NOT NULL,
  amount_cents INTEGER NOT NULL,
  category_id INTEGER REFERENCES categories(id),
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  UNIQUE(date, description, amount_cents)
);
```

## Source Config (`~/.config/budgetbuddy/sources.yaml`)

Auto-created on first use. Each source keyed by its exact CSV header string:

Example source config:
```yaml
sources:
  "Item #,Card #,Transaction Date,Posting Date,Transaction Amount,Description":
    name: "BMO Credit Card"
    mapping:
      date:
        header: "Transaction Date"
        format: "20060102"
      description:
        header: "Description"
      amount:
        single_column: true
        is_positive_money_in: false
        header_out: "Transaction Amount"
        header_in: "Transaction Amount"

  "Date,Transaction Details,Funds Out,Funds In":
    name: "Simplii"
    mapping:
      date:
        header: "Date"
        format: "01/02/2006"
      description:
        header: "Transaction Details"
      amount:
        single_column: false
        is_positive_money_in: true
        header_out: "Funds Out"
        header_in: "Funds In"
```

Source matching: compare CSV header row string exactly against YAML keys. On first encounter of an unknown header combination, launch interactive wizard.

## Amount Normalization

All amounts normalized to signed integer cents:

- **Single column, `is_positive_money_in: true`**: `amount_cents = int(csv_value * 100)`
- **Single column, `is_positive_money_in: false`**: `amount_cents = -int(csv_value * 100)`
- **Dual column**: `amount_cents = int(in_value * 100) - int(out_value * 100)`

## Interactive Flows (Bubble Tea)

1. **Source creation wizard** (during `parse`): Detect headers → name → date column + format → description column → amount type (single/dual) → column mapping → sign convention → save

2. **Category selector** (during `parse`): For uncategorized transactions, type-to-filter fuzzy list of categories. Saves new rule automatically, should prompt users if they want to update the pattern to be a glob otherwise pattern = exact description.

## Design Decisions

| Area | Decision |
|---|---|
| Build system | Bazel (deferred config) |
| CLI framework | Cobra |
| Prompt library | Bubble Tea + Bubbles |
| Persistence | SQLite at `~/.config/budgetbuddy/transactions.db` |
| Source config | `~/.config/budgetbudget/sources.yaml` |
| Amount storage | Integer cents |
| Pattern matching | Shell globs |
| Duplicate detection | Skip on `(date, description, amount_cents)` |
| Rule conflicts | Warn at parse-time if multiple rules match |
| Migration strategy | Embedded SQL with version table |
| First-run | Auto-create dirs/DB/sources.yaml; prompt to add sources |
| Category sort (report) | Alphabetical |
| Report sections | Detail first, then summary |
| Transaction edit | `edit <id> --category "Name"` |
| Module path | `github.com/chrisvandoo/budgetbuddy` |
