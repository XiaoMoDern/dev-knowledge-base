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
)

type fakeNotesStore struct {
	notes []store.Note
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

func (fake *fakeNotesStore) UpdateNote(noteID int64, input store.UpdateNoteInput) (store.Note, error) {
	for i, note := range fake.notes {
		if note.ID == noteID {
			fake.notes[i].Title = input.Title
			fake.notes[i].Content = input.Content
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
func (f *fakeNotesStore) SearchNotes(opts store.SearchOptions) (store.PaginatedNotes, error) {
	var filtered []store.Note
	for _, note := range f.notes {
		if opts.Query != "" {
			if !strings.Contains(note.Title, opts.Query) && !strings.Contains(note.Content, opts.Query) {
				continue
			}
		}
		if opts.CategoryID != nil {
			if note.CategoryID == nil || *note.CategoryID != *opts.CategoryID {
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
	handler := NewHandler(&fakeNotesStore{}, nil)
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
	handler := NewHandler(&fakeNotesStore{}, nil)

	for _, rawID := range []string{"0", "-1", "abc"} {
		request := httptest.NewRequest(http.MethodDelete, "/api/notes/"+rawID, nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		if response.Code != http.StatusBadRequest {
			t.Fatalf("delete id %q status code = %d, want %d", rawID, response.Code, http.StatusBadRequest)
		}
	}
}

func TestNotesHandlerDeleteMissingNoteReturns404(t *testing.T) {
	handler := NewHandler(&fakeNotesStore{}, nil)

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
	handler := NewHandler(&fakeNotesStore{}, nil)

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
	handler := NewHandler(&fakeNotesStore{}, nil)

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
	handler := NewHandler(&fakeNotesStore{}, nil)

	for _, raw := range []string{"abc", "-1"} {
		request := httptest.NewRequest(http.MethodGet, "/api/notes?categoryId="+raw, nil)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)

		if response.Code != http.StatusBadRequest {
			t.Fatalf("categoryId=%q status code = %d, want %d", raw, response.Code, http.StatusBadRequest)
		}
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
