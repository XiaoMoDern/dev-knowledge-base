package httpapi

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestNotesHandlerCreatesAndListsNotes(t *testing.T) {
	notesStore := &fakeNotesStore{}
	handler := NewHandler(notesStore)

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
	handler := NewHandler(&fakeNotesStore{})
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
	handler := NewHandler(notesStore)

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
	handler := NewHandler(&fakeNotesStore{})

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
	handler := NewHandler(&fakeNotesStore{})

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
	handler := NewHandler(notesStore)

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
	handler := NewHandler(&fakeNotesStore{})

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
	handler := NewHandler(notesStore)

	request := httptest.NewRequest(http.MethodPut, "/api/notes/1", bytes.NewBufferString(`{"title":"   ","content":"y"}`))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("update blank title status code = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

func TestNotesHandlerUpdateMissingNoteReturns404(t *testing.T) {
	handler := NewHandler(&fakeNotesStore{})

	request := httptest.NewRequest(http.MethodPut, "/api/notes/999", bytes.NewBufferString(`{"title":"x","content":"y"}`))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	if response.Code != http.StatusNotFound {
		t.Fatalf("update missing note status code = %d, want %d", response.Code, http.StatusNotFound)
	}
}
