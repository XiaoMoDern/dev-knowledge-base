---
id: vue-typescript-generics-and-http-status
title: "TypeScript 泛型与 HTTP 状态码在前端的实践"
category: Vue
tags:
  - TypeScript
  - 泛型
  - HTTP
  - fetch
  - 错误处理
summary: "用泛型 apiGet<T> 表达 API 响应类型；把 HTTP 4xx/5xx 包装成 ApiError，让组件用 try/catch 拿到用户可读错误。"
---

# TypeScript 泛型与 HTTP 状态码在前端的实践

## 一句话结论

> 用 `apiGet<T>` 泛型表达响应类型、用 `ApiError.status` 区分 4xx/5xx；组件只 `try/catch` 不关心 HTTP 细节，是前后端契约的"前端那一半"。

## TypeScript 泛型：把契约写进类型

```ts
// wrapper 的泛型签名
export function apiGet<T>(path: string): Promise<T>

// 调用方：明确告诉 TS 返回值类型
const result = await apiGet<NotesList>('/api/notes')
result.items  // <- TS 知道这是 Note[]，有提示和检查
```

`apiGet<NotesList>` 的 `<NotesList>` 是显式提供类型实参；TS 把它"塞"进 wrapper 里的 `Promise<T>`，于是返回值自动是 `Promise<NotesList>`。

### 泛型 vs any

| 写法 | 后果 |
| --- | --- |
| `apiGet<NotesList>('/api/notes')` | `result.items` 有类型检查 |
| `apiGet('/api/notes') as NotesList` | 类型断言，TS 信你但运行时崩了才知道 |
| `apiGet<any>('/api/notes')` | 放弃类型保护，所有字段访问都过编译 |

**优先用泛型**，只在第三方库类型不友好时退到 `as` 或 `unknown`。

### 后端改字段时的连锁反应

如果后端把 `Note.title` 改名 `Note.heading`，前端 `Note` 接口不动：

```ts
note.title  // 编译通过，但运行时 undefined
```

保持接口同步更新是约定。`tsc --noEmit` 不会自动发现"后端字段名变化"，但可以加运行时校验（zod / 自写 guard）补这一环——属于另一阶段的话题。

## HTTP 状态码与前端的对应

dev-notebook 后端用 5 个状态码：

| 状态 | 含义 | 前端 wrapper 处理 | 组件处理 |
| --- | --- | --- | --- |
| 200 OK | 成功 + body | `response.json()` | 用数据 |
| 201 Created | 创建成功 + body | `response.json()` | 用数据（POST 默认） |
| 204 No Content | 成功 + 无 body | `return undefined as T` 短路 | 不需要 body |
| 400 Bad Request | 入参错 | 读 `{ error }` 抛 `ApiError(400)` | 展示错误 |
| 404 Not Found | 资源不存在 | 读 `{ error }` 抛 `ApiError(404)` | 走"未找到"分支 |
| 500 Internal Server Error | 服务端错 | 读 `{ error }` 抛 `ApiError(500)` | 提示"服务暂不可用" |

### ApiError 设计

```ts
export class ApiError extends Error {
  status: number
  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}
```

- `name = 'ApiError'` 让 `e.name === 'ApiError'` 可以识别
- `status` 字段保留 HTTP 状态码，组件可以做精细分支
- `message` 是后端 `error` 字段，已经是用户可读字符串

组件用法：

```ts
try {
  await deleteNote(id)
} catch (e) {
  if (e instanceof ApiError && e.status === 404) {
    // 笔记已经不在了，刷新列表
    refresh()
  } else {
    error.value = e instanceof Error ? e.message : String(e)
  }
}
```

## 时间字段：RFC3339 → 本地化

后端 Go 用 `time.RFC3339` 格式：

```text
2026-07-20T11:30:00Z
2026-07-20T19:30:00+08:00
```

前端 `string` 接收，渲染时转 `Date`：

```ts
new Date(note.updatedAt).toLocaleString()
// "2026/7/20 19:30:00"（按用户系统语言 + 时区）
```

### 三个常见坑

- **时区错乱**——后端 `Z` 是 UTC，浏览器 `new Date()` 会按本地时区转；显示给中国用户就成 19:30 而不是 11:30，正确。
- **格式不统一**——别用 `note.updatedAt.slice(0, 10)` 截字符串"日期部分"，跨时区会少一天；统一用 `new Date()`。
- **排序问题**——列表按 `updatedAt DESC` 排序已经在后端 `ORDER BY` 做了，前端不要再排。

## 项目里的真实例子

```ts
// src/api/notes.ts
import { apiGet, apiPost, apiPut, apiDelete } from './client'
import type { Note, NoteInput, NotesList } from './types'

export function listNotes(): Promise<NotesList> {
  return apiGet<NotesList>('/api/notes')   // 泛型告诉 TS 返回是 NotesList
}

export function deleteNote(id: number): Promise<void> {
  return apiDelete(`/api/notes/${id}`)    // 204，无 body
}
```

```vue
<!-- App.vue -->
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { listNotes } from './api/notes'
import type { Note } from './api/types'

const notes = ref<Note[]>([])
const error = ref<string>('')

onMounted(async () => {
  try {
    const result = await listNotes()
    notes.value = result.items               // 类型安全
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)  // 友好提示
  }
})
</script>
```

## 常见误区

- **"我用 `as any` 一劳永逸"**——TypeScript 的价值就是"在编译期发现错"，放弃等于自废武功。
- **"HTTP 5xx 也走 alert 就行"**——精细化点：404 是用户操作错（提示重试），500 是服务端问题（提示稍后再试），message 也不同。
- **"前端再排一次序保险"**——重复劳动；后端 `ORDER BY` 已经做了，前端按接口顺序展示即可，除非有特殊需求。
- **"后端返什么时间格式都行，前端转换"**——约定 RFC3339 字符串最简单（`new Date()` 就能解析）；返时间戳 / 自定义字符串都要写解析代码。

## 调试位置

- TS 编译错 `Property 'items' does not exist on type ...`——泛型 `<NotesList>` 没加，或后端返回结构变了。
- DELETE 报 `Unexpected end of JSON input`——204 没短路；检查 wrapper。
- 时间显示 `Invalid Date`——后端字段不是合法 RFC3339；查 store 里 `time.RFC3339` 是不是被改了。
- 错误信息是英文默认 `请求失败 (500)`——后端没返 `{ error }` 字段；查 handler 是不是写错了。
