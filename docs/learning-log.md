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

## 2026-07-16：SQLite 自动迁移

- `database/sql` 提供统一数据库接口，空白导入 `_ "modernc.org/sqlite"` 用于注册 SQLite 驱动。
- `Open()` 不只打开文件，还应执行迁移，保证数据库结构可用。
- `CREATE TABLE IF NOT EXISTS` 让迁移可重复执行；事务让所有建表操作要么全部成功，要么全部撤销。
- `sqlite_master` 是 SQLite 的结构目录，测试可查询它来确认表是否真实存在。

## 2026-07-16：项目创建命令和协作方式

- `git init` 创建本地版本库，`go mod init` 创建 Go Module；它们分别管理项目版本和 Go 依赖边界。
- `New-Item -ItemType Directory` 在 PowerShell 中创建项目目录，`Set-Location` 切换当前工作目录。
- `go test ./...` 检查当前 Go Module 下的所有包；`npm run build` 检查 Vue 前端的生产构建。
- 新项目先看 `AGENTS.md`、`project-playbook.md` 和 Git 状态，再开始下一步，避免重复背景或覆盖未提交修改。
- Codex 负责技术主导和讲解，学习者亲自执行关键命令、编写关键代码并反馈结果。

## 2026-07-16：Go 函数、方法和多返回值

- `func add(a int, b int) int` 是普通函数；参数写在函数名后，最后的 `int` 是返回值类型。
- `(store *Store)` 是方法接收者，说明方法属于 `Store`；`(input CreateNoteInput)` 才是普通参数。
- `(Note, error)` 表示函数同时返回业务结果和错误结果，不是前端数组解构。
- HTTP 处理函数不需要 `Store` 接收者时，可以直接写成 `func healthHandler(response http.ResponseWriter, request *http.Request)`。

## 2026-07-16：Go 切片和 append

- `[]Note` 表示可以保存多条 `Note` 的切片，适合返回数据库列表结果。
- `notes` 是切片变量，`notes[0]` 才是第一个元素；下标从 `0` 开始。
- `append(notes, note)` 把元素追加到末尾，并通常写成 `notes = append(notes, note)`。
- 查询多行数据时，循环每次生成一条 `Note`，再追加到切片，最后返回完整列表。

## 2026-07-16：删除笔记的测试先行设计

- 删除接口使用 `DELETE /api/notes/{id}`，成功返回 `204`，不存在返回 `404`。
- DELETE 请求会经过路由、handler、store、SQLite，再由 handler 转成 HTTP 响应。
- `sql.Result.RowsAffected()` 可以判断 SQL 是否真的删除了记录。
- 本次按“先存储层测试、再存储实现、再 HTTP 测试、最后手动接口验证”的顺序学习。

## 2026-07-17：Go 指针正式入门

- Go 里所有传参都是值传递（复制一份），想让函数改到原值就传地址（指针）进去。
- 三个符号别混：`*T` 是类型（指针类型）、`&x` 是操作（取地址）、`*p` 是操作（解引用）。
- JS 的对象默认引用传递，Go 全部是值传递——这是前后端最大的思维差异。
- `*Store` 用指针接收者的两个原因：方法要修改内部状态、避免每次调用复制整个结构体。
- 前端类比：JS 里 `function f(o){o.x=1}` 能改到原对象；Go 默认不能，要 `*Obj` 才行。

## 2026-07-17：编辑笔记完整流程

- 同样走 TDD：先写失败测试 → 实现 → 补接口/fake → HTTP 测试 → 手动接口验证。
- 存储层 `UPDATE` 后再 `SELECT` 一次拿完整 Note，因为 `UPDATE` 只返回受影响行数，不返回行内容。
- `RowsAffected == 0` 表示"找不到"，是判断存在性的标准做法（和 Delete 同套路）。
- 状态码对比：Update 成功 200 + body（要回传新数据），Delete 成功 204 + 空 body（没数据要回）。
- HTTP 处理器四分支：id 非法 400、title 空 400、找不到 404、其它错误 500。

## 2026-07-17：跨 workspace 数据隔离模式

- 所有按 id 查工作空间资源的 SQL 都必须 `WHERE id = ? AND workspace_id = ?`，不能只按 id 查。
- 这不是性能优化，是安全边界——别的工作空间的 id 就算猜中也不该返回数据。
- 修改 SQL 加条件要同步检查参数列表：占位符 `?` 数量必须严格等于参数数量。
- UPDATE 用了 workspace 限定、SELECT 没限定 → 即使 UPDATE 失败，SELECT 也会查到跨 workspace 数据，是真实漏洞。

## 2026-07-17：调试习惯——看服务端日志

