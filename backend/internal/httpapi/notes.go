package httpapi

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/XiaoMoDern/dev-knowledge-base/backend/internal/store"
)

type notesHandler struct {
	notesStore NotesStore
}

func (handler notesHandler) create(response http.ResponseWriter, request *http.Request) {
	var input struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeJSON(response, http.StatusBadRequest, map[string]string{"error": "请求体必须是包含 title 和 content 的 JSON"})
		return
	}

	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		writeJSON(response, http.StatusBadRequest, map[string]string{"error": "title 不能为空"})
		return
	}

	note, err := handler.notesStore.CreateNote(store.CreateNoteInput{
		Title:   input.Title,
		Content: input.Content,
	})
	if err != nil {
		writeJSON(response, http.StatusInternalServerError, map[string]string{"error": "创建笔记失败"})
		return
	}

	log.Println("create note", input.Title, input.Content)

	writeJSON(response, http.StatusCreated, note)
}

func (handler notesHandler) list(response http.ResponseWriter, request *http.Request) {
	notes, err := handler.notesStore.ListNotes()
	if err != nil {
		writeJSON(response, http.StatusInternalServerError, map[string]string{"error": "查询笔记失败"})
		return
	}

	writeJSON(response, http.StatusOK, struct {
		Items []store.Note `json:"items"`
	}{Items: notes})
}

func (handler notesHandler) delete(response http.ResponseWriter, request *http.Request) {
	rawID := request.PathValue("id")

	noteID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || noteID <= 0 {
		writeJSON(response, http.StatusBadRequest, map[string]string{"error": "笔记 ID 必须是正整数"})
		return
	}

	err = handler.notesStore.DeleteNote(noteID)
	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(response, http.StatusNotFound, map[string]string{"error": "笔记不存在"})
		return
	}

	if err != nil {
		writeJSON(response, http.StatusInternalServerError, map[string]string{"error": "删除笔记失败"})
		return
	}

	response.WriteHeader(http.StatusNoContent)

}

func (handler notesHandler) update(response http.ResponseWriter, request *http.Request) {
	rawID := request.PathValue("id")
	noteID, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil || noteID <= 0 {
		writeJSON(response, http.StatusBadRequest, map[string]string{"error": "笔记 ID 必须是正整数"})
		return
	}
	var input struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeJSON(response, http.StatusBadRequest, map[string]string{"error": "请求体必须是包含 title 和 content 的 JSON"})
		return
	}

	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		writeJSON(response, http.StatusBadRequest, map[string]string{"error": "title 不能为空"})
		return
	}

	note, err := handler.notesStore.UpdateNote(noteID, store.UpdateNoteInput{
		Title:   input.Title,
		Content: input.Content,
	})
	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(response, http.StatusNotFound, map[string]string{"error": "笔记不存在"})
		return
	}

	if err != nil {
		writeJSON(response, http.StatusInternalServerError, map[string]string{"error": "更新笔记失败"})
		return
	}

	writeJSON(response, http.StatusOK, note)
}
func writeJSON(response http.ResponseWriter, status int, body any) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(body)
}
