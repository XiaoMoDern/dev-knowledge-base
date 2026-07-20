package store

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func Open(databasePath string) (*Store, error) {
	database, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, fmt.Errorf("open SQLite database: %w", err)
	}

	if err := database.Ping(); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("ping SQLite database: %w", err)
	}

	// SQLite 的 PRAGMA 设置绑定到单个连接；第一版限制为一个连接，
	// 确保每次操作都会开启外键检查。
	database.SetMaxOpenConns(1)
	if _, err := database.Exec(`PRAGMA foreign_keys = ON`); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("enable SQLite foreign keys: %w", err)
	}

	store := &Store{db: database}
	if err := store.migrate(); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("migrate SQLite database: %w", err)
	}
	if err := store.ensureDefaults(); err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("seed SQLite defaults: %w", err)
	}

	return store, nil
}

func (store *Store) Close() error {
	return store.db.Close()
}
