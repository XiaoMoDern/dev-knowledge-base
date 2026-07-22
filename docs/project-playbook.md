# Dev Knowledge Base 项目实施记录

## 项目目标
构建一个本地运行的学习知识库网站。
前端使用 Vue 3，后端使用 Go，数据保存到 SQLite。

## 技术环境
- Go: 1.26.4
- Node.js: 24.9.0
- npm: 11.6.0
- Git: 2.44.0.windows.1

## Step 0.2：初始化项目骨架
- 日期：2026-07-16
- 目标：创建项目目录、Git 仓库和项目文档。
- 执行内容：初始化 Git，创建 `docs`、`backend`、`frontend`。
- 结果：完成。已初始化 Git，并创建 `docs`、`backend`、`frontend` 及基础文档文件。
- 学习点：前后端分目录是为了让 Vue 和 Go 各自管理依赖、构建和测试；`docs` 保存可复用的开发流程。

## Step 0.3：关联远程仓库
- 日期：2026-07-16
- 目标：为本地仓库配置 GitHub 远程地址。
- 执行内容：添加 `origin`，地址为 `https://github.com/XiaoMoDern/dev-knowledge-base.git`。
- 结果：完成。本地 `fetch` 和 `push` 地址均已正确配置。
- 备注：当前网络无法连接 GitHub 的 443 端口，远程分支状态校验暂缓；这不影响本地开发和提交。
- 学习点：本地 Git 仓库与远程仓库是两层独立能力。配置 `origin` 不会自动上传任何内容。

## Step 1.1：创建 Go Module 和健康检查测试
- 日期：2026-07-16
- 目标：建立 Go 后端模块，并先定义健康检查接口的预期行为。
- 执行内容：初始化 Module `github.com/XiaoMoDern/dev-knowledge-base/backend`；创建 `cmd/server`、`internal/httpapi`；新增 `health_test.go`。
- 验证：运行 `go test ./...`，测试按预期失败，错误为 `undefined: NewHandler`。
- 实现：新增 `internal/httpapi/health.go`，使用 `http.NewServeMux` 注册 `GET /api/health`，返回 `{"status":"ok"}`。
- 验证：再次运行 `go test ./...`，`internal/httpapi` 测试通过。
- 结果：完成 Green 阶段。HTTP 路由器可以处理健康检查请求。
- 学习点：测试先描述“应该有什么行为”，失败证明当前项目确实缺少该能力；随后只实现让该测试通过的最小代码。

## Step 1.2：启动真实 HTTP 服务
- 日期：2026-07-16
- 目标：创建后端程序入口，并监听本地 8181 端口。
- 执行内容：新增 `cmd/server/main_test.go`，先验证 `newServer()` 返回的服务地址为 `127.0.0.1:8181` 且包含路由器；新增 `cmd/server/main.go`，使用 `httpapi.NewHandler()` 创建 HTTP 服务。
- 验证：`go test ./...` 通过；启动 `main.go` 后，浏览器访问 `http://127.0.0.1:8181/api/health` 返回 `{"status":"ok"}`。
- 结果：完成。Go 后端已具备可运行的真实 HTTP 服务。
- 学习点：`main.go` 是 Go 应用入口；`main` 包负责组装服务，`httpapi` 包负责处理 HTTP 请求，职责保持分离。

## Step 1.3：保存健康检查里程碑
- 日期：2026-07-16
- 目标：保存可运行后端和学习文档的稳定版本。
- 执行内容：提交并推送 Go Module、健康检查路由、服务入口、测试和文档。
- 结果：完成。
- 学习点：每个可运行、已验证的阶段都应独立提交，便于回退、对比和定位后续问题。

## Step 2.0：定义数据库模型
- 日期：2026-07-16
- 目标：在创建 SQLite 表之前确定未来可扩展的数据归属关系。
- 执行内容：新增 `docs/data-model.md`，定义用户、工作空间、成员关系、分类和笔记。
- 结果：完成设计，尚未创建数据库文件或表。
- 学习点：先确定实体关系和数据所有权，再写迁移代码，能显著降低后期扩展多用户与公开内容的成本。

