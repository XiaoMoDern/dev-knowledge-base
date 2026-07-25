# dev-notebook

一个本地化的笔记 / 知识库工具 — 用来给你自己写 markdown 笔记、做分类、搜索、批量导入导出。后端 Go + SQLite，前端 Vue 3 + TypeScript。

> 注意：本项目的**目的不是"做完 dev-notebook"**，是**"学 Go"**。dev-notebook 只是载体。如果你是来用工具的，它对你来说功能太简陋；如果你是来学习的，每个 Phase 都能学到 Go 的某个方向。

---

## 技术栈

| 层 | 技术 |
|---|---|
| 后端 | Go 1.26 + `modernc.org/sqlite`（纯 Go SQLite 驱动，不需要 CGO）|
| 前端 | Vue 3.5 + TypeScript 6 + Vite 8 + vue-router 4 + Element Plus 2.14（按需自动引入）|
| 数据库 | SQLite 单文件 `backend/data/dev-notes.db` |
| 渲染 | marked v18（markdown → HTML）+ DOMPurify v3（XSS 净化）|
| 测试 | Go `testing` 标准库 + 前端 Vitest（jsdom 环境）|

---

## 项目结构

```
dev-notebook/
├── backend/                # Go 服务
│   ├── cmd/server/         # main 入口
│   ├── internal/
│   │   ├── store/          # SQLite 数据访问层
│   │   └── httpapi/        # HTTP handler + NotesStore 接口
│   └── data/               # SQLite 文件（运行时生成）
├── frontend/               # Vue 3 SPA
│   ├── src/
│   │   ├── api/            # 后端 fetch 封装
│   │   ├── components/     # 通用组件（NoteCard / CategorySidebar 等）
│   │   ├── views/          # 页面级组件（NoteListView / NoteDetailView 等）
│   │   ├── utils/          # markdown 解析 + XSS 净化
│   │   └── styles/         # 设计 token（颜色 / 间距 / 阴影）
│   └── vite.config.ts      # 同时管 vite build 和 vitest
├── docs/                   # 学习文档
│   ├── project-playbook.md # 项目铁律 + 命令行手册 + 教学点
│   ├── learning-log.md     # 每日学习记录
│   ├── data-model.md       # 5 张表 schema
│   ├── knowledge/          # Go / Vue 知识库（编号文章）
│   └── superpowers/specs/  # 各功能的设计文档
├── README.md               # 你正在看的这一页
```

---

## 快速开始

### 1. 后端（占 8181 端口）

```bash
cd backend
go run ./cmd/server
# → server listening on http://127.0.0.1:8181
```

无需配置数据库 — 首次启动会自动 `mkdir -p data && 创建 dev-notes.db`。

健康检查：
```bash
curl http://127.0.0.1:8181/api/health
# → {"status":"ok"}
```

### 2. 前端（占 5173 端口）

```bash
cd frontend
npm install     # 第一次或拉了新依赖时
npm run dev
# → http://localhost:5173
```

前端开发服务器通过 Vite proxy 把 `/api/*` 转发到 `http://127.0.0.1:8181`（见 `frontend/vite.config.ts`），所以浏览器里看不到 CORS 问题。

> 生产部署的局限：当前 Go 服务**没有托管前端 dist**，只托管 API。生产环境需要：`go build` 出二进制，再额外跑 nginx / Go http.FileServer 等静态服务。这是 P2 待办，不是当前 milestone。

---

## 跑测试

### 后端（38 个测试）

```bash
cd backend
go test ./... -v
```

测试覆盖：`store` 层 18 个（含 CRUD / 搜索 / 分页 / 按 ID 查 / HasNote）+ `httpapi` 层 19 个（含 handler 守门 / 404 / 400 / 各种响应码）+ `cmd` 层 1 个。

**注意**：Go 的 `go build` 默认**不编译 `*_test.go` 文件**——测试文件的错误（缺 faker 方法、类型不匹配）只有 `go test` 或 `go vet` 才抓得到。

### 前端（5 个测试）

```bash
cd frontend
npm test            # 跑一次
npm run test:watch  # 文件改了就跑
```

测试覆盖：`src/utils/markdown.test.ts` — 验证 DOMPurify 净化掉 `<img onerror>` / `<script>` / `javascript:` 三类危险 HTML。

---

## 当前能力（P0-P1）

| 功能 | API |
|---|---|
| 列出 / 搜索 / 分页 notes | `GET /api/notes?q=&categoryId=&page=&pageSize=` |
| 按 ID 取单条 note | `GET /api/notes/{id}` |
| 创建 note | `POST /api/notes`（body: `{title, content, categoryId?}`）|
| 更新 note | `PUT /api/notes/{id}` |
| 删除 note | `DELETE /api/notes/{id}`（返 204）|
| 批量导入 .md 文件 | `POST /api/notes/import`（状态码 201 / 207 / 400）|
| 列出 / 创建分类 | `GET /api/categories` / `POST /api/categories` |
| Markdown 渲染 | 前端 `renderMarkdown(text)` —— marked + DOMPurify |

UI：dashboard 布局（顶部 + 侧边栏 + 卡片网格） + 详情页 sticky 操作栏 + 删除按钮（无 confirm 弹窗二次确认）。

---

## 当前边界（已知未做）

| 不做 | 为什么 |
|---|---|
| **多用户 / 登录** | dev-notebook 是单用户本地工具，没必要 |
| **多工作空间切换** | 现在只有"默认工作空间"，store 层有 workspace 概念但 UI 没暴露 |
| **全文搜索（FTS5）** | 当前用 SQLite `LIKE`，数据 < 万条够用 |
| **Markdown 导出** | `docs/superpowers/specs/2026-07-24-phase-d-export-design.md` 有设计稿，待实现 |
| **生产部署** | Vite proxy 仅开发期；Go 服务还没托管前端 dist |
| **服务端审计日志** | 单用户本地工具，YAGNI；如要审计就用 `log.Printf` 打 ID + 长度，不打内容 |
| **跨平台打包** | 没用 `embed` + 打包构建脚本 |

---

## 关键约定（项目铁律）

1. **跨边界类型必须有 json tag**——跟前端 TS types 1:1（教训：silently fail）
2. **测试"过了"要看完整 `-v` 输出**——不凭口头判断
3. **改 SQL 加条件要同步检查参数列表**——参数错位是 fail 的常见原因
4. **HTTP 响应是给用户的，`go run` 终端日志才是给开发者的**——敏感数据永远不进日志
5. **Rule of Three**：两处重复不抽
6. **commit / push 要 Ray 授权**（学员模式铁律）
7. **新功能先写设计文档**：`docs/superpowers/specs/YYYY-MM-DD-{feature}-design.md`

详见 [`docs/project-playbook.md`](docs/project-playbook.md)。

---

## 关联文档

| 想了解 ... | 看 ... |
|---|---|
| 项目铁律 + 命令行手册 + 各 Phase 教学点 | [`docs/project-playbook.md`](docs/project-playbook.md) |
| 数据库 5 张表 schema | [`docs/data-model.md`](docs/data-model.md) |
| 每日学习记录（已完成哪些 Phase / 学到什么）| [`docs/learning-log.md`](docs/learning-log.md) |
| 各功能设计文档 | [`docs/superpowers/specs/`](docs/superpowers/specs/) |
| Go 知识库（命名约定 / 接口设计 / 事务 / 并发）| [`docs/knowledge/go/`](docs/knowledge/go/) |
| Vue / TypeScript 知识库 | [`docs/knowledge/vue/`](docs/knowledge/vue/) |