- 500 错误优先看 `go run` 终端的服务端日志（带 err 链的输出），不看 Apifox 的 HTTP 响应。
- HTTP 响应是给用户的友好提示（故意隐藏细节），日志是给开发者的根因。
- 改完代码两步走：Ctrl+S 保存 → Ctrl+C 停服务 → `go run ./cmd/server` 重启。
- Go 编译器报错读法：包名 → `file:line:col` → 主句（类型不匹配）→ `have ... want ...` 对比签名。

## 2026-07-17：Go"克制抽象"哲学

- 谚语："A little copying is better than a little dependency"——少量复制胜过少量依赖。
- "Rule of Three"——重复 3 次再考虑抽象，重复 2 次先忍着。
- create 和 update 共享 6-8 行 body 解析代码，第一版**不抽**比抽更好：
  - helper 签名/返回值会变复杂（要返回 title、content、bool、error），可读性下降；
  - 未来 update 可能变 PATCH（title 可选），helper 反而要重写。
- 前端抽组件/工具函数的成本收益和 Go 后端不同：前端 UI 复用多，后端逻辑各异，过早抽象会束缚扩展。

## 2026-07-20：Vite 开发服务器代理
- 开发期前端 5173、后端 8181 是不同端口，浏览器同源策略直接拦截跨端口请求。
- 解决方案：Vite 的 `server.proxy` 配置把匹配前缀的请求在 dev server 侧转发给后端；浏览器看到的 URL 还是同源的 `/api/...`。
- 关键限制：proxy **只对 dev 生效**。生产环境前端构建后是静态文件，必须由后端同源服务（后续阶段处理）。
- 配置位置：`vite.config.ts` 的 `defineConfig({ server: { proxy: { '/api': 'http://127.0.0.1:8181' } } })`。
- 验证方式：DevTools Network 面板看请求 URL 是不是 `localhost:5173/api/...`（不是 8181）——这就是 proxy 生效的证据。
- 前端类比：Nginx 的 `location /api/ { proxy_pass http://backend; }`——开发期就是把 Nginx 行为搬到 dev server 里。

## 2026-07-20：fetch wrapper 模式
- 原生 `fetch()` 在 HTTP 4xx/5xx 时**不会 reject**，只有网络错误才 reject；必须手动 `if (!response.ok) throw`。
- 后端错误响应统一用 `{ "error": "..." }` JSON；wrapper 用 `response.json()` 解析这个字段抛成 `Error`，调用方统一 `try/catch`。
- DELETE 成功返回 204 No Content，**没有 body**；`response.json()` 会抛错，必须提前 `return undefined as T` 短路掉。
- 模式：薄 wrapper 导出 `apiGet<T>` / `apiPost<T>` / `apiPut<T>` / `apiDelete`；组件 / composable 不直接 `fetch()`，避免到处复制 baseURL 和错误处理。
- 前端类比：axios 的 `axios.create({ baseURL })` + 拦截器——但更轻，零依赖，TypeScript 泛型直接表达响应类型。
- 自定义 `ApiError extends Error` 带 `status` 字段：调用方可以 `if (e instanceof ApiError && e.status === 404)` 做精细处理。

## 2026-07-20：前后端类型契约
- 前端 `Note` 字段（`id` / `categoryId` / `title` / `content` / `visibility` / `createdAt` / `updatedAt`）必须与后端 Go struct 的 JSON tag 一一对应。
- 后端 Go 用 `json:"categoryId,omitempty"` → 前端用 `categoryId?: number`（可选 + 数字）；后端 `*int64` 是因为 NULL 也要支持。
- 时间字段后端用 `time.RFC3339` 格式字符串 → 前端用 `string`；渲染时按需 `new Date(note.createdAt).toLocaleString()` 转换。
- TS strict 模式会强制你处理所有可选字段；漏一个 `?.` 链就会编译失败，这是前后端契约错误最便宜的发现方式。

## 2026-07-20：TypeScript 泛型与 HTTP 状态码
- `apiGet<T>(path)` 的 `<T>` 告诉 wrapper 返回类型 T；调用方拿到的 `result.items` 才有类型提示和检查。
- 后端改返回结构时，TS 编译会立刻报错，比运行时崩早一步发现。
- HTTP 状态码与前端的对应：200/201 拿 JSON body；204 DELETE 成功无 body（wrapper 已用 `undefined as T` 短路）；4xx/5xx wrapper 抛 `ApiError`，`e.message` 是后端 `error` 字段。
- 组件可以用 `e instanceof ApiError` + `e.status` 做精细处理，比如 404 走"未找到"分支。
- 后端时间字段是 RFC3339 字符串（如 `2026-07-20T11:30:00Z`），前端用 `new Date(s).toLocaleString()` 转本地化显示，零依赖。

