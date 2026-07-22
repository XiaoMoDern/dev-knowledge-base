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
		Title      string `json:"title"`
		Content    string `json:"content"`
		CategoryID *int64 `json:"categoryId"`
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
		Title:      input.Title,
		Content:    input.Content,
		CategoryID: input.CategoryID,
	})
	if err != nil {
		writeJSON(response, http.StatusInternalServerError, map[string]string{"error": "创建笔记失败"})
		return
	}

	log.Println("create note", input.Title, input.Content)

	writeJSON(response, http.StatusCreated, note)
}

func (handler notesHandler) list(response http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()
	q := query.Get("q")

	// 解析 ?categoryId=N query 参数
	var categoryID *int64
	if raw := request.URL.Query().Get("categoryId"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed < 0 {
			writeJSON(response, http.StatusBadRequest, map[string]string{"error": "categoryId 必须是非负整数"})
			return
		}
		if parsed > 0 {
			categoryID = &parsed
		}
	}

	page := 1
	if raw := query.Get("page"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			writeJSON(response, http.StatusBadRequest, map[string]string{"error": "page 必须是非负整数"})
			return
		}
		page = parsed
	}
	pageSize := 20
	if raw := query.Get("pageSize"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 0 {
			writeJSON(response, http.StatusBadRequest, map[string]string{"error": "pageSize 必须是非负整数"})
			return
		}
		pageSize = parsed
	}
	// 调 SearchNotes
	result, err := handler.notesStore.SearchNotes(store.SearchOptions{
		Query:      q,
		CategoryID: categoryID,
		Page:       page,
		PageSize:   pageSize,
	})


	//notes, err := handler.notesStore.ListNotes()
	if err != nil {
		writeJSON(response, http.StatusInternalServerError, map[string]string{"error": "查询笔记失败"})
		return
	}

	writeJSON(response, http.StatusOK, result)
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
		Title      string `json:"title"`
		Content    string `json:"content"`
		CategoryID *int64 `json:"categoryId"`
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
		Title:      input.Title,
		Content:    input.Content,
		CategoryID: input.CategoryID,
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

// importBatch 处理 POST /api/notes/import：批量创建笔记
// 状态码：201 全成功 / 207 部分成功 / 400 全失败 / 500 DB 错
// （不用 import 这个名字因为是 Go 关键字，handler 方法名避开）
func (handler notesHandler) importBatch(response http.ResponseWriter, request *http.Request) {
	//  解析 JSON body 为 []store.ImportNoteInput
	var input struct {
		Notes []store.ImportNoteInput `json:"notes"`
	}
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeJSON(response, http.StatusBadRequest, map[string]string{"error": "请求体必须是包含 notes 的 JSON"})
		return
	}
	// 调 store
	result, err := handler.notesStore.ImportNotes(input.Notes)
	if err != nil {
		writeJSON(response, http.StatusInternalServerError, map[string]string{"error": "导入笔记失败"})
		return
	}

	// 按结果选状态码
	// 三分支：全成功 / 全失败 / 部分成功
	// 用 switch + case 比 if/else if 链更清晰地表达"互斥三态"
	status := http.StatusMultiStatus
	switch {
	case result.Failed == 0:
		status = http.StatusCreated // 201
	case result.Imported == 0:
		status = http.StatusBadRequest // 400
	default:
		status = http.StatusMultiStatus // 207
	}
	writeJSON(response, status, result)

}

func writeJSON(response http.ResponseWriter, status int, body any) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(body)
}