## Step 2.1：实现 SQLite 核心表迁移
- 日期：2026-07-16
- 目标：使任意新 SQLite 数据库在首次打开时自动拥有核心表结构。
- Red：新增 `TestOpenCreatesCoreTables`，通过查询 SQLite 的 `sqlite_master` 要求存在 `users`、`workspaces`、`workspace_members`、`categories`、`notes`；迁移实现前测试按预期失败。
- 实现：新增 `internal/store/migration.go`，在一个事务中按外键依赖顺序执行 `CREATE TABLE IF NOT EXISTS`；`Open()` 启用 SQLite 外键检查并调用迁移。
- 验证：运行 `go mod tidy` 后，`go test ./...` 通过 `cmd/server`、`internal/httpapi` 和 `internal/store` 的全部测试。
- 结果：完成。测试使用的每个新 SQLite 文件都会自动创建五张核心表。
- 学习点：迁移应可重复执行、使用事务保证原子性，并用参数占位符而不是字符串拼接执行 SQL 查询。

## Step 2.2：初始化默认本地工作空间
- 日期：2026-07-16
- 目标：让新 SQLite 数据库具备第一版所需的本地用户、默认工作空间和 owner 成员关系。
- 实现：`Open()` 在迁移完成后调用 `ensureDefaults()`；初始化过程使用事务，并通过唯一约束和冲突忽略保证重复打开同一数据库不会生成重复默认数据。
- 验证：`go test ./internal/store -v` 通过默认工作空间创建和重复打开计数测试。
- 结果：完成。第一版所有笔记均可归属到"我的知识库"。
- 学习点：初始化数据和建表迁移一样必须具备幂等性，否则应用重启会产生重复记录。

## Step 2.3：笔记 API 最小闭环
- 日期：2026-07-16
- 目标：跑通 HTTP 请求、参数校验、SQLite 写入和 SQLite 查询的完整路径。
- 实现：存储层提供默认工作空间中的笔记创建与列表查询；HTTP 层提供 `POST /api/notes` 与 `GET /api/notes`；服务入口打开 `data/dev-notes.db` 并注入存储依赖。
- 验证：`go test ./...` 通过；本地运行服务后创建笔记并查询列表成功。
- 结果：完成。后端已具备第一个持久化业务 API。
- 学习点：服务入口负责组装依赖，HTTP 层只处理请求协议，存储层只处理数据访问。

## Step 2.4：删除笔记
- 日期：2026-07-16（设计） / 2026-07-17（接续收尾）
- 目标：为笔记 API 增加真正删除笔记的能力。
- 实现：存储层 `Store.DeleteNote(noteID)` 执行 `DELETE FROM notes WHERE id = ? AND workspace_id = ?`，用 `RowsAffected == 0` 判断不存在并返回 `sql.ErrNoRows`；`NotesStore` 接口增加 `DeleteNote(int64) error`，`fakeNotesStore` 用切片 `append(s[:i], s[i+1:]...)` 模拟删除；HTTP 处理器在 `sql.ErrNoRows` 时返回 404，非法 id 返回 400，成功返回 204。
- 验证：`go test ./... -v` 全部 PASS；通过 curl / Apifox 验证 204 / 400 / 404 三种状态码。
- 设计依据：`docs/superpowers/specs/2026-07-16-delete-note-design.md`，明确不��软删除、批量删除、权限系统。
- 状态码约定：204 No Content = 成功且无 body；前端/工具看到空 body 不等于"没响应"，状态码 204 才是成功信号。
- 教学点：删除接口故意不返回 body（REST 惯例）；Go 接口加方法后所有实现者必须同步补齐（`fakeNotesStore` 漏补 `DeleteNote` 会导致整个 httpapi 包编译失败）。

