package httpapi

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/XiaoMoDern/dev-knowledge-base/backend/internal/store"
	"github.com/stretchr/testify/assert"
)

type fakeNotesStore struct {
	notes []store.Note
}

// newTestHandler 给 handler 测试一个用空 fakeNotesStore 的 http.Handler。
// 不需要预填数据的测试用它——1 行调用代替 2 行 NewHandler(&fakeNotesStore{}, nil)。
//
// 用法：
//   handler := newTestHandler(t)
//
// 需要 pre-populated notes 的测试不走这个 helper——直接 NewHandler(&fakeNotesStore{notes: ...}, nil)。
func newTestHandler(t *testing.T) http.Handler {
	t.Helper()
	return NewHandler(&fakeNotesStore{}, nil)
}

func (fake *fakeNotesStore) CreateNote(input store.CreateNoteInput) (store.Note, error) {
	note := store.Note{
		ID:         int64(len(fake.notes) + 1),
		Title:      input.Title,
		Content:    input.Content,
		Visibility: "private",
		CreatedAt:  "2026-07-16T00:00:00Z",
		UpdatedAt:  "2026-07-16T00:00:00Z",
	}
	fake.notes = append(fake.notes, note)
	return note, nil
}

func (fake *fakeNotesStore) ListNotes() ([]store.Note, error) {
	return fake.notes, nil
}

func (fake *fakeNotesStore) DeleteNote(noteID int64) error {
	for i, note := range fake.notes {
		if note.ID == noteID {
			fake.notes = append(fake.notes[:i], fake.notes[i+1:]...)
			return nil
		}
	}
	return sql.ErrNoRows
}

// GetNoteByID 是 fake 实现：跟真 store 行为一致——按 ID 找，返 note 或 sql.ErrNoRows。
// 不维护 CategoryName（fake 没建 categories 表），跟 SearchNotes 同款简化策略。
func (fake *fakeNotesStore) GetNoteByID(noteID int64) (store.Note, error) {
	for _, note := range fake.notes {
		if note.ID == noteID {
			return note, nil
		}
	}
	return store.Note{}, sql.ErrNoRows
}

// UpdateNote 是 fake 实现：跟真 store 行为一致——按 ID 找 + 三态 CategoryID。
//
// 三态语义（跟 store.UpdateNote 对齐）：
//   input.CategoryID == nil                  → 字段 omitted → 不动 note.CategoryID
//   *input.CategoryID == nil（双 nil）         → explicit null → 清空 note.CategoryID
//   *input.CategoryID != nil                 → 设 note.CategoryID = *(*input.CategoryID)
//
// fake 的校验要和真 store 保持一致，否则 httpapi 测出来的状态码和真接口对不上
// ——这条注释从 ImportNotes 沿用到 UpdateNote。
func (fake *fakeNotesStore) UpdateNote(noteID int64, input store.UpdateNoteInput) (store.Note, error) {
	for i, note := range fake.notes {
		if note.ID == noteID {
			fake.notes[i].Title = input.Title
			fake.notes[i].Content = input.Content
			switch {
			case input.CategoryID == nil:
				// omitted：不更新 category_id
			case *input.CategoryID == nil:
				// explicit null：清空
				fake.notes[i].CategoryID = nil
			default:
				newID := **input.CategoryID
				fake.notes[i].CategoryID = &newID
			}
			return fake.notes[i], nil
		}
	}
	return store.Note{}, sql.ErrNoRows
}

// ListNotesByCategory 是 fake 实现：跟真 store 行为一致——只返 category_id 匹配的。
// 注：category_id == 0 / nil 的 note 不算"匹配任何分类"，所以 0 不在结果里。
func (fake *fakeNotesStore) ListNotesByCategory(categoryID int64) ([]store.Note, error) {
	var filtered []store.Note
	for _, note := range fake.notes {
		if note.CategoryID != nil && *note.CategoryID == categoryID {
			filtered = append(filtered, note)
		}
	}
	return filtered, nil
}

