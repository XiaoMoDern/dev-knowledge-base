# 前端初始化设计

## 目标

为 dev-notebook 增加一个能跑在浏览器里的前端，覆盖笔记的"列表 + 查看 + 新建 + 编辑 + 删除"五条交互路径，让后端 CRUD 不再只能靠 curl / Apifox 验证。技术栈 Vue 3 + TypeScript + Vite + vue-router；通过 Vite 代理绕开跨域问题，不需要后端加 CORS 中间件。

## 范围

第一版只做单工作空间、单用户场景（与后端 `data-model.md` 第一版定义一致）。前端不实现登录、不切换工作空间、不引入 Pinia 状态管理（先用 `ref` + composable），不写测试（Vite 模板自带 TypeScript 类型检查作为基本防线；后续再讨论 vitest 引入）。

## 技术选型

| 角色 | 选择 | 理由 |
| --- | --- | --- |
| 框架 | Vue 3（Composition API + `<script setup>`） | 官方模板默认，类型推断最友好 |
| 语言 | TypeScript（strict 模式） | 后端已经是强类型语言，前端不能弱 |
| 构建 | Vite 5 | 后端 `playbook.md` 已记录脚手架命令，直接复用 |
| 路由 | vue-router 4 | 模板自带，免去额外选型 |
| 状态 | `ref` / `reactive` + composable | 状态规模还小，过早引入 Pinia 是过度设计 |
| HTTP | 原生 `fetch` + 自写薄 wrapper | 不引 axios；学习"用 fetch 也能写得舒服" |
| 跨域 | Vite devServer.proxy | 不动后端，配置集中在 `vite.config.ts` |

## 目录结构

```text
frontend/
├── index.html
├── package.json
├── tsconfig.json
├── vite.config.ts
├── src/
│   ├── main.ts              # 入口：挂载 App、注册路由
│   ├── App.vue              # 根组件：放 <RouterView />
│   ├── router.ts            # 路由表
│   ├── api/
│   │   ├── client.ts        # fetch wrapper：baseURL、错误处理、JSON 解析
│   │   ├── notes.ts         # 笔记相关 5 个 API 调用
│   │   └── types.ts         # 与后端 Note / 输入类型一一对应的 TS 类型
│   ├── views/
│   │   ├── NoteListView.vue # 列表 + 删除按钮
│   │   ├── NoteEditView.vue # 同一组件同时承载新建和编辑（用 route.params.id 区分）
│   │   └── NotFoundView.vue # 404 页
│   └── components/
│       └── NoteForm.vue     # 标题 + 内容表单，编辑/新建复用
└── public/
```

## API 对接

后端已经存在的契约（见 `backend/internal/httpapi/notes.go`）：

| 方法 | 路径 | 入参 | 出参 | 状态码 |
| --- | --- | --- | --- | --- |
| GET | `/api/health` | — | `{ status: "ok" }` | 200 |
| GET | `/api/notes` | — | `{ items: Note[] }` | 200 |
| POST | `/api/notes` | `{ title, content }` | `Note` | 201 / 400 / 500 |
| PUT | `/api/notes/{id}` | `{ title, content }` | `Note` | 200 / 400 / 404 / 500 |
| DELETE | `/api/notes/{id}` | — | — | 204 / 400 / 404 / 500 |

类型契约（前端 `src/api/types.ts`）必须与后端 `store.Note` 字段名 1:1 对齐，否则列表和编辑来回就接不上：

```ts
// 与后端 store.Note 字段保持一致
export interface Note {
  id: number;
  categoryId?: number;          // 后端 CategoryID *int64, omitempty
  title: string;
  content: string;
  visibility: 'private' | 'public';
  createdAt: string;            // ISO 8601 / RFC3339
  updatedAt: string;
}

export interface NoteInput {
  title: string;
  content: string;
}
```

## 路由

```text
/                  -> NoteListView
/notes/new         -> NoteEditView (mode = create)
/notes/:id         -> NoteEditView (mode = edit, 预填表单)
/api/health...     -> 由 Vite 代理，不进 vue-router
```

未匹配路由 → `NotFoundView`（也用于"打开一个不存在的笔记 id" 时的 404 兜底）。

## 组件拆分原则

