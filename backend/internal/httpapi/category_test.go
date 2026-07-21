package httpapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/XiaoMoDern/dev-knowledge-base/backend/internal/store"
)

// fakeCategoryStore 是 CategoryStore 接口的 mock 实现。
// 跟 fakeNotesStore 同款模式：测试时不连真 db，只验 handler 行为。
type fakeCategoryStore struct {
	categories []store.Category
}

func (fake *fakeCategoryStore) CreateCategory(name string) (store.Category, error) {
	category := store.Category{
		ID:        int64(len(fake.categories) + 1),
		Name:      name,
		CreatedAt: "2026-07-21T00:00:00Z",
	}
	fake.categories = append(fake.categories, category)
	return category, nil
}

func (fake *fakeCategoryStore) ListCategories() ([]store.Category, error) {
	return fake.categories, nil
}

func TestCategoryHandlerCreatesAndListsCategories(t *testing.T) {
	categoriesStore := &fakeCategoryStore{}
	handler := NewHandler(nil, categoriesStore)

	// 1. POST /api/categories 创建分类
	createRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/categories",
		bytes.NewBufferString(`{"name":"Go"}`),
	)
	createResponse := httptest.NewRecorder()
	handler.ServeHTTP(createResponse, createRequest)

	if createResponse.Code != http.StatusCreated {
		t.Fatalf("create status code = %d, want %d", createResponse.Code, http.StatusCreated)
	}

	var created store.Category
	if err := json.NewDecoder(createResponse.Body).Decode(&created); err != nil {
		t.Fatalf("decode created category: %v", err)
	}
	if created.Name != "Go" {
		t.Fatalf("created name = %q, want %q", created.Name, "Go")
	}

	// 2. GET /api/categories 列表
	listRequest := httptest.NewRequest(http.MethodGet, "/api/categories", nil)
	listResponse := httptest.NewRecorder()
	handler.ServeHTTP(listResponse, listRequest)

	if listResponse.Code != http.StatusOK {
		t.Fatalf("list status code = %d, want %d", listResponse.Code, http.StatusOK)
	}

	var body struct {
		Items []store.Category `json:"items"`
	}
	if err := json.NewDecoder(listResponse.Body).Decode(&body); err != nil {
		t.Fatalf("decode listed categories: %v", err)
	}
	if len(body.Items) != 1 {
		t.Fatalf("listed category count = %d, want 1", len(body.Items))
	}
	if body.Items[0].ID != created.ID {
		t.Fatalf("listed category ID = %d, want %d", body.Items[0].ID, created.ID)
	}
}

func TestCategoryHandlerRejectsBlankName(t *testing.T) {
	handler := NewHandler(nil, &fakeCategoryStore{})

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/categories",
		bytes.NewBufferString(`{"name":"   "}`),
	)
	response := httptest.NewRecorder()

	handler.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status code = %d, want %d", response.Code, http.StatusBadRequest)
	}
}
