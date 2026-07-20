# 笔记 API 最小闭环 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让本地用户能够通过 HTTP 创建笔记并查询当前默认工作空间中的笔记，验证 SQLite、存储层、路由和 JSON 响应的完整链路。

**Architecture:** 保持 `store` 负责 SQLite 数据访问、`httpapi` 负责 HTTP 校验与 JSON、`cmd/server` 负责打开数据库并注入依赖。默认用户和工作空间继续只由 `store.Open()` 创建；笔记接口不接受工作空间或用户 ID，始终使用该默认工作空间。

**Tech Stack:** Go 1.26、标准库 `net/http`、`encoding/json`、`database/sql`、`modernc.org/sqlite`、SQLite。

---

## 文件结构

- 修改：`backend/internal/store/sqlite_test.go`，覆盖默认数据的重复打开行为。
- 新建：`backend/internal/store/note.go`，定义笔记数据结构和 SQLite 读写。
- 新建：`backend/internal/store/note_test.go`，覆盖笔记存储读写。
- 修改：`backend/internal/httpapi/health.go`，让路由器接收笔记存储依赖并注册笔记路由。
- 新建：`backend/internal/httpapi/notes.go`，处理创建与查询笔记的 HTTP 请求。
- 新建：`backend/internal/httpapi/notes_test.go`，使用内存假仓储验证 API 契约。
- 修改：`backend/internal/httpapi/health_test.go`，传入空依赖后继续验证健康检查。
- 修改：`backend/cmd/server/main.go`，创建 `data/dev-notes.db` 的父目录、打开存储并注入 HTTP 路由器。
- 修改：`backend/cmd/server/main_test.go`，验证服务器仍使用本地开发地址并能接收依赖。
- 修改：`docs/project-playbook.md`、`docs/learning-log.md`，记录默认数据与笔记 API 阶段。
- 新建：`docs/knowledge/go/006-store-and-http-dependency-injection.md`，解释接口注入和存储层边界。

### Task 1: 补齐默认数据的幂等性测试

**Files:**
- Modify: `backend/internal/store/sqlite_test.go`

- [ ] **Step 1: 添加重复打开同一数据库的测试**

在 `sqlite_test.go` 末尾添加：

```go
func TestOpenDoesNotDuplicateDefaultWorkspace(t *testing.T) {
	databasePath := filepath.Join(t.TempDir(), "dev-notes.db")

	firstStore, err := Open(databasePath)
	if err != nil {
		t.Fatalf("open database for the first time: %v", err)
	}
	if err := firstStore.Close(); err != nil {
		t.Fatalf("close first database: %v", err)
	}

	secondStore, err := Open(databasePath)
	if err != nil {
		t.Fatalf("open database for the second time: %v", err)
	}
	t.Cleanup(func() {
		if err := secondStore.Close(); err != nil {
			t.Errorf("close second database: %v", err)
		}
	})

	var userCount int
	var workspaceCount int
	var membershipCount int

	if err := secondStore.db.QueryRow(`SELECT COUNT(*) FROM users WHERE username = ?`, defaultUsername).Scan(&userCount); err != nil {
		t.Fatalf("count default users: %v", err)
	}
	if err := secondStore.db.QueryRow(`SELECT COUNT(*) FROM workspaces WHERE name = ?`, defaultWorkspaceName).Scan(&workspaceCount); err != nil {
		t.Fatalf("count default workspaces: %v", err)
	}
	if err := secondStore.db.QueryRow(`SELECT COUNT(*) FROM workspace_members WHERE role = 'owner'`).Scan(&membershipCount); err != nil {
		t.Fatalf("count default memberships: %v", err)
	}

	if userCount != 1 {
		t.Fatalf("default user count = %d, want 1", userCount)
	}
	if workspaceCount != 1 {
		t.Fatalf("default workspace count = %d, want 1", workspaceCount)
	}
	if membershipCount != 1 {
		t.Fatalf("default membership count = %d, want 1", membershipCount)
	}
}
```

- [ ] **Step 2: 运行测试确认已有默认数据实现保持幂等**

Run: `go test ./internal/store -run TestOpenDoesNotDuplicateDefaultWorkspace -v`

Expected: `PASS`。当前 `ON CONFLICT DO NOTHING` 和已有工作空间查询应使第二次打开不新增默认数据。

### Task 2: 添加笔记存储读写

**Files:**
- Create: `backend/internal/store/note.go`
- Create: `backend/internal/store/note_test.go`

