package store

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// setupTestStore 为 store 包测试初始化一个干净的 SQLite 数据库。
//
// 调用方一行搞定：
//   database := setupTestStore(t)
//   // 紧接着写测试逻辑
//
// 不用管临时目录、清理、初始化路径——全在这。
//
// 设计要点：
//   1. t.Helper() — 失败时打印调用方行号，不打 helper 内部
//   2. t.TempDir() — 每个测试独立临时目录，自动清理 + 跨测试隔离
//   3. store.Open() 自动迁移（CREATE TABLE IF NOT EXISTS），不用测试手动 init
//   4. t.Cleanup vs defer — 测试函数专属，自动调度、跨 sub-test 共享
//   5. return database 单一返回 — 不返 cleanup 出去，cleanup 走 t.Cleanup 自动挂钩（接口最小化）
//   6. t.Fatalf vs t.Errorf — 数据库没准备好后续断言必失败，所以致命

func setupTestStore(t *testing.T) *Store {
	t.Helper()

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

	return database
}

func TestStoreCreatesAndListsNotes(t *testing.T) {
	database := setupTestStore(t)

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
	database := setupTestStore(t)

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
	database := setupTestStore(t)

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
	database := setupTestStore(t)

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
	database := setupTestStore(t)

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
	database := setupTestStore(t)

	_, err := database.UpdateNote(999, UpdateNoteInput{
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
	database := setupTestStore(t)

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
	database := setupTestStore(t)

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

// TestStoreSearchNotesByKeyword 验证搜索关键字过滤：title 或 content 含关键字的 note 都返回。
// 当前 SearchNotes 还没实现 —— Red。
func TestStoreSearchNotesByKeyword(t *testing.T) {
	database := setupTestStore(t)

	workspaceID, err := database.defaultWorkspaceID()
	if err != nil {
		t.Fatalf("find workspace: %v", err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)

	// 3 个 note：title 含 Go / Python / Vue
	testNotes := []struct {
		title   string
		content string
	}{
		{"Go 入门", "学习 Go 基础语法"},
		{"Python 入门", "学习 Python 基础"},
		{"Vue 入门", "前端框架学习"},
	}
	for _, n := range testNotes {
		_, err = database.db.Exec(`
			INSERT INTO notes (workspace_id, title, content, visibility, created_at, updated_at)
			VALUES (?, ?, ?, 'private', ?, ?)
		`, workspaceID, n.title, n.content, now, now)
		if err != nil {
			t.Fatalf("insert note %q: %v", n.title, err)
		}
	}

	// 搜 "Go" → 只返 "Go 入门"
	result, err := database.SearchNotes(SearchOptions{Query: "Go", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("search notes: %v", err)
	}
	if len(result.Items) != 1 {
		t.Fatalf("result count = %d, want 1", len(result.Items))
	}
	if result.Items[0].Title != "Go 入门" {
		t.Fatalf("first result title = %q, want %q", result.Items[0].Title, "Go 入门")
	}
	if result.Total != 1 {
		t.Fatalf("total = %d, want 1", result.Total)
	}
}

// TestStoreListsNotesWithPagination 验证分页：建 25 个 note，page=2 pageSize=10 返第 11-20 条。
// ORDER BY updated_at DESC：note-25 在前，note-1 在后。
// 当前 SearchNotes 还没实现 —— Red。
func TestStoreListsNotesWithPagination(t *testing.T) {
	database := setupTestStore(t)

	workspaceID, err := database.defaultWorkspaceID()
	if err != nil {
		t.Fatalf("find workspace: %v", err)
	}

	// 插 25 个 note：每条 updated_at 差 1 秒，保证 ORDER BY 顺序确定
	base := time.Now().UTC()
	for i := 1; i <= 25; i++ {
		updatedAt := base.Add(time.Duration(i) * time.Second).Format(time.RFC3339Nano)
		_, err = database.db.Exec(`
			INSERT INTO notes (workspace_id, title, content, visibility, created_at, updated_at)
			VALUES (?, ?, '', 'private', ?, ?)
		`, workspaceID, fmt.Sprintf("note-%d", i), updatedAt, updatedAt)
		if err != nil {
			t.Fatalf("insert note-%d: %v", i, err)
		}
	}

	// page=2 pageSize=10 → 第 11-20 条
	// page 1: note-25..note-16 / page 2: note-15..note-6 / page 3: note-5..note-1
	result, err := database.SearchNotes(SearchOptions{Page: 2, PageSize: 10})
	if err != nil {
		t.Fatalf("search notes: %v", err)
	}
	if len(result.Items) != 10 {
		t.Fatalf("page 2 count = %d, want 10", len(result.Items))
	}
	if result.Total != 25 {
		t.Fatalf("total = %d, want 25", result.Total)
	}
	if result.Items[0].Title != "note-15" {
		t.Fatalf("page 2 first title = %q, want %q", result.Items[0].Title, "note-15")
	}
	if result.Items[9].Title != "note-6" {
		t.Fatalf("page 2 last title = %q, want %q", result.Items[9].Title, "note-6")
	}
}

// TestStoreSearchAndPaginate 验证搜索 + 分页组合：建 30 个 note（10 Go + 20 Python），
// 搜 "Go" page=1 pageSize=5 → 5 条 total=10。
// 当前 SearchNotes 还没实现 —— Red。
func TestStoreSearchAndPaginate(t *testing.T) {
	database := setupTestStore(t)

	workspaceID, err := database.defaultWorkspaceID()
	if err != nil {
		t.Fatalf("find workspace: %v", err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)

	// 10 个 Go + 20 个 Python
	for i := 1; i <= 30; i++ {
		title := fmt.Sprintf("Python 笔记 %d", i)
		if i <= 10 {
			title = fmt.Sprintf("Go 笔记 %d", i)
		}
		_, err = database.db.Exec(`
			INSERT INTO notes (workspace_id, title, content, visibility, created_at, updated_at)
			VALUES (?, ?, '', 'private', ?, ?)
		`, workspaceID, title, now, now)
		if err != nil {
			t.Fatalf("insert note %d: %v", i, err)
		}
	}

	// 搜 "Go" page=1 pageSize=5 → 5 条 total=10
	result, err := database.SearchNotes(SearchOptions{Query: "Go", Page: 1, PageSize: 5})
	if err != nil {
		t.Fatalf("search notes: %v", err)
	}
	if len(result.Items) != 5 {
		t.Fatalf("result count = %d, want 5", len(result.Items))
	}
	if result.Total != 10 {
		t.Fatalf("total = %d, want 10", result.Total)
	}
}

// TestStoreGetNoteByID：建一条 note，按 ID 拿回完整数据（含 categoryName）
// 验证：GetNoteByID 是按 ID 精确查，跟 ListNotes 的全表扫描不同路径
//
// Phase E.1：assert.Equal 替代 if t.Fatalf 样板。
//
//	Before：每个字段 3 行 assert（want xx / got xx / error msg），共 9 行
//	After ：每个字段 1 行 assert.Equal，共 3 行
func TestStoreGetNoteByID(t *testing.T) {
	database := setupTestStore(t)

	created, err := database.CreateNote(CreateNoteInput{
		Title:   "按 ID 查",
		Content: "这条用来验证 GetNoteByID",
	})
	assert.NoError(t, err)

	fetched, err := database.GetNoteByID(created.ID)
	assert.NoError(t, err)

	// assert.Equal 自动 %v 格式化 + 失败时打印 want/got/diff，省掉 3 行手写模板
	assert.Equal(t, created.ID, fetched.ID, "ID 应一致")
	assert.Equal(t, created.Title, fetched.Title, "Title 应一致")
	assert.Equal(t, created.Content, fetched.Content, "Content 应一致")
}

// TestStoreGetNoteByIDMissingReturnsErrNoRows：找不到的 ID 必须返 sql.ErrNoRows，
// 给 handler 用 errors.Is(err, sql.ErrNoRows) 判 404。
// 这是 P1-A 第二页 bug 修复的核心约定：如果 store 错误处理变了，handler 404 逻辑会全坏。
func TestStoreGetNoteByIDMissingReturnsErrNoRows(t *testing.T) {
	database := setupTestStore(t)

	// 不存在的 ID（单用户数据库全新建，ID 从 1 起，999 一定没有）
	_, err := database.GetNoteByID(999)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

// TestStoreHasNoteTrue：建一条 note，HasNote(id) 返回 true。
// 这是 Ray 第一次独立写的 store 方法——验证功能对、跑得通。
func TestStoreHasNoteTrue(t *testing.T) {
	database := setupTestStore(t)

	created, err := database.CreateNote(CreateNoteInput{
		Title:   "存在的笔记",
		Content: "用于验证 HasNote 返 true",
	})
	if err != nil {
		t.Fatalf("create note: %v", err)
	}

	exists, err := database.HasNote(created.ID)
	if err != nil {
		t.Fatalf("has note: %v", err)
	}
	if !exists {
		t.Fatalf("has note for existing ID %d = false, want true", created.ID)
	}
}

// TestStoreHasNoteFalse：不存在的 ID，HasNote 返回 false（不是 error，跟 GetNoteByID 区分）。
// 关键差别：HasNote 用 boolean 而非 ErrNoRows 表达"不存在"——调用方不写 errors.Is。
func TestStoreHasNoteFalse(t *testing.T) {
	database := setupTestStore(t)

	exists, err := database.HasNote(999)
	if err != nil {
		t.Fatalf("has note: %v", err)
	}
	if exists {
		t.Fatalf("has note for missing ID 999 = true, want false")
	}
}