// SearchNotes 是 fake 实现：跟真 store 行为一致——按 q / categoryId 过滤 + 分页。
// fake 不维护 CategoryName（不建 categories 表），Items 里 CategoryName 永远是 nil。
//
// categoryId 特殊值：*opts.CategoryID == 0 表示"未分类"（前端 URL 约定），
// fake 跟真 store 一致——翻译成 note.CategoryID == nil。
func (f *fakeNotesStore) SearchNotes(opts store.SearchOptions) (store.PaginatedNotes, error) {
	var filtered []store.Note
	for _, note := range f.notes {
		if opts.Query != "" {
			if !strings.Contains(note.Title, opts.Query) && !strings.Contains(note.Content, opts.Query) {
				continue
			}
		}
		if opts.CategoryID != nil {
			if *opts.CategoryID == 0 {
				// 未分类：note.CategoryID 必须是 nil
				if note.CategoryID != nil {
					continue
				}
			} else if note.CategoryID == nil || *note.CategoryID != *opts.CategoryID {
				continue
			}
		}
		filtered = append(filtered, note)
	}

	page := opts.Page
	if page < 1 {
		page = 1
	}
	pageSize := opts.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	start := (page - 1) * pageSize
	if start > len(filtered) {
		start = len(filtered)
	}
	end := start + pageSize
	if end > len(filtered) {
		end = len(filtered)
	}

	return store.PaginatedNotes{
		Items:    filtered[start:end],
		Total:    len(filtered),
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func TestNotesHandlerCreatesAndListsNotes(t *testing.T) {
	notesStore := &fakeNotesStore{}
	handler := NewHandler(notesStore, nil)

	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/notes",
		bytes.NewBufferString(`{"title":"SQLite 自动迁移","content":"迁移会自动创建表。"}`),
	)
	createResponse := httptest.NewRecorder()
	handler.ServeHTTP(createResponse, createRequest)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf("create status code = %d, want %d", createResponse.Code, http.StatusCreated)
	}

	var created store.Note
	if err := json.NewDecoder(createResponse.Body).Decode(&created); err != nil {
		t.Fatalf("decode created note: %v", err)
	}
	if created.Title != "SQLite 自动迁移" {
		t.Fatalf("created title = %q, want %q", created.Title, "SQLite 自动迁移")
	}

	listRequest := httptest.NewRequest(http.MethodGet, "/api/notes", nil)
	listResponse := httptest.NewRecorder()
	handler.ServeHTTP(listResponse, listRequest)

	if listResponse.Code != http.StatusOK {
		t.Fatalf("list status code = %d, want %d", listResponse.Code, http.StatusOK)
	}

	var body struct {
		Items []store.Note `json:"items"`
	}
	if err := json.NewDecoder(listResponse.Body).Decode(&body); err != nil {
		t.Fatalf("decode listed notes: %v", err)
	}
	if len(body.Items) != 1 {
		t.Fatalf("listed note count = %d, want 1", len(body.Items))
	}
	if body.Items[0].ID != created.ID {
		t.Fatalf("listed note ID = %d, want %d", body.Items[0].ID, created.ID)
	}
}

func TestNotesHandlerRejectsBlankTitle(t *testing.T) {
	handler := newTestHandler(t)
	request := httptest.NewRequest(http.MethodPost, "/api/notes", bytes.NewBufferString(`{"title":"   ","content":"正文"}`))
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestNotesHandlerDeletesNote(t *testing.T) {
	notesStore := &fakeNotesStore{}
	notesStore.notes = append(notesStore.notes, store.Note{ID: 1, Title: "待删除", Content: "内容"})
	handler := NewHandler(notesStore, nil)

	request := httptest.NewRequest(http.MethodDelete, "/api/notes/1", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("delete status code = %d, want %d", response.Code, http.StatusNoContent)
	}
	if response.Body.Len() != 0 {
		t.Fatalf("delete response body = %q, want empty", response.Body.String())
	}
	if len(notesStore.notes) != 0 {
		t.Fatalf("notes count after delete = %d, want 0", len(notesStore.notes))
	}
}

func TestNotesHandlerRejectsInvalidDeleteID(t *testing.T) {
	handler := newTestHandler(t)

	//表驱动
	tests := []struct {
		name  string
		rawID string
	}{
		{"ID 为 0", "0"},
		{"负数 ID", "-1"},
		{"非数字 ID", "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodDelete, "/api/notes/"+tt.rawID, nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			assert.Equal(t, http.StatusBadRequest, response.Code,
				"delete id %q 应返回 400，实际: %d", tt.rawID, response.Code)
		})
	}
}

