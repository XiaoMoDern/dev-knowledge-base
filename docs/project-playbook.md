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
