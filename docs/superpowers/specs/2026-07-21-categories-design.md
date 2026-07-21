# 2026-07-21 Phase B：分类（Categories）落地设计

## 目标

把 `docs/data-model.md` 里早就定义、但一直未落地的 `categories` 表启用，让笔记可以按分类组织、列表可按分类筛选。

## 现状（2026-07-21 已就绪）

- `categories` 表在 `backend/internal/store/migration.go` 已建好：
  - `id` / `workspace_id`（外键 ON DELETE CASCADE） / `name` / `created_at`
  - `UNIQUE (workspace_id, name)` —— 同工作空间内分类名唯一
- `notes.category_id` 列在 `migration.go` 已建好：`REFERENCES categories(id) ON DELETE SET NULL` —— 删 category，关联 note 的 category_id 自动 NULL（笔记保留）
- `store.Note.CategoryID *int64` 字段已存在（`json:"categoryId,omitempty"`），前端 `Note.categoryId?: number` 也已存在
- **但**：没有任何 categories 业务代码；CreateNote/UpdateNote/ListNotes 都没接 categoryId

## 范围

### 后端（Ray 写，diff 教学）

**新增** `backend/internal/store/category.go`：
- `Category` 类型（带 json tag）
- `CreateCategory(name string) (Category, error)`
- `ListCategories() ([]Category, error)`
- 教学点：跟 note.go 的 `defaultWorkspaceID()` 配合用事务或单 SQL 都可以

**改** `backend/internal/store/note.go`：
- `CreateNoteInput` / `UpdateNoteInput` 加 `CategoryID *int64` 字段
- `CreateNote` / `UpdateNote` SQL 加 `category_id` 字段（NULL → nil 走 ON DELETE SET NULL 安全路径）
- 新增 `ListNotesByCategory(categoryID int64) ([]Note, error)`：WHERE category_id = ? 过滤
- 教学点：外键约束 + ON DELETE SET NULL + JOIN 分类名

**改** `backend/internal/store/note_test.go`（测试我先写，Ray 跑红 + 改）：

**新增** `backend/internal/httpapi/category.go`：
- `categoryHandler` + `create` + `list` 两个方法
- 教学点：跟 notesHandler 模板几乎一样，照葫芦画瓢

**改** `backend/internal/httpapi/notes.go`：
- `create` / `update` 的 JSON struct 加 `CategoryID *int64 \`json:"categoryId"\``
- `list` 保持不变（按分类筛选走新接口 `ListNotesByCategory`）

### 前端（我直接写完整代码）

- `frontend/src/api/types.ts` 加 `Category` 类型
- `frontend/src/api/categories.ts` 新建：`listCategories()` / `createCategory()`
- `frontend/src/views/NoteEditView.vue`：
  - 顶部加 `<el-select v-model="categoryId">`，options 从 `listCategories()` 拿
  - "无分类"用 `el-option :value="null"`
  - 提交时把 `categoryId` 带进 CreateNoteInput / UpdateNoteInput
- `frontend/src/views/NoteListView.vue`：
  - 顶部加 `<el-select v-model="filterCategoryId">` 分类筛选
  - 选 "全部" → `listNotes()`；选某分类 → `listNotesByCategory(id)`（新接口）
- `frontend/src/views/NoteDetailView.vue`：
  - 标题下加一行小字显示分类（如果有），没分类就不显示

## API 设计

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/categories` | 列出当前工作空间所有分类 |
| POST | `/api/categories` | 创建分类（name 必填、唯一） |
| GET | `/api/notes?categoryId=N` | 新增 query 参数；不传就是全部 |

> 走 query 参数（不是独立 `/api/categories/N/notes`）的好处：跟前端 `listNotes()` 共用一套错误处理和兜底逻辑。

## 关键决策点（✅ 2026-07-21 Ray 拍板全选 A）

### 决策 1：categories UX 模式 → **A. 预定义模式**
先去 NoteEditView 下拉里加 "新建分类" 创建，再用。

### 决策 2：note 列表带 category 名字 → **A. 后端 JOIN 一次返回**
`Note` 加 `CategoryName *string` 字段，SQL `LEFT JOIN categories ON n.category_id = c.id`。
前端不需 listCategories 拼装，详情页/列表/编辑页共用同一份数据。

### 决策 3：删除 category 行为 → **A. 弹窗确认 + 警告**
"删除分类 X 之后，关联 N 篇笔记会变成无分类，是否继续？"
（关联篇数 = 后端 ListNotesByCategory 计数后返回；Phase B 第一版可先不传具体篇数，只警告）

## 教学点（Phase B Go 重点）

1. **外键约束在 Go 层的体现**：`INSERT INTO notes` 时如果 `category_id` 引用一个不存在的 id，SQLite 会返回 FOREIGN KEY constraint failed 错——错误是 SQL 层抛的，Go 层用 `errors.Is` 包不出来
2. **`ON DELETE SET NULL` 的安全路径**：删除 category 不需要先解绑 note，DB 自动处理
3. **JOIN 查询**：`SELECT n.*, c.name FROM notes n LEFT JOIN categories c ON n.category_id = c.id`
4. **指针类型在 SQL Scan**：`CategoryID *int64` 跟 `CategoryName *string` 都要走 `sql.NullInt64` / `sql.NullString` 中转
5. **多 workspace 下的数据隔离**：ListCategories / ListNotesByCategory 都要带 `WHERE workspace_id = ?` 跟现有 note.go 保持一致

## 实施步骤

1. 我写 design 文档（**本文件**）+ Ray 拍板 3 个决策
2. 我写 `category_test.go` 第一个失败测试 `TestStoreCreatesAndListsCategories`
3. Ray 跑红 + 读测试 + 我贴 diff 教 `Category` 类型 + `CreateCategory` / `ListCategories`
4. Ray 跑绿 + 跟着我贴的 diff 写 `ListNotesByCategory` 测试（**Ray 写**第一个测试）
5. 接着 `CreateNote` / `UpdateNote` 支持 categoryId（我写测试、Ray 改 store）
6. 后端 handler（Ray 写 category handler + 改 notes handler）
7. 前端我接（types / API / 3 个 View）
8. 端到端验证 + commit + push