func TestNotesHandlerDeleteMissingNoteReturns404(t *testing.T) {
	handler := newTestHandler(t)

	request := httptest.NewRequest(http.MethodDelete, "/api/notes/999", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("delete missing note status code = %d, want %d", response.Code, http.StatusNotFound)
	}
}

func TestNotesHandlerUpdatesNote(t *testing.T) {
	notesStore := &fakeNotesStore{}
	notesStore.notes = append(notesStore.notes, store.Note{ID: 1, Title: "原标题", Content: "原内容"})
	handler := NewHandler(notesStore, nil)

	request := httptest.NewRequest(
		http.MethodPut,
		"/api/notes/1",
		bytes.NewBufferString(`{"title":"新标题","content":"新内容"}`),
	)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("update status code = %d, want %d", response.Code, http.StatusOK)
	}

	var updated store.Note
	if err := json.NewDecoder(response.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated note: %v", err)
	}
	if updated.Title != "新标题" {
		t.Fatalf("updated title = %q, want %q", updated.Title, "新标题")
	}
	if notesStore.notes[0].Title != "新标题" {
		t.Fatalf("store title = %q, want %q", notesStore.notes[0].Title, "新标题")
	}
}

func TestNotesHandlerRejectsInvalidUpdateID(t *testing.T) {
	handler := newTestHandler(t)

	for _, rawID := range []string{"0", "-1", "abc"} {
		request := httptest.NewRequest(http.MethodPut, "/api/notes/"+rawID, bytes.NewBufferString(`{"title":"x","content":"y"}`))
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		if response.Code != http.StatusBadRequest {
			t.Fatalf("update id %q status code = %d, want %d", rawID, response.Code, http.StatusBadRequest)
		}
	}
}

func TestNotesHandlerUpdateBlankTitle(t *testing.T) {
	notesStore := &fakeNotesStore{}
	notesStore.notes = append(notesStore.notes, store.Note{ID: 1, Title: "原", Content: "原"})
	handler := NewHandler(notesStore, nil)

	request := httptest.NewRequest(http.MethodPut, "/api/notes/1", bytes.NewBufferString(`{"title":"   ","content":"y"}`))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("update blank title status code = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

// fake 的 ImportNotes：不做事务（test fixture 不需要事务边界）
// 行为和真 Store 一样：title 空记 Errors，合法 append 进切片
func (f *fakeNotesStore) ImportNotes(inputs []store.ImportNoteInput) (store.ImportResult, error) {
	result := store.ImportResult{
		Items:  make([]store.Note, 0),
		Errors: make([]store.ImportError, 0),
	}
	nextID := int64(len(f.notes) + 100) // 偏移 ID 避免和真实数据冲突
	for i, input := range inputs {
		// 注意：fake 的校验要和真 store 保持一致，否则 httpapi 测出来
		// 的状态码和真接口对不上
		if strings.TrimSpace(input.Title) == "" {
			result.Failed++
			result.Errors = append(result.Errors, store.ImportError{
				Index:  i,
				Title:  input.Title,
				Reason: "title 不能为空",
			})
			continue
		}
		note := store.Note{
			ID:         nextID,
			Title:      strings.TrimSpace(input.Title),
			Content:    input.Content,
			Visibility: "private",
			CreatedAt:  "2026-07-20T00:00:00Z",
			UpdatedAt:  "2026-07-20T00:00:00Z",
		}
		f.notes = append(f.notes, note)
		result.Imported++
		result.Items = append(result.Items, note)
		nextID++
	}
	return result, nil
}

func TestNotesHandlerUpdateMissingNoteReturns404(t *testing.T) {
	handler := newTestHandler(t)

	request := httptest.NewRequest(http.MethodPut, "/api/notes/999", bytes.NewBufferString(`{"title":"x","content":"y"}`))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("update missing note status code = %d, want %d", response.Code, http.StatusNotFound)
	}
}

// TestNotesHandlerListsByCategory 验证 ?categoryId=N query 参数。
// 当前 notesHandler.list 不解析 query —— Red。
func TestNotesHandlerListsByCategory(t *testing.T) {
	catID := int64(1)
	notesStore := &fakeNotesStore{}
	notesStore.notes = append(notesStore.notes,
		store.Note{ID: 1, Title: "Go 入门", CategoryID: &catID},
		store.Note{ID: 2, Title: "Vue 入门", CategoryID: nil},
	)
	handler := NewHandler(notesStore, nil)

	request := httptest.NewRequest(http.MethodGet, "/api/notes?categoryId=1", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("list by category status code = %d, want %d", response.Code, http.StatusOK)
	}

	var body store.PaginatedNotes
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Items) != 1 {
		t.Fatalf("listed note count = %d, want 1", len(body.Items))
	}
	if body.Items[0].ID != 1 {
		t.Fatalf("listed note ID = %d, want 1", body.Items[0].ID)
	}
}

