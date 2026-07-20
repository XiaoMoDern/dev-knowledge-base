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
