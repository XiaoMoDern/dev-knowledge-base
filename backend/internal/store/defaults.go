package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

const (
	defaultUsername      = "local"
	defaultWorkspaceName = "我的知识库"
)

func (store *Store) ensureDefaults() error {
	transaction, err := store.db.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer transaction.Rollback()

	createdAt := time.Now().UTC().Format(time.RFC3339)

	// username 有唯一约束，重复打开同一数据库不会创建第二个本地用户。
	if _, err := transaction.Exec(`
		INSERT INTO users (username, created_at)
		VALUES (?, ?)
		ON CONFLICT (username) DO NOTHING
	`, defaultUsername, createdAt); err != nil {
		return fmt.Errorf("create local user: %w", err)
	}

	var userID int64
	if err := transaction.QueryRow(`SELECT id FROM users WHERE username = ?`, defaultUsername).Scan(&userID); err != nil {
		return fmt.Errorf("find local user: %w", err)
	}

	var workspaceID int64
	err = transaction.QueryRow(`
		SELECT id FROM workspaces
		WHERE owner_user_id = ? AND name = ?
		LIMIT 1
	`, userID, defaultWorkspaceName).Scan(&workspaceID)
	if errors.Is(err, sql.ErrNoRows) {
		result, err := transaction.Exec(`
			INSERT INTO workspaces (name, owner_user_id, created_at)
			VALUES (?, ?, ?)
		`, defaultWorkspaceName, userID, createdAt)
		if err != nil {
			return fmt.Errorf("create default workspace: %w", err)
		}

		workspaceID, err = result.LastInsertId()
		if err != nil {
			return fmt.Errorf("read default workspace ID: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("find default workspace: %w", err)
	}

	if _, err := transaction.Exec(`
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES (?, ?, 'owner')
		ON CONFLICT (workspace_id, user_id) DO NOTHING
	`, workspaceID, userID); err != nil {
		return fmt.Errorf("create default workspace membership: %w", err)
	}

	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