// TestNotesHandlerRejectsInvalidCategoryID 验证 ?categoryId= 非整数 / 负数 → 400
func TestNotesHandlerRejectsInvalidCategoryID(t *testing.T) {
	handler := newTestHandler(t)

	for _, raw := range []string{"abc", "-1"} {
		request := httptest.NewRequest(http.MethodGet, "/api/notes?categoryId="+raw, nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		if response.Code != http.StatusBadRequest {
			t.Fatalf("categoryId=%q status code = %d, want %d", raw, response.Code, http.StatusBadRequest)
		}
	}
}

// TestNotesHandlerList_UncategorizedReturnsNullCategoryNotes 修复 P0-4：
//
//	前端 CategorySidebar "未分类"按钮写入 ?categoryId=0。
//	当前 handler 跳过 0 + store 跳过 0 → URL 是"未分类"但实际返全部笔记（数据/UI 不一致）。
//
// 回归：handler 必须接受 categoryId=0，store.SearchNotes 翻译成 IS NULL。
func TestNotesHandlerList_UncategorizedReturnsNullCategoryNotes(t *testing.T) {
	catID := int64(1)
	notesStore := &fakeNotesStore{}
	notesStore.notes = append(notesStore.notes,
		store.Note{ID: 1, Title: "Go 入门", CategoryID: &catID},
		store.Note{ID: 2, Title: "无分类笔记 A", CategoryID: nil},
		store.Note{ID: 3, Title: "无分类笔记 B", CategoryID: nil},
	)
	handler := NewHandler(notesStore, nil)

	request := httptest.NewRequest(http.MethodGet, "/api/notes?categoryId=0", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	assert.Equal(t, http.StatusOK, response.Code)

	var body store.PaginatedNotes
	assert.NoError(t, json.NewDecoder(response.Body).Decode(&body))
	assert.Equal(t, 2, body.Total, "?categoryId=0 应只返 2 条无分类笔记")
	assert.Len(t, body.Items, 2)
	for _, item := range body.Items {
		assert.Nil(t, item.CategoryID, "未分类结果里 CategoryID 应为 nil")
	}
}

// TestNotesHandlerSearchesByKeyword 验证 ?q=xxx 搜索：title 或 content 含关键字的 note 都返回
func TestNotesHandlerSearchesByKeyword(t *testing.T) {
	notesStore := &fakeNotesStore{}
	notesStore.notes = append(notesStore.notes,
		store.Note{ID: 1, Title: "Go 入门", Content: "学习 Go 基础语法"},
		store.Note{ID: 2, Title: "Python 入门", Content: "学习 Python 基础"},
		store.Note{ID: 3, Title: "Vue 入门", Content: "前端框架学习"},
	)
	handler := NewHandler(notesStore, nil)

	request := httptest.NewRequest(http.MethodGet, "/api/notes?q=Go", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("search status code = %d, want %d", response.Code, http.StatusOK)
	}

	var body store.PaginatedNotes
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Items) != 1 {
		t.Fatalf("result count = %d, want 1", len(body.Items))
	}
	if body.Items[0].Title != "Go 入门" {
		t.Fatalf("first result title = %q, want %q", body.Items[0].Title, "Go 入门")
	}
	if body.Total != 1 {
		t.Fatalf("total = %d, want 1", body.Total)
	}
}

// TestNotesHandlerPaginatesNotes 验证 ?page=&pageSize= 分页：page=2 pageSize=10 返第 11-20 条
func TestNotesHandlerPaginatesNotes(t *testing.T) {
	notesStore := &fakeNotesStore{}
	for i := 1; i <= 25; i++ {
		notesStore.notes = append(notesStore.notes, store.Note{
			ID:    int64(i),
			Title: "note-" + strconv.Itoa(i),
		})
	}
	handler := NewHandler(notesStore, nil)

	request := httptest.NewRequest(http.MethodGet, "/api/notes?page=2&pageSize=10", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("paginate status code = %d, want %d", response.Code, http.StatusOK)
	}

	var body store.PaginatedNotes
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(body.Items) != 10 {
		t.Fatalf("page 2 count = %d, want 10", len(body.Items))
	}
	if body.Total != 25 {
		t.Fatalf("total = %d, want 25", body.Total)
	}
	if body.Page != 2 {
		t.Fatalf("page = %d, want 2", body.Page)
	}
	if body.PageSize != 10 {
		t.Fatalf("pageSize = %d, want 10", body.PageSize)
	}
}

// TestNotesHandlerGetByID：GET /api/notes/{id} 返完整 Note JSON
// 这是 P1-A 第二页 bug 修复的核心 API：详情页/编辑页用它替换 listNotes() 找 ID。
func TestNotesHandlerGetByID(t *testing.T) {
	notesStore := &fakeNotesStore{}
	notesStore.notes = append(notesStore.notes, store.Note{
		ID:      7,
		Title:   "第二页笔记",
		Content: "这条在原列表第 21 条之后",
	})
	handler := NewHandler(notesStore, nil)

	request := httptest.NewRequest(http.MethodGet, "/api/notes/7", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusOK)
	}

	var fetched store.Note
	if err := json.NewDecoder(response.Body).Decode(&fetched); err != nil {
		t.Fatalf("decode fetched note: %v", err)
	}
	if fetched.ID != 7 {
		t.Fatalf("fetched ID = %d, want 7", fetched.ID)
	}
	if fetched.Title != "第二页笔记" {
		t.Fatalf("fetched title = %q, want %q", fetched.Title, "第二页笔记")
	}
}

