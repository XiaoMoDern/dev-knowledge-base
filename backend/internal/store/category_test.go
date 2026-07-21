package store

import (
	"path/filepath"
	"testing"
)

// TestStoreCreatesAndListsCategories：建 2 个分类，列出应该有 2 个，后建的在前。
// 教学点：跟 note.go 的 CreateNote / ListNotes 是同一套模式——
// 区别在于 categories 是工作空间级别（先找 workspace_id 再 INSERT），
// 而且没有 "visibility" 那种业务字段，更简单。
func TestStoreCreatesAndListsCategories(t *testing.T) {
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

	// 1. 创建 2 个分类
	first, err := database.CreateCategory("Go")
	if err != nil {
		t.Fatalf("create category Go: %v", err)
	}
	if first.ID <= 0 {
		t.Fatalf("created category ID = %d, want a positive ID", first.ID)
	}
	if first.Name != "Go" {
		t.Fatalf("created category name = %q, want %q", first.Name, "Go")
	}

	second, err := database.CreateCategory("Vue")
	if err != nil {
		t.Fatalf("create category Vue: %v", err)
	}
	if second.ID == first.ID {
		t.Fatalf("two categories got the same ID = %d", first.ID)
	}

	// 2. 列出应该有 2 个
	categories, err := database.ListCategories()
	if err != nil {
		t.Fatalf("list categories: %v", err)
	}
	if len(categories) != 2 {
		t.Fatalf("category count = %d, want 2", len(categories))
	}

	// 3. 顺序：跟 note 列表保持一致——后建的在前（按 id DESC）
	if categories[0].Name != "Vue" {
		t.Fatalf("first listed category = %q, want %q", categories[0].Name, "Vue")
	}
	if categories[1].Name != "Go" {
		t.Fatalf("second listed category = %q, want %q", categories[1].Name, "Go")
	}
}