## Step 2.5：编辑笔记
- 日期：2026-07-17
- 目标：为笔记 API 增加编辑能力，引入 Go 指针接收者的真业务用法。
- 实现：新增 `UpdateNoteInput{Title, Content string}` 与 `Store.UpdateNote(noteID, input)`；存储层 `UPDATE notes SET title=?, content=?, updated_at=? WHERE id=? AND workspace_id=?`，`RowsAffected==0` 返回 `sql.ErrNoRows`，成功后再 `SELECT` 一次拿完整 Note（created_at、visibility、category_id）；`NotesStore` 接口加 `UpdateNote`，`fakeNotesStore` 直接修改切片元素；HTTP 处理器 `update` 走和 delete 一样的四分支（id 400 / title 空 400 / 找不到 404 / 成功 200）。
- 验证：`go test ./... -v` 全部 PASS（含 6 个 store 测试 + 11 个 httpapi 测试）；用 Apifox 手动验证创建→列表→更新→列表→删除的完整链路。
- 设计依据：`docs/superpowers/specs/2026-07-17-update-note-design.md`；明确不做部分更新（PATCH）、categoryId 修改、乐观锁、编辑历史。
- 学习点：
  - 指针接收者 `*Store` 在 Update 和 Delete 里都用到，原因是方法需要访问 store 内部状态（如 db）。
  - 接口签名修改要同步修改所有实现者：`*store.Store` 和 `fakeNotesStore` 都必须按接口签名调整。
  - SQL 修改占位符数量要同步修改参数列表：所有按 id 查的 SQL 都必须 `WHERE id = ? AND workspace_id = ?` 防止跨工作空间数据泄漏。
  - 调试 500 错误的正确姿势是看 `go run` 终端的服务端日志，不是看 Apifox 的 HTTP 响应——HTTP 响应是给用户的友好提示，日志才是给开发者的根因。
  - 状态码：Update 成功返回 200 + JSON body（有数据要回传），Delete 成功返回 204 + 空 body（无数据要回传）。

## Step 3.0：前端脚手架
- 日期：2026-07-20
- 目标：用 Vite 官方模板在空 `frontend/` 目录初始化 Vue 3 + TypeScript 前端。
- 执行内容：`Remove-Item -Recurse -Force .\frontend`（清空目录）→ `npm create vite@latest frontend -- --template vue-ts` → `npm install`。
- 验证：`frontend/package.json` 的 dependencies 是 `vue ^3.5.39`，devDependencies 包含 `vite ^8.1.1` / `vue-tsc ^3.3.5` / `typescript ~6.0.2`；`npm run dev` 启动 Vite 默认页 5173 端口。
- 设计依据：`docs/superpowers/specs/2026-07-20-frontend-design.md`。
- 学习点：
  - Vite 模板的 5 个核心文件：`index.html` / `src/main.ts` / `src/App.vue` / `vite.config.ts` / `tsconfig.json`。
  - `package.json` 的 `"type": "module"` 让项目成为 ESM 入口，`import` 不再需要 `.ts` 后缀。
  - 脚手架默认带 `src/components/HelloWorld.vue` 和 `src/style.css` 示例组件，本项目后续会删掉换成自己的组件。

## Step 3.1：Vite 代理 + API client + 健康检查
- 日期：2026-07-20
- 目标：让前端能通过相对路径 `/api/*` 调到后端 8181，绕开浏览器同源策略限制；用 fetch wrapper 统一 baseURL 和错误处理；先打通 `/api/health` 一条路径验证全链路。
- 执行内容：
  - `vite.config.ts` 增 `server: { proxy: { '/api': 'http://127.0.0.1:8181' } }`。
  - 新建 `src/api/types.ts`：定义 `Note` / `NoteInput` / `NotesList` 类型，字段名与后端 `store.Note` 的 JSON tag 1:1。
  - 新建 `src/api/client.ts`：薄 fetch wrapper，导出 `apiGet` / `apiPost` / `apiPut` / `apiDelete` 和 `ApiError` 类；非 2xx 响应读取后端 `{ error }` 字段抛成 `ApiError`；204 走 `undefined as T`。
  - 新建 `src/api/notes.ts`：先只导出 `getHealth()`，作为打通验证的最小 API。
  - `src/App.vue` 临时重写为 `onMounted` 调 `getHealth()` 把结果显示在页面上。
- 验证：浏览器打开 5173，页面显示「后端健康检查：**ok**」；DevTools Network 面板请求 URL 是 `localhost:5173/api/health`、Status 200（**不是** `127.0.0.1:8181`，这是 proxy 生效的证据）。
- 设计依据：`docs/superpowers/specs/2026-07-20-frontend-design.md` 的「API 对接 / 组件拆分 / 实现顺序 Step 2」章节。
- 学习点：
  - Vite dev server proxy 只对 dev 生效；生产环境前端构建后是静态文件，必须由后端同源服务（这一步后面再处理）。
  - `fetch()` 在 HTTP 4xx/5xx 时**不会 reject**，必须手动 `if (!response.ok) throw`；网络错误才会 reject。
  - 后端错误响应是 `{ "error": "..." }` JSON；wrapper 用 `response.json()` 解析这个字段抛成 `Error`，让调用方统一 `try/catch`。
  - 类型契约是前后端的合同：前端 `Note` 的字段名（`categoryId` / `createdAt` / `updatedAt`）必须与后端 Go struct 的 JSON tag 一致，否则 `tsc --noEmit` 报错或运行时字段缺失。