// TestNotesHandlerGetMissingReturns404：找不到的 ID 必须 404，不能用 500 蒙混过关。
// 404 是约定的"客户端错了（指错 ID）"语义；500 是"服务端炸了"语义——不能混。
func TestNotesHandlerGetMissingReturns404(t *testing.T) {
	handler := newTestHandler(t)

	request := httptest.NewRequest(http.MethodGet, "/api/notes/999", nil)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusNotFound)
	}
}

// TestNotesHandlerRejectsInvalidGetID：路径参数守门——非正整数 ID 必须 400，跟 delete/update 同款。
//
// Phase E.1：for-range 跑多场景改表驱动 + assert.Equal。
//
//	Before：for 循环 + t.Fatalf 包在循环里，失败就终止整个测试
//	After ：tests := []struct{...}{...} 表驱动 + t.Run 子测试，每个失败独立报不影响其他场景
//
// 表驱动优势：
//  1. 加新场景只加一行 struct，不用复制 paste for 循环
//  2. 子测试名自动出现在测试输出（t.Run(tt.name, ...)），失败容易定位
//  3. 想跑某个特定场景用 `go test -run TestX/<subname>`
func TestNotesHandlerRejectsInvalidGetID(t *testing.T) {
	handler := newTestHandler(t)

	tests := []struct {
		name  string
		rawID string
	}{
		{name: "zero", rawID: "0"},
		{name: "negative", rawID: "-1"},
		{name: "non-numeric", rawID: "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodGet, "/api/notes/"+tt.rawID, nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			assert.Equal(t, http.StatusBadRequest, response.Code,
				"非正整数 ID %q 应该被拒，但 handler 返了 %d", tt.rawID, response.Code)
		})
	}
}
