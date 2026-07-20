package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenCreatesDatabaseFile(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "dev-notes.db")

	database, err := Open(databasePath)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("close database: %v", err)
		}
	})

	if _, err := os.Stat(databasePath); err != nil {
		t.Fatalf("database file should exist: %v", err)
	}
}

func TestOpenCreatesCoreTables(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "dev-notes.db")

	database, err := Open(databasePath)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("close database: %v", err)
		}
	})

	coreTableNames := []string{
		"users",
		"workspaces",
		"workspace_members",
		"categories",
		"notes",
	}

	for _, tableName := range coreTableNames {
		var count int

		err := database.db.QueryRow(
			`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?`,
			tableName,
		).Scan(&count)
		if err != nil {
			t.Fatalf("check table %q: %v", tableName, err)
		}

		if count != 1 {
			t.Fatalf("table %q should exist", tableName)
		}
	}
}

func TestOpenCreatesDefaultWorkspace(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "dev-notes.db")

	database, err := Open(databasePath)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() {
		if err := database.Close(); err != nil {
			t.Errorf("close database: %v", err)
		}
	})

	var workspaceID int64
	var username string
	var workspaceName string
	var role string

	err = database.db.QueryRow(`
		SELECT w.id, u.username, w.name, wm.role
		FROM users u
		JOIN workspaces w ON w.owner_user_id = u.id
		JOIN workspace_members wm ON wm.workspace_id = w.id AND wm.user_id = u.id
		WHERE u.username = ? AND w.name = ?
	`, "local", "我的知识库").Scan(&workspaceID, &username, &workspaceName, &role)
	if err != nil {
		t.Fatalf("find default workspace: %v", err)
	}

	if workspaceID <= 0 {
		t.Fatalf("workspace ID = %d, want a positive ID", workspaceID)
	}

	if username != "local" {
		t.Fatalf("username = %q, want %q", username, "local")
	}

	if workspaceName != "我的知识库" {
		t.Fatalf("workspace name = %q, want %q", workspaceName, "我的知识库")
	}

	if role != "owner" {
		t.Fatalf("workspace role = %q, want %q", role, "owner")
	}
}

func TestOpenDoesNotDuplicateDefaultWorkspace(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "dev-notes.db")

	firstStore, err := Open(databasePath)
	if err != nil {
		t.Fatalf("open database for the first time: %v", err)
	}
	if err := firstStore.Close(); err != nil {
		t.Fatalf("close first database: %v", err)
	}

	secondStore, err := Open(databasePath)
	if err != nil {
		t.Fatalf("open database for the second time: %v", err)
	}
	t.Cleanup(func() {
		if err := secondStore.Close(); err != nil {
			t.Errorf("close second database: %v", err)
		}
	})

	var userCount int
	var workspaceCount int
	var membershipCount int

	if err := secondStore.db.QueryRow(`SELECT COUNT(*) FROM users WHERE username = ?`, defaultUsername).Scan(&userCount); err != nil {
		t.Fatalf("count default users: %v", err)
	}
	if err := secondStore.db.QueryRow(`SELECT COUNT(*) FROM workspaces WHERE name = ?`, defaultWorkspaceName).Scan(&workspaceCount); err != nil {
		t.Fatalf("count default workspaces: %v", err)
	}
	if err := secondStore.db.QueryRow(`SELECT COUNT(*) FROM workspace_members WHERE role = 'owner'`).Scan(&membershipCount); err != nil {
		t.Fatalf("count default memberships: %v", err)
	}

	if userCount != 1 {
		t.Fatalf("default user count = %d, want 1", userCount)
	}
	if workspaceCount != 1 {
		t.Fatalf("default workspace count = %d, want 1", workspaceCount)
	}
	if membershipCount != 1 {
		t.Fatalf("default membership count = %d, want 1", membershipCount)
	}
}