## Step 3.2：CRUD API 函数 + 临时列表验证
- 日期：2026-07-20
- 目标：补齐 `listNotes` / `createNote` / `updateNote` / `deleteNote` 四个 CRUD API 函数；在 App.vue 临时显示列表，验证 5 个 API（含 health）全部能联通。
- 执行内容：
  - `src/api/notes.ts` 增 4 个函数：`listNotes` / `createNote` / `updateNote` / `deleteNote`；每个都用对应的 `apiGet` / `apiPost` / `apiPut` / `apiDelete`，并以泛型显式标注响应类型。
  - `src/App.vue` 重写：`onMounted` 调 `listNotes()`，渲染 `<li v-for>` 列表 + `new Date(updatedAt).toLocaleString()` 格式化时间。
- 验证：用 curl 创建 2-3 条笔记后刷新浏览器 5173，列表正确显示；空数据库时显示"暂无笔记"；Network 面板 `GET /api/notes` 200。
- 设计依据：`docs/superpowers/specs/2026-07-20-frontend-design.md` 的「API 对接 / 实现顺序 Step 3」章节。
- 学习点：
  - TypeScript 泛型 `apiGet<NotesList>` 让 TS 知道 `result.items` 的类型，后端改返回结构时编译期立刻报错。
  - HTTP 状态码对应：200/201 拿 body；204 wrapper 已用 `undefined as T` 短路；4xx/5xx wrapper 抛 `ApiError`，组件用 `e.message` 展示用户可读错误。
  - 后端时间字段用 `time.RFC3339` 格式（如 `2026-07-20T11:30:00Z`），前端 `new Date(s).toLocaleString()` 按系统语言和时区显示，零依赖。
  - `v-if / v-else-if / v-else` 三分支处理 error / 空 / 有数据，模板里直接表达状态机。
  - 列表渲染用 `<ul>` + `<li v-for="note in notes" :key="note.id">`；`key` 用稳定 id（不是数组 index），后续做删除/排序才不会错位。

## 命令行手册：从零创建同类项目

下面的命令以 Windows PowerShell 为例。创建新项目时，把 `<project-name>`、`<owner>` 和 `<repo>` 替换成自己的值；不要把尖括号原样输入。

### 1. 创建目录和 Git 仓库

```powershell
New-Item -ItemType Directory -Force -Path <project-name>
Set-Location <project-name>
git init
New-Item -ItemType Directory -Force -Path docs, backend, frontend
```

- `New-Item -ItemType Directory` 创建目录。
- `Set-Location` 进入项目目录，后面的命令都在当前目录执行。
- `git init` 创建本地 Git 仓库。
- `docs` 保存学习文档，`backend` 和 `frontend` 分别管理 Go 和 Vue。

### 2. 初始化 Go 后端

```powershell
Set-Location backend
go mod init github.com/<owner>/<repo>/backend
New-Item -ItemType Directory -Force -Path cmd/server, internal/httpapi, internal/store
go test ./...
```

- `go mod init` 创建 `go.mod`，最后的 Module 路径应与未来远程仓库路径一致。
- `cmd/server` 放程序入口，`internal/httpapi` 放 HTTP 层，`internal/store` 放数据库访问层。
- 第一次运行 `go test ./...` 可能提示没有测试，这是正常的基线结果。

### 3. 初始化 Vue 前端

如果 `frontend` 还是空目录，可以在项目根目录执行：

```powershell
npm create vite@latest frontend -- --template vue-ts
Set-Location frontend
npm install
npm run dev
```

- `npm create vite@latest` 使用 Vite 模板生成 Vue 3 + TypeScript 项目。
- `npm install` 安装 `package.json` 中声明的依赖。
- `npm run dev` 启动前端开发服务器，默认地址通常是 `http://127.0.0.1:5173`。

### 4. 日常开发前先恢复状态

```powershell
Set-Location <project-root>
git status --short
Get-Content .\docs\project-playbook.md
Get-Content .\docs\learning-log.md
```

