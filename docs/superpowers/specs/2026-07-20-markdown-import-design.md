# Markdown 批量导入设计

## 目标

为已有笔记增加批量导入 .md 文件的能力。前后端分工明确：

- **前端（我维护）**：用 File API 读本地 .md 文件，解析 YAML front matter，转成 JSON 批量发给后端
- **后端（Ray学）**：接 JSON、事务批量 INSERT、错误聚合

不走 multipart 上传（前端处理文件，后端只接结构化数据）——这样后端不用碰 YAML 解析，专注于 Go 的事务、批量、错误处理这三个教学点。

## API

### `POST /api/notes/import`

请求（`application/json`）：

```json
{
  "notes": [
    { "title": "foo", "content": "bar" },
    { "title": "baz", "content": "qux" }
  ]
}
```

成功响应 **201**（全部成功）：

```json
{
  "imported": 2,
  "failed": 0,
  "items": [{ "id": 1, "title": "foo", "...": "..." }, { "id": 2, "title": "baz", "...": "..." }]
}
```

部分成功响应 **207**（Multi-Status）：

```json
{
  "imported": 1,
  "failed": 1,
  "items": [{ "id": 1, "title": "foo", "...": "..." }],
  "errors": [
    { "index": 1, "title": "", "reason": "title 不能为空" }
  ]
}
```

全部失败响应 **400**（一条都没插入）：

```json
{
  "imported": 0,
  "failed": 2,
  "items": [],
  "errors": [{ "index": 0, "title": "", "reason": "..." }, { "index": 1, "title": "", "reason": "..." }]
}
```

意外错误 **500**（事务失败 / DB 错误）：

```json
{ "error": "导入失败" }
```

## 数据契约

`ImportNoteInput`：

```go
type ImportNoteInput struct {
    Title   string
    Content string
}
```

`ImportResult`：

```go
type ImportResult struct {
    Imported int                // 成功数
    Failed   int                // 失败数
    Items    []Note             // 成功导入的完整 Note
    Errors   []ImportError      // 失败项
}

type ImportError struct {
    Index  int    // 在请求数组中的位置（0-based）
    Title  string // 当时尝试导入的 title（可能为空）
    Reason string // 失败原因
}
```

## 调用链

```text
HTTP POST /api/notes/import
  -> ServeMux 路由
  -> notesHandler.import
  -> 解析 JSON body 为 []ImportNoteInput
  -> 业务校验（title 非空）
  -> NotesStore.ImportNotes
  -> Store.ImportNotes（事务）
  -> 逐条 INSERT（每条读 LastInsertId + SELECT 完整行）
  -> 错误聚合
  -> HTTP 状态码：201 / 207 / 400 / 500
```

## 实现顺序

1. **TDD Red（Ray）**：先在 `internal/store/note_test.go` 写两个失败测试
   - `TestStoreImportsNotes`：3 条全部合法，期望 Imported=3 / Failed=0 / Items 长度 3
   - `TestStoreImportSkipsInvalidNotes`：3 条中 1 条 title 空，期望 Imported=2 / Failed=1 / Errors 含 index=1
2. **存储实现（Ray）**：`Store.ImportNotes` 用 `db.Begin()` 开事务；事务里遍历 INSERT；commit；defer rollback 兜底
3. **接口扩展（Ray）**：`NotesStore` 接口加 `ImportNotes`；`fakeNotesStore` 补实现（append 进切片即可）
4. **HTTP 实现（Ray）**：`notesHandler.import` 解析 body，按 Imported/Failed 数量返回 201/207/400
5. **HTTP 路由（Ray）**：`router.HandleFunc("POST /api/notes/import", notes.import)`
6. **前端 API（我）**：`src/api/notes.ts` 加 `importNotes(input: ImportNoteInput[]): Promise<ImportResult>`
7. **前端 UI（我）**：NoteListView header 加"导入"按钮 → 弹 `<el-dialog>` 选本地 .md 文件 → 用 File API 读内容 + 简单 YAML front matter 解析 → 调 `importNotes` → `<el-result>` 显示 X 成功 Y 失败
8. **手动验证**：浏览器导入 2-3 个 .md 文件，验证事务行为（一个 title 为空不影响其他）

## 明确不做

- 不做单文件 import（必须批量，避免多 API）
- 不做导入进度条（教学项目不值得；事务应该秒级完成）
- 不做 .md 文件去重（按内容哈希）——导入永远是创建新笔记
- 不支持非 .md 文件（zip 包、二进制等）
- 不做 markdown 渲染（与导入无关，单独阶段）

## 教学点（Ray写后端时遇到）

- **`database/sql.Tx` 事务**：用 `db.Begin()` 拿到 `*Tx`；成功后 `tx.Commit()`，失败 `tx.Rollback()`
- **defer Rollback 的安全模式**：事务开头先 `defer tx.Rollback()`，Commit 成功后 Rollback 是 no-op（已 commit 的 tx 不能 rollback）——这样写最稳，不用手动分支
- **partial success 模式**：业务校验失败不入事务；只有真 SQL 错才回滚；两者分开返回
- **错误聚合**：定义 `ImportResult` 结构，把 per-item 错误装进 slice 而不是返回第一个错
- **HTTP 207 Multi-Status**：HTTP 协议本身支持多状态响应，partial success 场景下用 207 比 200 更准确（也避免和 201"已创建"语义混淆）
- **JSON 字段缺失**：客户端可能不传某些字段，要用指针 + omitempty 区分"没传"和"传了空"

## 前端教学点（我写时自己 review）

- `File.text()` Promise API：浏览器 File 对象直接 `.text()` 读文本，无需 FileReader
- 简单 front matter 解析：自己写 split + trim 即可（5-10 行），不引第三方库
- 批量上传：用 `<el-upload :auto-upload="false" multiple>` 让用户选多个文件，前端攒好一次性发
- `<el-result>` 组件：导入完成显示成功 / 部分成功 / 失败 + 错误详情
