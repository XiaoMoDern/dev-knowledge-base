package store

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
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