- [ ] **Step 1: 写入笔记读写的失败测试**

创建 `note_test.go`：

```go
package store

import (
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
```

- [ ] **Step 2: 运行测试确认当前缺少笔记存储 API**

Run: `go test ./internal/store -run TestStoreCreatesAndListsNotes -v`

Expected: 编译失败，提示 `CreateNoteInput`、`CreateNote` 和 `ListNotes` 未定义。

- [ ] **Step 3: 实现最小笔记存储 API**

创建 `note.go`：

```go
package store

import (
	"database/sql"
	"fmt"
	"time"
)

type Note struct {
	ID         int64  `json:"id"`
	CategoryID *int64 `json:"categoryId,omitempty"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Visibility string `json:"visibility"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

type CreateNoteInput struct {
	Title   string
	Content string
}

func (store *Store) CreateNote(input CreateNoteInput) (Note, error) {
	workspaceID, err := store.defaultWorkspaceID()
	if err != nil {
		return Note{}, err
	}

	now := time.Now().UTC().Format(time.RFC3339)
	result, err := store.db.Exec(`
		INSERT INTO notes (workspace_id, title, content, visibility, created_at, updated_at)
		VALUES (?, ?, ?, 'private', ?, ?)
	`, workspaceID, input.Title, input.Content, now, now)
	if err != nil {
		return Note{}, fmt.Errorf("create note: %w", err)
	}

	noteID, err := result.LastInsertId()
	if err != nil {
		return Note{}, fmt.Errorf("read created note ID: %w", err)
	}

	return Note{
		ID:         noteID,
		Title:      input.Title,
		Content:    input.Content,
		Visibility: "private",
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func (store *Store) ListNotes() ([]Note, error) {
	workspaceID, err := store.defaultWorkspaceID()
	if err != nil {
		return nil, err
	}

	rows, err := store.db.Query(`
		SELECT id, category_id, title, content, visibility, created_at, updated_at
		FROM notes
		WHERE workspace_id = ?
		ORDER BY updated_at DESC, id DESC
	`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}
	defer rows.Close()

	notes := make([]Note, 0)
	for rows.Next() {
		var note Note
		var categoryID sql.NullInt64

		if err := rows.Scan(
			&note.ID,
			&categoryID,
			&note.Title,
			&note.Content,
			&note.Visibility,
			&note.CreatedAt,
			&note.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan note: %w", err)
		}
		if categoryID.Valid {
			note.CategoryID = &categoryID.Int64
		}

		notes = append(notes, note)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notes: %w", err)
	}

	return notes, nil
}

func (store *Store) defaultWorkspaceID() (int64, error) {
	var workspaceID int64
	err := store.db.QueryRow(`
		SELECT w.id
		FROM workspaces w
		JOIN users u ON u.id = w.owner_user_id
		WHERE u.username = ? AND w.name = ?
		LIMIT 1
	`, defaultUsername, defaultWorkspaceName).Scan(&workspaceID)
	if err != nil {
		return 0, fmt.Errorf("find default workspace: %w", err)
	}

	return workspaceID, nil
}
```

- [ ] **Step 4: 运行存储层测试**

Run: `go test ./internal/store -v`

Expected: `TestOpenCreatesDatabaseFile`、核心表测试、默认数据测试、幂等测试和 `TestStoreCreatesAndListsNotes` 全部通过。

### Task 3: 暴露创建与查询笔记的 HTTP API

**Files:**
- Modify: `backend/internal/httpapi/health.go`
- Create: `backend/internal/httpapi/notes.go`
- Modify: `backend/internal/httpapi/health_test.go`
- Create: `backend/internal/httpapi/notes_test.go`

- [ ] **Step 1: 将健康检查测试改为显式传入无笔记依赖**

在 `health_test.go` 中替换路由器创建行：

```go
NewHandler(nil).ServeHTTP(response, request)
```

- [ ] **Step 2: 编写 HTTP 契约测试**

创建 `notes_test.go`：

```go
package httpapi

import (
	"bytes"
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
```

- [ ] **Step 3: 运行接口测试确认当前路由器签名和笔记路由尚未实现**

Run: `go test ./internal/httpapi -run TestNotesHandler -v`

Expected: 编译失败，提示 `NewHandler` 参数不匹配，且尚未存在笔记处理能力。

- [ ] **Step 4: 实现路由器依赖注入与笔记处理器**

将 `health.go` 改为：