- `NoteForm.vue` 只负责"标题 + 内容"两个字段 + 提交/取消按钮；不感知是新建还是编辑（数据由父组件传）。
- `NoteEditView` 负责"加载笔记（编辑模式）/ 提交 / 错误展示 / 路由跳转"。
- `NoteListView` 负责"拉列表 / 删除 / 跳编辑"，不做表单。
- `api/client.ts` 负责把 fetch 包装成 `apiGet<T>` / `apiPost<T>` / `apiPut<T>` / `apiDelete`，所有调用方都走它，禁止在组件里直接 `fetch('/api/...')`。

## 实现顺序（教学步骤）

按"先跑通骨架、再加功能、再打磨"的顺序，每步Ray亲自执行命令或编写关键代码：

1. **脚手架**：在 `F:\dev-knowledge-base\dev-notebook` 下执行 `npm create vite@latest frontend -- --template vue-ts`，Ray跑；删掉模板里 `HelloWorld.vue` / `style.css` 等示例文件。
2. **代理 + 跑通健康检查**：`vite.config.ts` 配 proxy，把 `/api` 转发到 `127.0.0.1:8181`；写 `api/client.ts` 和 `api/notes.ts` 的 `health()` 验证联通（`getHealth()` 调 `/api/health`，返回 `ok`）。
3. **类型 + API 五个函数**：写 `api/types.ts` + `api/notes.ts` 五个函数（`listNotes / getNote / createNote / updateNote / deleteNote`），其中 `getNote` 暂时后端没有，先在 API 层 stub 一个返回 `null` 的版本（占位；或者这一版不做单独 getNote，编辑模式直接走 list 找到 id）。
4. **路由 + 三个 View 骨架**：`router.ts` + `NoteListView.vue` / `NoteEditView.vue` / `NotFoundView.vue` 的最简版（先只显示标题，不接 API）。
5. **列表接 API**：在 `NoteListView` 调 `listNotes()`，渲染标题 + 更新时间 + 删除按钮；删除走 `deleteNote`，用 `confirm()` 二次确认。
6. **编辑/新建接 API**：`NoteEditView` 走 `createNote` / `updateNote`，提交后跳回列表；编辑模式从 `route.params.id` 拿 id。
7. **错误处理**：在 `api/client.ts` 把后端 `{ error: string }` 抛成 Error，让组件 `try/catch` 显示错误信息。
8. **手动验证**：启动后端 `go run ./cmd/server`，启动前端 `npm run dev`，浏览器走完"新建→列表→编辑→删除"全流程。

## 教学点（每个新概念第一次出现时单独教）

- **Vite 启动原理**：dev server 怎么把 `.vue` 文件实时编译、按需转 ESM；为什么 `npm run dev` 比传统 webpack 快几个数量级。
- **TypeScript 类型边界**：后端 `Note` → 前端 `Note` 必须 1:1；类型不对齐直接 `tsc --noEmit` 报错。
- **Composition API + `<script setup>`**：对比 Options API 思维；为什么 `<script setup>` 是当前推荐写法。
- **vue-router 4**：声明式 `<RouterLink>` vs 编程式 `useRouter().push`；动态参数 `route.params.id`。
- **Vite proxy 替代 CORS**：开发期跨域问题在浏览器侧无解，proxy 在 dev server 侧转发；为什么生产环境不会遇到（生产前端静态资源也由后端同源服务）。
- **fetch + Error 处理**：网络失败 / HTTP 非 2xx / 后端 `error` 字段三类错误怎么统一处理。
- **vue-router 4 + TypeScript**：路由参数 `useRoute().params.id` 默认是 `string | string[]`，必须自己收窄。

## 明确不做

- 用户登录、注册、Cookie / Token。
- 工作空间切换、成员邀请。
- 笔记分类（`categoryId` 字段保留，但 UI 不暴露）。
- 公开笔记页面（`visibility` 字段保留，全部走默认值 `private`）。
- 软删除、回收站、编辑历史。
- 全文搜索、分页、批量删除。
- Markdown 预览（`content` 暂用 `<textarea>`，引入渲染器是另一阶段的事）。
- 状态管理库（Pinia）；前端测试框架（vitest）；组件库（Element Plus / Naive UI）。
- 拖拽排序、富文本编辑器、文件附件、标签。

这些都等到对应阶段再开设计文档。

## 验证标准

完成"实现顺序"7 步后，需要达成：

1. `npm run build` 成功（TS 严格模式无报错）。
2. `go run ./cmd/server` + `npm run dev` 同时运行；浏览器 `http://127.0.0.1:5173` 能看到笔记列表（即使空也正常）。
3. 浏览器走完"新建→列表→编辑→删除"全流程，无 console error。
4. 删除 / 编辑不存在的 id 给出友好提示，不刷白屏。
