# 学习记录

## 2026-07-16：项目初始化

- Git 仓库用于记录每次变更，便于回退和复盘。
- `frontend` 和 `backend` 分开，是因为它们有不同的依赖、构建工具和运行方式。
- `docs` 不属于业务代码，专门沉淀项目过程、技术决策和学习笔记。

## 2026-07-16：Go 测试先行

- `go test ./...` 会查找当前 Module 下的所有测试文件并执行测试。
- `httptest.NewRequest` 和 `httptest.NewRecorder` 可以在不启动真实 HTTP 服务的情况下测试接口。
- `undefined: NewHandler` 是预期失败：测试需要的函数尚未实现，说明测试确实覆盖了目标行为。

## 2026-07-16：可导入的 Go 知识笔记

- `docs/knowledge/go/001-short-variable-declaration.md` 记录 `:=` 的规则、项目实例和前端类比。
- `docs/knowledge/go/002-httpapi-package.md` 记录 `httpapi` 包、路由器和处理器的职责。
- 这些笔记采用带 YAML 元数据的 Markdown；后续知识库网站可以将它们批量导入数据库。

## 2026-07-16：健康检查测试通过

- `http.NewServeMux()` 创建标准库路由器，负责按请求方法和路径分发请求。
- `router.HandleFunc("GET /api/health", healthHandler)` 将 GET 请求绑定到 `healthHandler` 函数。
- TDD 的 Green 阶段不是“写很多代码”，而是只写足以让已失败测试通过的代码。

## 2026-07-16：Go 名称必须完全一致

- `serve` 和 `server` 是不同变量；声明后必须使用同一个名称。
- `NewServer` 和 `newServer` 是不同函数名；Go 还用首字母大小写区分是否对外导出。
- 修正测试中的命名错误后，测试应只因为 `newServer` 尚未实现而失败，这才是有效的 TDD Red 阶段。

## 2026-07-16：启动第一个 Go HTTP 服务

- `func main()` 是 Go 程序入口，类似 Vue 项目的 `main.ts`。
- `http.Server` 保存监听地址和请求处理器；`ListenAndServe()` 让它开始接受 HTTP 请求。
- `if err := ...; err != nil` 把错误值声明和错误判断放在一起，是 Go 的常用错误处理模式。
- 浏览器成功访问 `/api/health`，证明“浏览器 -> Go 服务 -> 路由 -> JSON 响应”的链路已跑通。