```go
package httpapi

import (
	"net/http"

	"github.com/XiaoMoDern/dev-knowledge-base/backend/internal/store"
)

type NotesStore interface {
	CreateNote(store.CreateNoteInput) (store.Note, error)
	ListNotes() ([]store.Note, error)
}

func NewHandler(notesStore NotesStore) http.Handler {
	router := http.NewServeMux()
	notes := notesHandler{notesStore: notesStore}

	router.HandleFunc("GET /api/health", healthHandler)
	router.HandleFunc("GET /api/notes", notes.list)
	router.HandleFunc("POST /api/notes", notes.create)

	return router
}

func healthHandler(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusOK)

	_, _ = response.Write([]byte(`{"status":"ok"}`))
}
```

创建 `notes.go`：

```go
package httpapi

import (
	"encoding/json"
	"net/http"
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

func writeJSON(response http.ResponseWriter, status int, body any) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(status)
	_ = json.NewEncoder(response).Encode(body)
}
```

- [ ] **Step 5: 运行 HTTP API 测试**

Run: `go test ./internal/httpapi -v`

Expected: 健康检查仍返回 `200`，笔记创建返回 `201`，空标题返回 `400`，查询返回含 `items` 数组的 `200` JSON。

### Task 4: 在服务入口打开 SQLite 并注入 API

**Files:**
- Modify: `backend/cmd/server/main.go`
- Modify: `backend/cmd/server/main_test.go`

- [ ] **Step 1: 改写服务器构造测试**

将 `main_test.go` 改为：

```go
package main

import "testing"

func TestNewServerUsesLocalDevelopmentAddress(t *testing.T) {
	server := newServer(nil)

	if server.Addr != "127.0.0.1:8181" {
		t.Fatalf("server address = %q, want %q", server.Addr, "127.0.0.1:8181")
	}

	if server.Handler == nil {
		t.Fatal("server handler must not be nil")
	}
}
```

- [ ] **Step 2: 运行服务器测试确认构造函数签名尚未更新**

Run: `go test ./cmd/server -run TestNewServerUsesLocalDevelopmentAddress -v`

Expected: 编译失败，提示调用 `newServer(nil)` 时参数数量不匹配。

- [ ] **Step 3: 实现数据库目录创建、存储打开与注入**

将 `main.go` 改为：

```go
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/XiaoMoDern/dev-knowledge-base/backend/internal/httpapi"
	"github.com/XiaoMoDern/dev-knowledge-base/backend/internal/store"
)

const databasePath = "data/dev-notes.db"

func newServer(notesStore httpapi.NotesStore) *http.Server {
	return &http.Server{
		Addr:    "127.0.0.1:8181",
		Handler: httpapi.NewHandler(notesStore),
	}
}

func openStore() (*store.Store, error) {
	if err := os.MkdirAll(filepath.Dir(databasePath), 0o755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	database, err := store.Open(databasePath)
	if err != nil {
		return nil, fmt.Errorf("open application store: %w", err)
	}

	return database, nil
}

func main() {
	database, err := openStore()
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	server := newServer(database)

	log.Printf("server listening on http://%s", server.Addr)

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
```

- [ ] **Step 4: 运行后端完整测试**

Run: `go test ./...`

Expected: `cmd/server`、`internal/httpapi` 和 `internal/store` 全部为 `ok`。

- [ ] **Step 5: 执行本地服务冒烟验证**

Run: `go run ./cmd/server`

Expected: 输出 `server listening on http://127.0.0.1:8181`，并在 `backend/data/` 下生成 SQLite 文件。

在另一终端运行：

```powershell
Invoke-RestMethod -Method Post -Uri http://127.0.0.1:8181/api/notes -ContentType 'application/json' -Body '{"title":"第一篇笔记","content":"通过 API 写入 SQLite。"}'
Invoke-RestMethod -Method Get -Uri http://127.0.0.1:8181/api/notes
```

Expected: 第一条命令返回含正整数 `id` 的笔记；第二条命令返回 `items` 数组，且其中包含“第一篇笔记”。

### Task 5: 同步项目过程与学习文档

**Files:**
- Modify: `docs/project-playbook.md`
- Modify: `docs/learning-log.md`
- Create: `docs/knowledge/go/006-store-and-http-dependency-injection.md`

- [ ] **Step 1: 在项目实施记录追加两个已验证阶段**

在 `project-playbook.md` 末尾追加：

