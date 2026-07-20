package store

import "fmt"

type migration struct {
	name string
	sql  string
}

var migrations = []migration{
	{
		name: "users",
		sql: `CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			created_at TEXT NOT NULL
		)`,
	},
	{
		name: "workspaces",
		sql: `CREATE TABLE IF NOT EXISTS workspaces (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			owner_user_id INTEGER NOT NULL REFERENCES users(id),
			created_at TEXT NOT NULL
		)`,
	},
	// 先创建父表，再创建通过外键引用它们的子表。
	{
		name: "workspace_members",
		sql: `CREATE TABLE IF NOT EXISTS workspace_members (
			workspace_id INTEGER NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			role TEXT NOT NULL CHECK (role IN ('owner', 'editor', 'viewer')),
			PRIMARY KEY (workspace_id, user_id)
		)`,
	},
	{
		name: "categories",
		sql: `CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY,
			workspace_id INTEGER NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			name TEXT NOT NULL,
			created_at TEXT NOT NULL,
			UNIQUE (workspace_id, name)
		)`,
	},
	{
		name: "notes",
		sql: `CREATE TABLE IF NOT EXISTS notes (
			id INTEGER PRIMARY KEY,
			workspace_id INTEGER NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
			category_id INTEGER REFERENCES categories(id) ON DELETE SET NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL DEFAULT '',
			visibility TEXT NOT NULL DEFAULT 'private' CHECK (visibility IN ('private', 'public')),
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
	},
}

func (store *Store) migrate() error {
	transaction, err := store.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer transaction.Rollback()

	for _, migration := range migrations {
		if _, err := transaction.Exec(migration.sql); err != nil {
			return fmt.Errorf("create %s table: %w", migration.name, err)
		}
	}

	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