先看 Git 状态是为了避免覆盖未提交修改；再读两份文档是为了知道当前阶段和学习目标。

### 5. 后端测试、运行和接口验证

```powershell
Set-Location .\backend
go test ./...
go run ./cmd/server
```

服务启动后，在另一个 PowerShell 窗口验证健康检查：

```powershell
Invoke-RestMethod -Method Get -Uri http://127.0.0.1:8181/api/health
```

如果返回 `status: ok`，说明“请求 -> Go 服务 -> 路由 -> JSON 响应”的链路已跑通。测试通过和服务能启动是两个不同层次的验证，都要记录。

### 6. 前端验证

```powershell
Set-Location .\frontend
npm run build
```

`npm run build` 验证 TypeScript、Vue 模板和生产构建；开发阶段再用 `npm run dev` 进行浏览器交互验证。

### 7. 完成一个阶段后的检查

```powershell
Set-Location <project-root>
git diff --check
git status --short
```

确认没有空白错误、没有把运行时数据库或敏感信息加入变更后，再由用户决定是否执行：

```powershell
git add <files>
git commit -m "feat: <short-description>"
git push origin <branch-name>
```

提交和推送会改变外部 Git 状态，必须先由用户明确授权。每次命令都要知道“命令做什么、预期输出是什么、失败后如何定位”，不能只复制粘贴。

## Step 3.6：NoteEditView 风格统一 + 编辑/详情分页
- 日期：2026-07-20
- 目标：让 NoteEditView 跟 NoteListView/NoteDetailView 风格统一（都用 Element Plus 组件），加"编辑"按钮路由 `/notes/:id/edit`。
- 执行内容：
  - 重写 `NoteListView.vue`：注释精简到 1 行；标题跳 `/notes/:id`（详情）；加 `<el-button>编辑</el-button>` 在删除按钮前。
  - 新建 `NoteDetailView.vue`：只读展示 + 编辑/删除/返回三个按钮 + 找不到笔记走 `el-empty`。
  - 改 `router.ts`：`/notes/:id` 走 `NoteDetailView`；`/notes/:id/edit` 走 `NoteEditView`。
  - 改 `NoteEditView.vue`：`<el-form>` / `<el-form-item>` / `<el-input>` / `<el-button>` 替换原生 input/textarea；`<el-alert>` / `<el-empty>` 跟另外两个 View 对齐；`ElMessage` 提示保存成功/失败；`native-type="submit"` 让 Enter 键也能提交。
- 验证：浏览器走 5 个流程（新增/删除/编辑/不存在 id/详情）都通过；NoteEditView 视觉/交互跟 NoteListView/NoteDetailView 一致。
- 设计依据：Ray 风格约定"项目风格统一，不一半用原生一半用组件"——CRUD UI 风格一致是基本盘。
- 学习点：
  - "测试通过"只能证明功能正确，不等于满足自己的风格约定——review 时要按自己的标准对照检查。
  - 路由设计：详情页和编辑页是两个 URL，不是同一个 URL 加 mode 切换——URL 是状态的天然 bookmark。
  - Element Plus d.ts：`components.d.ts` / `auto-imports.d.ts` 由 unplugin 自动生成，**`npm run dev` 第一次跑才会生成**——`vue-tsc --noEmit` 才会通过；新加的 el-form-item 类型不在时先跑 dev 一次。

## Step 4：批量导入 .md（Phase A）
- 日期：2026-07-20
- 目标：让用户能选多个 .md 文件一次性导入为笔记（保留 Markdown 正文，自动用文件名作 title）。
- 执行内容：
  - **后端**：
    - `store/note.go` 加 `ImportNoteInput` / `ImportError` / `ImportResult` 类型（**3 个类型必须加 json tag**，否则 Phase A 教训重演——前端拿全 undefined silent fail）。
    - `Store.ImportNotes(notes)`：业务校验 title 非空 + 一个事务批量插入 + 三状态码（201/207/400）。
    - `httpapi/notes.go` 加 `importBatch` 方法：`POST /api/notes/import`，body 是 `{ "notes": [...] }`。
  - **前端**：
    - `types.ts` 加 `ImportNoteInput` / `ImportError` / `ImportResult`。
    - `notes.ts` 加 `importNotes(inputs)` API 函数。
    - `utils/markdown.ts` 加极简 front matter 解析（不引 js-yaml 库——少依赖 = 好学）。
    - `components/ImportDialog.vue` 新建：`<el-upload>` + `<el-dialog>` + `<el-table>` 统一风格。
    - `views/NoteListView.vue` 顶部加 `<el-button>导入 .md</el-button>`。
