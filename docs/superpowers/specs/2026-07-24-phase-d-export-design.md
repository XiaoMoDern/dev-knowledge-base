# 2026-07-24 Phase D Markdown 导出 / 备份 设计文档

## 目标

让 dev-notebook 能把 notes 导出成 Markdown 文件：
- **单 note 导出**——详情页"导出"按钮，下单个 .md
- **批量导出 / 备份**——选分类或全部，导出 zip 包

让 dev-notebook 跨过"数据进得来、出不去"的阶段（之前只有"批量导入 .md"，没"导出"）。

---

## Go 教学点（Phase D 重点）

1. **`io.Writer` 接口流式输出**——`zip.Writer` / `os.File` / `http.ResponseWriter` 都实现 `io.Writer`，代码复用同一套 `io.Copy`
2. **`archive/zip` 标准库**——`zip.NewWriter(w)` → `Create(name)` → `Write([]byte)` → `Close()`，4 步流程
3. **流式 vs 一次性**——zip 不一次性加载全部 note 到内存，按 note 逐个 Create+Write，10w+ note 不爆
4. **HTTP `Content-Disposition`**——`w.Header().Set("Content-Disposition", "attachment; filename=...")` 触发浏览器下载
5. **SQLite 一致性快照**——导出时 `BEGIN IMMEDIATE` 事务隔离，避免导出过程中数据被修改
6. **`bufio.Writer` 包装**——zip writer 直接写 http response，每次 Write 都 flush，性能差；包 bufio 减少系统调用
7. **导出时的"游标"分页**——大量 note 不能一次 `SELECT *`，要 `LIMIT/OFFSET` 分批（学过的分页技能复用）
8. **元数据 frontmatter**——导出 .md 头部加 YAML（`---\ntitle: ...\ncategory: ...\ncreatedAt: ...\n---\n`），方便重新导入

---

## 第一步任务（Ray 的，按能力台阶）

### 阶段 1：单 note 导出（基础）

1. **我写完整测试** → Ray 跑、读、改
2. 2 个失败测试（TDD Red）：
   - `TestStoreGetNoteForExport`：取单 note 完整数据（含 categoryName）
   - `TestExportSingleNoteHandler`：GET `/api/notes/:id/export` 返 .md 文件流

### 阶段 2：批量导出 zip（进阶）

3. **我写完整测试** → Ray 跑、读、改
4. 2 个失败测试：
   - `TestExportAllNotesHandler`：GET `/api/notes/export?format=zip` 返 zip 流
   - `TestExportByCategoryHandler`：GET `/api/notes/export?categoryId=N&format=zip`

---

## API 计划

### 单 note 导出

```
GET /api/notes/:id/export
```

**Response**：
- `Content-Type: text/markdown; charset=utf-8`
- `Content-Disposition: attachment; filename="note-{id}-{title}.md"`
- Body：markdown 文本（含 frontmatter）

**示例输出**：

```markdown
---
id: 42
title: Go 1.26 新特性
category: Go
categoryId: 1
createdAt: 2026-07-20T11:30:00Z
updatedAt: 2026-07-22T15:00:00Z
---

# Go 1.26 新特性

正文内容...
```

### 批量导出 zip

```
GET /api/notes/export?categoryId=N&format=zip
GET /api/notes/export?format=zip    # 全部
```

**Response**：
- `Content-Type: application/zip`
- `Content-Disposition: attachment; filename="dev-notebook-backup-2026-07-24.zip"`
- Body：zip 流（按分类分文件夹，每个 note 一个 .md）

**zip 结构**：

```
dev-notebook-backup-2026-07-24.zip
├── README.md                          # 备份说明（导出时间 / 数量 / dev-notebook 版本）
├── uncategorized/
│   ├── note-1-未分类笔记.md
│   └── note-2-前端E2E.md
├── Go/
│   ├── note-3-Go基础.md
│   └── note-4-Go并发.md
└── 前端/
    ├── note-5-Vue3.md
    └── note-6-ElementPlus.md
```

### 范围参数

| Query | 行为 |
| --- | --- |
| 无参数 | 全部 note |
| `categoryId=N` | 单分类 note |
| `categoryId=0` | 未分类 note（哨兵值，跟前端一致） |

---

## 前端改动

### 详情页"导出"按钮

