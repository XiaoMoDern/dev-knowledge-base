package httpapi

import (
	"net/http"

	"github.com/XiaoMoDern/dev-knowledge-base/backend/internal/store"
)

type NotesStore interface {
	CreateNote(store.CreateNoteInput) (store.Note, error)
	ListNotes() ([]store.Note, error)
	DeleteNote(int64) error
	UpdateNote(int64, store.UpdateNoteInput) (store.Note, error)
}

func NewHandler(notesStore NotesStore) http.Handler {
	router := http.NewServeMux()
	notes := notesHandler{notesStore: notesStore}

	router.HandleFunc("GET /api/health", healthHandler)
	router.HandleFunc("GET /api/notes", notes.list)
	router.HandleFunc("POST /api/notes", notes.create)
	//新增一个删除路由
	router.HandleFunc("DELETE /api/notes/{id}", notes.delete)
	router.HandleFunc("PUT /api/notes/{id}", notes.update)

	return router
}

func healthHandler(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)

	_, _ = response.Write([]byte(`{"status":"ok"}`))
}
