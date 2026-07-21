package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/XiaoMoDern/dev-knowledge-base/backend/internal/store"
)

// categoryHandler 跟 notesHandler 同构——只少了 update/delete（分类只增删/查看，
// 改名字/合并是后续 Phase 的事）。
type categoryHandler struct {
	categoriesStore CategoryStore
}

// create 处理 POST /api/categories。
// 业务校验：name 必填且 trim 后非空。跟 notesHandler.create 同款模式。
func (handler categoryHandler) create(response http.ResponseWriter, request *http.Request) {
	var input struct {
		Name string `json:"name"`
	}

	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		writeJSON(response, http.StatusBadRequest, map[string]string{"error": "请求体必须是包含 name 的 JSON"})
		return
	}

	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		writeJSON(response, http.StatusBadRequest, map[string]string{"error": "name 不能为空"})
		return
	}
	category, err := handler.categoriesStore.CreateCategory(input.Name)
	if err != nil {
		writeJSON(response, http.StatusInternalServerError, map[string]string{"error": "创建分类失败"})
		return
	}

	writeJSON(response, http.StatusCreated, category)

}

// list 处理 GET /api/categories。
func (handler categoryHandler) list(response http.ResponseWriter, request *http.Request) {
	categories, err := handler.categoriesStore.ListCategories()
	if err != nil {
		writeJSON(response, http.StatusInternalServerError, map[string]string{"error": "查询分类失败"})
		return
	}

	writeJSON(response, http.StatusOK, struct {
		Items []store.Category `json:"items"`
	}{Items: categories})
}