- 验证：`vue-tsc --noEmit` 0 错；`go test ./...` 全绿（含 2 个新测试 `TestStoreImportsNotes` + `TestStoreImportSkipsInvalidNotes`）；浏览器走"选 3 个 .md → 批量导入"流程，列表立即显示 3 条。
- 设计依据：`docs/superpowers/specs/2026-07-20-markdown-import-design.md`。
- 学习点（**Phase A 核心教训**）：
  - **跨边界类型必须有 json tag**——Go encoding/json 默认 PascalCase 字段名，前端 types 是 camelCase，没 tag = silent fail。
  - **TS 严格类型不验证运行时 JSON 字段名**——前端拿到全 undefined 不会编译报错，是运行时崩。
  - **defer tx.Rollback() 是 no-op** 模式：commit 成功后 rollback 不报错，事务里写 defer 在前面最稳。
  - **错误聚合 vs 立即返回**：批量接口不要第一个错就返回——收集所有 Errors 一起给前端，前端能展示完整失败明细。
  - **HTTP 207 Multi-Status** 是"部分成功"标准状态码（来自 WebDAV），不要用 200 凑合。

## Step 5：详情页 Markdown 渲染（Phase A+）
- 日期：2026-07-20
- 目标：详情页 raw markdown 太丑（`#` `##` ``` 都显示成字符），改成 GitHub README 风格的渲染输出。
- 执行内容：
  - `frontend/src/utils/markdown.ts` 加 `marked.use({...})` 配置 + `renderMarkdown(text)` 函数（GFM + GitHub 风格）。
  - `NoteDetailView.vue` 改 `<div v-html="renderMarkdown(note.content)">` + 加 GitHub README 风格 `<style>`。
  - 引入 `marked` v18 + `dompurify`（**不**用——单机信任源，省一个依赖）。
- 验证：详情页有 `# 标题` / `## 副标题` / ```` ```go ```` 代码块 / 表格 / 链接 / 引用 / 列表 都正确渲染；样式跟 GitHub README 一致。
- 学习点（**Phase A+ 核心教训**）：
  - **`<style scoped>` 不会作用 v-html 注入的 DOM**——v-html 注入的 DOM 不带 scope attribute，要么非 scoped，要么 `:deep()`。
  - **marked v18 API**：`marked.parse(text, { async: false })` 强制同步；module-level `marked.use({...})` 配置。
  - **DOMPurify 可选**——单机信任源（自己写自己读）可以不加 XSS 过滤；多用户/外部输入必须加。
  - **XSS 风险**：v-html 注入的 DOM 信任源风险——只有"自己输入自己读"才安全；公共输入必须 DOMPurify。

## Step 6：分类落地（Phase B）
- 日期：2026-07-21
- 目标：把 `data-model.md` 早就定义但一直未落地的 `categories` 表启用，让笔记可以按分类组织、列表可按分类筛选。
- 执行内容（**后端 store**）：
  - 新建 `store/category.go`：`Category` 类型（带 json tag）+ `CreateCategory(name)` + `ListCategories()`。
  - 改 `store/note.go`：
    - `Note` 类型加 `CategoryName *string` 字段。
    - `CreateNoteInput` / `UpdateNoteInput` 加 `CategoryID *int64` 字段（**必须 `*int64` 不是 `int64`**——nullable 双指针原则）。
    - `CreateNote` / `UpdateNote` SQL 加 `category_id` 列（用 `sql.NullInt64` 中转）。
    - `ListNotes` / `UpdateNote` SQL 加 `LEFT JOIN categories c ON n.category_id = c.id` + 多 `SELECT c.name AS category_name`。
    - `ListNotes` / `UpdateNote` Scan 加 `&note.CategoryName` + `var categoryName sql.NullString` + `if categoryName.Valid { ... }`。
    - 新增 `ListNotesByCategory(categoryID)`：WHERE `n.category_id = ?` 过滤。
