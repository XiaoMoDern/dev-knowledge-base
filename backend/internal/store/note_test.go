package store

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func TestStoreCreatesAndListsNotes(t *testing.T) {
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

	created, err := database.CreateNote(CreateNoteInput{
		Title:   "SQLite 自动迁移",
		Content: "迁移会让空数据库自动具备应用需要的表结构。",
	})
	if err != nil {
		t.Fatalf("create note: %v", err)
	}

	if created.ID <= 0 {
		t.Fatalf("created note ID = %d, want a positive ID", created.ID)
	}
	if created.Visibility != "private" {
		t.Fatalf("created note visibility = %q, want %q", created.Visibility, "private")
	}
	if created.CreatedAt == "" || created.UpdatedAt == "" {
		t.Fatalf("created timestamps must not be empty: %#v", created)
	}

	notes, err := database.ListNotes()
	if err != nil {
		t.Fatalf("list notes: %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("note count = %d, want 1", len(notes))
	}
	if notes[0].ID != created.ID {
		t.Fatalf("listed note ID = %d, want %d", notes[0].ID, created.ID)
	}
	if notes[0].Title != created.Title {
		t.Fatalf("listed note title = %q, want %q", notes[0].Title, created.Title)
	}
	if notes[0].Content != created.Content {
		t.Fatalf("listed note content = %q, want %q", notes[0].Content, created.Content)
	}
}

func TestStoreDeletesNote(t *testing.T) {
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

	created, err := database.CreateNote(CreateNoteInput{
		Title:   "待删除笔记",
		Content: "这条笔记马上会被删除",
	})
	if err != nil {
		t.Fatalf("create note: %v", err)
	}

	err = database.DeleteNote(created.ID)
	if err != nil {
		t.Fatalf("delete note: %v", err)
	}

	notes, err := database.ListNotes()
	if err != nil {
		t.Fatalf("list notes: %v", err)
	}

	if len(notes) != 0 {
		t.Fatalf("note count = %d, want 0", len(notes))
	}
}

// TestStoreUpdatesNote：先建一条笔记 、 TestStoreUpdateMissingNoteReturnsErrNoRows：更新
func TestStoreUpdatesNote(t *testing.T) {
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
	created, err := database.CreateNote(CreateNoteInput{
		Title:   "待更新笔记",
		Content: "这条笔记马上会被更新",
	})
	if err != nil {
		t.Fatalf("create note: %v", err)
	}
	updated, err := database.UpdateNote(created.ID, UpdateNoteInput{
		Title:   "新标题",
		Content: "新内容",
	})
	if err != nil {
		t.Fatalf("update note: %v", err)
	}

	if updated.Title != "新标题" {
		t.Fatalf("updated title = %q, want %q", updated.Title, "新标题")
	}
	if updated.Content != "新内容" {
		t.Fatalf("updated content = %q, want %q", updated.Content, "新内容")
	}
	if updated.UpdatedAt == created.UpdatedAt {
		t.Fatalf("updated_at should change: %q == %q", updated.UpdatedAt, created.UpdatedAt)
	}

	notes, err := database.ListNotes()
	if err != nil {
		t.Fatalf("list notes: %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("note count = %d, want 1", len(notes))
	}
	if notes[0].Title != "新标题" {
		t.Fatalf("listed note title = %q, want %q", notes[0].Title, "新标题")
	}

}

func TestStoreImportsNotes(t *testing.T) {
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

	notes := []ImportNoteInput{
		{Title: "笔记一", Content: "内容一"},
		{Title: "笔记二", Content: "内容二"},
		{Title: "笔记三", Content: "内容三"},
	}
	result, err := database.ImportNotes(notes)
	if err != nil {
		t.Fatalf("import notes: %v", err)
	}

	//断言
	if result.Imported != 3 {
		t.Fatalf("Imported = %d, want 3", result.Imported)
	}
	if result.Failed != 0 {
		t.Fatalf("Failed = %d, want 0", result.Failed)
	}
	if len(result.Items) != 3 {
		t.Fatalf("len(Items) = %d, want 3", len(result.Items))
	}

	if len(result.Errors) != 0 {
		t.Fatalf("len(Errors) = %d, want 0", len(result.Errors))
	}

}

func TestStoreImportSkipsInvalidNotes(t *testing.T) {
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

	notes := []ImportNoteInput{
		{Title: "笔记一", Content: "内容一"},
		{Title: "", Content: "内容二"}, // 空 title 应该被记到 Errors
		{Title: "笔记三", Content: "内容三"},
	}
	result, err := database.ImportNotes(notes)
	if err != nil {
		t.Fatalf("import notes: %v", err)
	}

	if result.Imported != 2 {
		t.Fatalf("Imported = %d, want 2", result.Imported)
	}
	if result.Failed != 1 {
		t.Fatalf("Failed = %d, want 1", result.Failed)
	}
	if len(result.Items) != 2 {
		t.Fatalf("len(Items) = %d, want 2", len(result.Items))
	}
	if len(result.Errors) != 1 {
		t.Fatalf("len(Errors) = %d, want 1", len(result.Errors))
	}
	// len(Errors) > 0 是为了避免空 slice 索引导致另一个错
	if len(result.Errors) > 0 && result.Errors[0].Index != 1 {
		t.Fatalf("Errors[0].Index = %d, want 1", result.Errors[0].Index)
	}
}

func TestStoreUpdateMissingNoteReturnsErrNoRows(t *testing.T) {
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

	_, err = database.UpdateNote(999, UpdateNoteInput{
		Title:   "不存在",
		Content: "内容",
	})
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("update missing note err = %v, want sql.ErrNoRows", err)
	}
}

// TestStoreListsNotesWithCategoryName 验证 ListNotes 的 LEFT JOIN：
// 关联分类的 note 应带 CategoryName，没分类的应是 nil。
// 当前 ListNotes 还没加 JOIN，CategoryName 永远是 nil —— Red。
func TestStoreListsNotesWithCategoryName(t *testing.T) {
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

	// 1. 建分类
	category, err := database.CreateCategory("Go")
	if err != nil {
		t.Fatalf("create category: %v", err)
	}

	// 2. 直接 SQL 插 2 个 note（CreateNote 还没支持 categoryId，绕开 store 接口）
	workspaceID, err := database.defaultWorkspaceID()
	if err != nil {
		t.Fatalf("find workspace: %v", err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)

	// 2a. 关联 "Go" 的 note
	_, err = database.db.Exec(`
		INSERT INTO notes (workspace_id, category_id, title, content, visibility, created_at, updated_at)
		VALUES (?, ?, 'Go 笔记', '内容', 'private', ?, ?)
	`, workspaceID, category.ID, now, now)
	if err != nil {
		t.Fatalf("insert note with category: %v", err)
	}

	// 2b. 没分类的 note（category_id 显式 NULL）
	_, err = database.db.Exec(`
		INSERT INTO notes (workspace_id, category_id, title, content, visibility, created_at, updated_at)
		VALUES (?, NULL, '无分类笔记', '内容', 'private', ?, ?)
	`, workspaceID, now, now)
	if err != nil {
		t.Fatalf("insert note without category: %v", err)
	}

	// 3. ListNotes
	notes, err := database.ListNotes()
	if err != nil {
		t.Fatalf("list notes: %v", err)
	}
	if len(notes) != 2 {
		t.Fatalf("note count = %d, want 2", len(notes))
	}

	// 4. 断言 CategoryName
	// 排序按 updated_at DESC 同秒、id DESC —— 后插的"无分类笔记"在前
	if notes[0].CategoryName != nil {
		t.Fatalf("first note (无分类) CategoryName = %q, want nil", *notes[0].CategoryName)
	}
	if notes[1].CategoryName == nil || *notes[1].CategoryName != "Go" {
		t.Fatalf("second note (Go 笔记) CategoryName = %v, want %q", notes[1].CategoryName, "Go")
	}
}

// TestStoreListsNotesByCategory 验证按分类过滤：
// 关联 cat1 的 note 出现，关联 cat2 的 note 不出现，没分类的 note 不出现。
// 当前 ListNotesByCategory 还没实现 —— Red。
func TestStoreListsNotesByCategory(t *testing.T) {
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

	// 1. 建 2 个分类
	catGo, err := database.CreateCategory("Go")
	if err != nil {
		t.Fatalf("create category Go: %v", err)
	}
	catVue, err := database.CreateCategory("Vue")
	if err != nil {
		t.Fatalf("create category Vue: %v", err)
	}

	// 2. 直接 SQL 插 3 个 note（绕过 CreateNote 限制）
	workspaceID, err := database.defaultWorkspaceID()
	if err != nil {
		t.Fatalf("find workspace: %v", err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)

	// 2a. 关联 Go
	_, err = database.db.Exec(`
		INSERT INTO notes (workspace_id, category_id, title, content, visibility, created_at, updated_at)
		VALUES (?, ?, 'Go 入门', '', 'private', ?, ?)
	`, workspaceID, catGo.ID, now, now)
	if err != nil {
		t.Fatalf("insert note for Go: %v", err)
	}

	// 2b. 关联 Vue
	_, err = database.db.Exec(`
		INSERT INTO notes (workspace_id, category_id, title, content, visibility, created_at, updated_at)
		VALUES (?, ?, 'Vue 入门', '', 'private', ?, ?)
	`, workspaceID, catVue.ID, now, now)
	if err != nil {
		t.Fatalf("insert note for Vue: %v", err)
	}

	// 2c. 没分类
	_, err = database.db.Exec(`
		INSERT INTO notes (workspace_id, category_id, title, content, visibility, created_at, updated_at)
		VALUES (?, NULL, '无分类笔记', '', 'private', ?, ?)
	`, workspaceID, now, now)
	if err != nil {
		t.Fatalf("insert note without category: %v", err)
	}

	// 3. ListNotesByCategory(Go) —— 应该只返回 "Go 入门"
	notes, err := database.ListNotesByCategory(catGo.ID)
	if err != nil {
		t.Fatalf("list notes by category Go: %v", err)
	}
	if len(notes) != 1 {
		t.Fatalf("note count for Go = %d, want 1", len(notes))
	}
	if notes[0].Title != "Go 入门" {
		t.Fatalf("first note title = %q, want %q", notes[0].Title, "Go 入门")
	}
	// 复用 JOIN：CategoryName 应该 = "Go"
	if notes[0].CategoryName == nil || *notes[0].CategoryName != "Go" {
		t.Fatalf("first note CategoryName = %v, want %q", notes[0].CategoryName, "Go")
	}
}