```markdown
## Step 2.2：初始化默认本地工作空间
- 日期：2026-07-16
- 目标：让新 SQLite 数据库具备第一版所需的本地用户、默认工作空间和 owner 成员关系。
- 实现：`Open()` 在迁移完成后调用 `ensureDefaults()`；初始化过程使用事务，并通过唯一约束和冲突忽略保证重复打开同一数据库不会生成重复默认数据。
- 验证：`go test ./internal/store -v` 通过默认工作空间创建和重复打开计数测试。
- 结果：完成。第一版所有笔记均可归属到“我的知识库”。
- 学习点：初始化数据和建表迁移一样必须具备幂等性，否则应用重启会产生重复记录。

## Step 2.3：笔记 API 最小闭环
- 日期：2026-07-16
- 目标：跑通 HTTP 请求、参数校验、SQLite 写入和 SQLite 查询的完整路径。
- 实现：存储层提供默认工作空间中的笔记创建与列表查询；HTTP 层提供 `POST /api/notes` 与 `GET /api/notes`；服务入口打开 `data/dev-notes.db` 并注入存储依赖。
- 验证：`go test ./...` 通过；本地运行服务后创建笔记并查询列表成功。
- 结果：完成。后端已具备第一个持久化业务 API。
- 学习点：服务入口负责组装依赖，HTTP 层只处理请求协议，存储层只处理数据访问。
```

- [ ] **Step 2: 在学习记录追加要点**

在 `learning-log.md` 末尾追加：

```markdown
## 2026-07-16：默认数据与笔记 API

- 默认用户和工作空间必须可重复初始化；唯一约束、查询已有记录和 `ON CONFLICT DO NOTHING` 可以避免应用重启重复插入。
- `store` 返回业务数据，`httpapi` 校验 JSON 和 HTTP 状态码，`main` 负责把两者组装起来；这叫依赖注入。
- `POST /api/notes` 返回 `201 Created`，空标题返回 `400 Bad Request`，`GET /api/notes` 返回 `{"items": [...]}`。
- SQLite 数据文件属于运行时数据，不应写进源代码或提交到 Git。
```

- [ ] **Step 3: 创建依赖注入知识文章**

创建 `docs/knowledge/go/006-store-and-http-dependency-injection.md`：

```markdown
---
id: go-store-http-dependency-injection
title: "Go 中的存储层、HTTP 层与依赖注入"
category: Go
tags:
  - Go
  - net/http
  - SQLite
  - 依赖注入
summary: "把 SQLite 访问留在 store，把 HTTP 协议留在 httpapi，由 main 组装依赖，能让业务代码更容易测试和演进。"
---

# Go 中的存储层、HTTP 层与依赖注入

## 一句话结论

HTTP 处理器不应该自己打开 SQLite；程序入口创建 `Store` 后把它传给处理器，处理器只通过接口调用笔记读写方法。

## 当前项目的调用链

```text
main
  -> store.Open("data/dev-notes.db")
  -> httpapi.NewHandler(database)
  -> POST /api/notes
  -> Store.CreateNote(...)
  -> SQLite notes 表
```

## 为什么分层

- `store` 只知道 SQL、事务和数据结构，不知道 URL、JSON 或 HTTP 状态码。
- `httpapi` 只知道请求体、参数校验和响应状态码，不知道 SQLite 的连接细节。
- `main` 是唯一负责创建数据库并把依赖交给 HTTP 层的位置。

这样测试 HTTP 路由时可以传入假仓储，测试 SQLite 时可以直接调用真实 `Store`，两类测试不会彼此依赖。

## 接口注入

```go
type NotesStore interface {
    CreateNote(store.CreateNoteInput) (store.Note, error)
    ListNotes() ([]store.Note, error)
}
```

`NotesStore` 描述 HTTP 层真正需要的能力，而不是要求测试必须启动 SQLite。生产环境传入 `*store.Store`，测试传入实现同一接口的假对象。

## HTTP 状态码

- `201 Created`：笔记已经成功创建。
- `400 Bad Request`：请求不是合法 JSON 或标题为空。
- `500 Internal Server Error`：存储层执行失败，服务端不向调用方暴露底层数据库错误。
```

- [ ] **Step 4: 检查文档和改动边界**

Run: `git diff --check`

Expected: 无输出且退出码为 `0`。

- [ ] **Step 5: 经用户明确授权后提交阶段性改动**

Run: `git add backend docs && git commit -m "feat: add notes API"`

Expected: Git 创建一个只包含笔记 API、相关测试和同步文档的提交；未授权时跳过此步骤。
