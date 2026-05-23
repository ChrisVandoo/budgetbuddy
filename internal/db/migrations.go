package db

import "fmt"

type migration struct {
	version int
	sql     string
}

var migrations = []migration{
	{
		version: 1,
		sql: `CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			description TEXT NOT NULL DEFAULT '',
			monthly_budget_cents INTEGER NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		);`,
	},
	{
		version: 2,
		sql: `CREATE TABLE IF NOT EXISTS categorization_rules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			category_id INTEGER NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
			pattern TEXT NOT NULL,
			created_at TEXT NOT NULL DEFAULT (datetime('now'))
		);`,
	},
	{
		version: 3,
		sql: `CREATE TABLE IF NOT EXISTS transactions (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			source TEXT NOT NULL,
			date TEXT NOT NULL,
			description TEXT NOT NULL,
			amount_cents INTEGER NOT NULL,
			category_id INTEGER REFERENCES categories(id),
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			UNIQUE(date, description, amount_cents)
		);`,
	},
}

func (d *DB) runMigrations() error {
	// Bootstrap: ensure _migrations table exists before checking it
	_, err := d.db.Exec(`CREATE TABLE IF NOT EXISTS _migrations (
		version INTEGER PRIMARY KEY,
		applied_at TEXT NOT NULL DEFAULT (datetime('now'))
	)`)
	if err != nil {
		return fmt.Errorf("bootstrap migrations table: %w", err)
	}

	for _, m := range migrations {
		var count int
		err := d.db.QueryRow("SELECT COUNT(*) FROM _migrations WHERE version = ?", m.version).Scan(&count)
		if err != nil {
			return fmt.Errorf("check migration %d: %w", m.version, err)
		}
		if count > 0 {
			continue
		}

		_, err = d.db.Exec(m.sql)
		if err != nil {
			return fmt.Errorf("apply migration %d: %w", m.version, err)
		}

		_, err = d.db.Exec("INSERT INTO _migrations (version) VALUES (?)", m.version)
		if err != nil {
			return fmt.Errorf("record migration %d: %w", m.version, err)
		}
	}
	return nil
}