- 执行内容（**后端 httpapi**）：
  - 新建 `httpapi/category.go`：`categoryHandler` + `create` + `list`（**跟 notesHandler 模板同构**——照葫芦画瓢）。
  - 改 `httpapi/health.go`：
    - 加 `CategoryStore` 接口（**接口分离**原则——跟 NotesStore 独立）。
    - `NewHandler(notesStore, categoryStore)` 改签名（**改签名 = 改所有调用方**，Go 编译会抓 10+ 处）。
  - 改 `httpapi/notes.go`：
    - `create` / `update` 的 JSON struct 加 `CategoryID *int64 \`json:"categoryId"\``。
    - `list` 加 `?categoryId=N` query 参数解析（用 `request.URL.Query().Get` 不是 `PathValue`）。
  - 改 `cmd/server/main.go`：`newServer(notesStore, categoryStore)` 传 2 个 store。
- 执行内容（**前端**）：
  - `api/types.ts` 加 `Category` / `CategoriesList` / `CreateCategoryInput`，`NoteInput` 加 `categoryId?: number | null`，`Note` 加 `categoryName?: string`。
  - `api/categories.ts` 新建：`listCategories()` / `createCategory()`。
  - `views/NoteEditView.vue`：加 `<el-select v-model="categoryId">` 分类下拉 + `+ 新建` 按钮（弹窗输入名字 + 立即可用）。
  - `views/NoteDetailView.vue`：标题下加一行 `<p v-if="note.categoryName">分类：{{ note.categoryName }}</p>`。
  - 改错误提示：所有 `catch` 加 `ElMessage.error(msg)` 弹窗（不只是页面顶部 el-alert）。
- 验证：`go test ./... -v` 27 个全 PASS（13 store + 12 httpapi + 1 server + 1 ...）；`vue-tsc --noEmit` 0 错；浏览器走"创建分类 → 创建带分类的 note → 详情页看分类"全流程。
- 设计依据：`docs/superpowers/specs/2026-07-21-categories-design.md`。
- 学习点（**Phase B 核心教训**）：
  - **外键约束 + ON DELETE SET NULL**：`notes.category_id REFERENCES categories(id) ON DELETE SET NULL`——删 category 自动让相关 note 变"无分类"，handler 不需要先解绑。
  - **SQL 别名 + LEFT JOIN 语义**：`FROM notes n LEFT JOIN categories c ON n.category_id = c.id`——`n` / `c` 别名避免歧义；LEFT JOIN 让"没分类的 note"也保留（INNER JOIN 会过滤掉）。
  - **`sql.NullX` 中转模式**：Go database/sql **不能**直接 Scan 到 `*string`，必须 `sql.NullString` 中转再 `if Valid` 判断。
  - **Scan 顺序敏感性**：SQL 加列 Scan 必须同步加变量（`Scan` 按列出现顺序赋值，不是列名）。
  - **改签名 = 改所有调用方**：Go 编译会一次性暴露所有 `NewHandler(x)` 缺参数（10+ 处）——包括测试文件。
  - **mock 模式隔离**：`fakeNotesStore` 不关心的接口参数可以传 `nil`（不测哪个接口传哪个 nil）。
  - **query vs path 参数**：`PathValue` vs `URL.Query().Get`——前端类比：path = `/users/:id`（路由级），query = `?page=1`（参数级）。

## Step 6.5：time.Format 精度修复（跨 Step fix）
- 日期：2026-07-21
- 目标：修复 phase A 一直潜伏、`go test ./...` 全跑时才暴露的 `TestStoreUpdatesNote` 失败。
- 根因：`time.Now().UTC().Format(time.RFC3339)` 默认精度到秒；创建和更新在同一秒内产生的字符串相等，`updated_at should change` 断言失败。
- 修复：3 处（`CreateNote` / `UpdateNote` / `ImportNotes`）改 `Format(time.RFC3339Nano)` 精度到纳秒。
- 验证：12 个测试全 PASS（含 `TestStoreUpdatesNote`）。
- 学习点：
  - **`time.Format(layout)` 的 layout 决定精度**——`RFC3339` 不带小数（秒）、`RFC3339Nano` 带 9 位小数（纳秒）。
  - **ISO 8601 字符串字典序 = 时间序**——前端按 `updatedAt` 排序逻辑完全不受精度影响。
  - **TDD 跑全量（`go test ./...`）是抓 silent fail 的唯一保险**——之前测试碰巧都在 1s 间隔跑没暴露；这次同秒才暴露。