```vue
<el-button @click="exportNote" type="default" plain>
  导出 Markdown
</el-button>

<script setup lang="ts">
async function exportNote() {
  const url = `/api/notes/${note.id}/export`
  const a = document.createElement('a')
  a.href = url
  a.download = `${note.title}.md`
  a.click()
  ElMessage.success('已下载')
}
</script>
```

**优点**：浏览器原生下载，零依赖。

### 列表页"批量导出"按钮

```vue
<el-button @click="exportAll" type="default">
  批量导出
</el-button>

<script setup lang="ts">
async function exportAll() {
  const params = new URLSearchParams()
  if (selectedCategoryId.value !== undefined) {
    params.set('categoryId', selectedCategoryId.value.toString())
  }
  params.set('format', 'zip')
  const url = `/api/notes/export?${params}`
  window.open(url, '_blank')  // 浏览器触发下载
  ElMessage.success('正在准备备份...')
}
</script>
```

**优点**：后端流式返回，浏览器自动开始下载。

---

## 待 Ray 拍板的决策（5 个）

### 决策 1：导出范围参数

| 方案 | 优点 | 缺点 |
| --- | --- | --- |
| **A 全部 + categoryId + categoryId=0**（推荐） | 跟现有 search API 一致，前端代码复用 | 0 是哨兵值，需要约定 |
| B only categoryId | 简单 | "全部"语义不清 |
| C 加 `?all=true` flag | 语义清晰 | 多一个参数 |

**推荐 A**——跟前端 search API 复用，前端 CategorySidebar 选分类后直接传 `categoryId`。

### 决策 2：批量导出格式

| 方案 | 优点 | 缺点 | 教学价值 |
| --- | --- | --- | --- |
| **A zip**（推荐） | 跨平台、解压方便、保留目录结构 | 要学 archive/zip | 学 Go 标准库 zip |
| B tar.gz | Linux 友好 | Windows 用户不熟 | 学 archive/tar |
| C 多文件 .md（不打包） | 简单 | 1000 条 note = 1000 次下载，浏览器不友好 | — |
| D 单个 .md 拼接 | 最简单 | 失去单 note 独立性、重新导入要 split | — |

**推荐 A**——zip 是 Phase D 核心 Go 教学点（`archive/zip` + `io.Writer` 流式）。

### 决策 3：导出时的元数据

| 方案 | 优点 | 缺点 | 重新导入 |
| --- | --- | --- | --- |
| **A YAML frontmatter**（推荐） | 人类可读、markdown 编辑器识别、保留分类/时间 | 头部 5-8 行 | ✅ 直接重导 |
| B JSON sidecar（.md + .json） | 完整结构、嵌套字段友好 | 2 个文件难管理 | ✅ 读 json |
| C 纯 .md 不带元数据 | 最干净 | 失去 id/时间/分类 | ⚠️ 当成新 note |

**推荐 A**——frontmatter 是 markdown 生态标准，Obsidian / VSCode / Hugo 都识别。

### 决策 4：导出时数据库一致性

| 方案 | 优点 | 缺点 |
| --- | --- | --- |
| **A `BEGIN IMMEDIATE` 事务**（推荐） | 导出过程中数据不被修改，备份一致 | 导出期间其他写入阻塞（但 dev-notebook 单用户，无影响） |
| B 不加锁，导出过程中允许修改 | 不阻塞其他操作 | 备份可能"半新半旧"（note 1 是旧版本，note 2 是新版本） |
| C 导出前 `VACUUM INTO` 备份整个 DB | 最安全 | dev-notebook 不是这个语义，是"导出 notes"不是"备份 DB" |

**推荐 A**——dev-notebook 单用户，BEGIN IMMEDIATE 阻塞时长 = 导出耗时（几秒），无感知。

### 决策 5：批量导出的分批策略

| 方案 | 优点 | 缺点 |
| --- | --- | --- |
| **A 流式（不缓存全量）**（推荐） | 内存恒定，10w+ note 不爆 | 进度条难做 |
| B 全量加载到内存再打包 | 实现简单 | 1w+ note 内存压力 |
| C 分批查询（每 1000 条一批） | 折中 | 复杂度 ↑ |

**推荐 A**——跟"分页查询"是同款技能，每查 1000 条就 zip 写入 + flush。学过的 Phase C 分页技能复用。

---

## 数据流（zip 导出）

```
HTTP GET /api/notes/export?format=zip
    ↓
handler.exportNotes:
    w.Header().Set("Content-Type", "application/zip")
    w.Header().Set("Content-Disposition", "attachment; filename=...")
    ↓
    zipWriter := zip.NewWriter(w)        // 写 HTTP response
    bufWriter := bufio.NewWriter(zipWriter)  // 包一层减少 flush
    ↓
    db.BeginTx()                         // BEGIN IMMEDIATE
    ↓
    rows := ListNotesPaged(1000)         // 分批查（复用 Phase C）
    for rows.Next():
        note := scan(rows)
        if categoryFolder not in zip:
            zipWriter.Create("uncategorized/")
        header := makeFrontmatter(note)
        body := header + note.Content
        f, _ := zipWriter.Create(filepath.Join(category, "note-{id}-{title}.md"))
        f.Write([]byte(body))
    ↓
    zipWriter.Close()                    // 收尾
    db.Commit()                          // 事务提交
```

**关键**：
- `zipWriter` 写 HTTP response（`io.Writer` 多态）
- `bufio.Writer` 减少系统调用
- 分批查询（1000 条一批）避免内存爆
- `BEGIN IMMEDIATE` 保证一致性

---

## 注意事项（铁律重复提醒）

1. **跨边界类型必须有 json tag**——`ExportOptions` / `ExportRequest` 加 tag（Phase A 教训）
2. **改 SQL 加条件要同步检查参数列表**（铁律 3）
3. **handler 严格守门 vs store 兜底**——`categoryId < 0` 走 0，page 参数校验
4. **HTTP 响应是给用户的，go run 终端日志才是给开发者的**——导出失败要在 zip 里写 `errors.txt`，不只日志
5. **Rule of Three**——`ListNotesPaged` 在 store 层已存在，导出直接复用，不要复制
6. **commit/push 要 Ray 授权**（铁律 6）
7. **新功能先写设计文档**（铁律 7，本次文档）
8. **Element Plus 类型**——前端 `npm run dev` 第一次跑生成 `ElMessage` 的 d.ts（已踩过）
9. **测试要看完整 `-v` 输出**（铁律 2）—— zip 导出要验证文件能解压、内容正确

---

## 风险点

1. **大 note 内容**——单 note > 100MB？`bufio.NewWriterSize(f, 1MB)` 包一层
2. **特殊字符文件名**——note title 含 `/` `\` `:`？用 `sanitize(filename)` 替换
3. **导出过程中分类被删**——`categoryName` 缓存到 export 时，避免导出时查询
4. **浏览器下载大文件**——zip 10MB+？后端 `Content-Length` + `Content-Encoding: gzip`（可选）
5. **重复导出同名 note**——`note-{id}-{title}.md` 中 `id` 保证唯一性，title 重复无所谓

---

## Phase D 之后路线

- **E 测试进阶**——表驱动、testify、benchmark、SQLite 夹具
  - 教学点：写多组 input/output 表驱动测试、用 testify assert 简化断言、benchmark 测性能
- **F 并发批处理**（可选）——goroutine、channel、errgroup
  - 教学点：导出时多分类并发打包？用 errgroup 并发 + 错误传播
  - 进阶：race condition 检测（`go test -race`）

---

## 关联文件

- `backend/internal/store/note.go` — 新增 `GetNoteForExport` + 复用 `ListNotesPaged`
- `backend/internal/store/category.go` — 新增 `GetAllCategories`
- `backend/internal/httpapi/notes.go` — 新增 `exportNote` + `exportNotesZip` handler
- `backend/internal/httpapi/health.go` — `NotesStore` 接口加 `GetNoteForExport`
- `frontend/src/api/notes.ts` — 新增 `exportNote(id)` + `exportNotesZip(options)`
- `frontend/src/views/NoteDetailView.vue` — 加"导出 Markdown"按钮
- `frontend/src/views/NoteListView.vue` — 加"批量导出"按钮

---

## 第一步执行计划

1. **Ray 拍板 5 个决策**（预计 2 分钟）
2. **我更新设计文档**（按决策调整）→ 提交
3. **我写完整测试**（TDD Red，2-3 个失败测试）
4. **Ray 跑测试**（确认 Red）→ 我贴 store 层 diff
5. **Ray 改 store** → 跑全量测试 → handler diff → 改 handler
6. **E2E 验证**——浏览器导出 .md / zip，验证内容、解压、重新导入
7. **前端我写**——详情页导出按钮 + 列表页批量导出按钮
8. **commit + push**（Ray 授权后）
