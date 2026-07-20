---
id: vue-fetch-wrapper
title: "fetch wrapper 模式：统一 baseURL、JSON 解析和错误处理"
category: Vue
tags:
  - fetch
  - API
  - 错误处理
  - TypeScript
summary: "薄 fetch 包装层统一 baseURL + JSON 解析 + 错误处理；组件和 composable 不直接 fetch，全部走 apiGet/apiPost/apiPut/apiDelete。"
---

# fetch wrapper 模式：统一 baseURL、JSON 解析和错误处理

## 一句话结论

> 用一个薄 fetch 包装层统一 baseURL、JSON 解析和 HTTP 错误处理；组件 / composable 全部走 `apiGet<T> / apiPost<T> / apiPut<T> / apiDelete`，**禁止直接 `fetch()`**。

## 为什么需要 wrapper

原生 `fetch` 有三个常见坑：

1. **HTTP 4xx/5xx 不会 reject**——只有网络错误（DNS 失败、连接断开）才会。`try { fetch(...) } catch` 抓不到 404/500。
2. **每个调用都要写 `baseURL`**——一旦后端换端口要全局改。
3. **错误信息分散**——后端 `{ "error": "..." }`、网络错误、JSON 解析错误都各走各的，组件要写一堆分支。

wrapper 把这三件事集中到一处。

## 最小例子

```ts
// src/api/client.ts
const baseURL = ''  // 同源走 Vite proxy；生产再换绝对地址

export class ApiError extends Error {
  status: number
  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.status = status
  }
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${baseURL}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  })

  // 204 No Content：DELETE 成功，无 body
  if (response.status === 204) {
    return undefined as T
  }

  if (!response.ok) {
    let message = `请求失败 (${response.status})`
    try {
      const body = await response.json() as { error?: string }
      if (body?.error) message = body.error
    } catch {
      // 后端可能没返 JSON，保留默认 message
    }
    throw new ApiError(message, response.status)
  }

  return response.json() as Promise<T>
}

export const apiGet = <T>(path: string) => request<T>(path, { method: 'GET' })
export const apiPost = <T>(path: string, body: unknown) =>
  request<T>(path, { method: 'POST', body: JSON.stringify(body) })
export const apiPut = <T>(path: string, body: unknown) =>
  request<T>(path, { method: 'PUT', body: JSON.stringify(body) })
export const apiDelete = (path: string) =>
  request<void>(path, { method: 'DELETE' })
```

## 三个关键点

### 1. HTTP 4xx/5xx 必须手动 throw

```ts
if (!response.ok) {  // !ok = 4xx 或 5xx
  throw new ApiError(message, response.status)
}
```

`response.ok` 是 `status >= 200 && status < 300` 的简写。**网络错**才会让 `await fetch(...)` reject，HTTP 错必须自己处理。

### 2. 204 No Content 短路掉

DELETE 成功返回 204，**没有 body**。`response.json()` 会抛 "Unexpected end of JSON input"。必须提前 `return undefined as T`：

```ts
if (response.status === 204) {
  return undefined as T  // 告诉调用方：成功但没数据
}
```

否则所有 DELETE 调用都会进 catch。

### 3. 后端 `{ error }` 抛成 Error

后端 handler 写错误响应时是：

```go
writeJSON(response, http.StatusBadRequest, map[string]string{"error": "title 不能为空"})
```

wrapper 解析这个字段抛成 `ApiError("title 不能为空", 400)`。调用方 `try/catch` 拿到的 `e.message` 直接是用户可读的错误信息。

## 前端类比

axios 的 `axios.create({ baseURL })` + 拦截器：

```ts
// axios 版（更重，但功能多）
const api = axios.create({ baseURL: '...' })
api.interceptors.response.use(
  response => response.data,
  error => Promise.reject(new ApiError(error.response?.data?.error, error.response?.status))
)
```

fetch wrapper 是 axios 的轻量替代：零依赖，TypeScript 泛型直接表达响应类型，控制权完全在你手里。

## 项目里的真实例子

```ts
// frontend/src/api/notes.ts
import { apiGet } from './client'

export function getHealth(): Promise<{ status: string }> {
  return apiGet<{ status: string }>('/api/health')
}
```

```vue
<!-- App.vue -->
<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { getHealth } from './api/notes'

const status = ref<string>('checking...')
const error = ref<string>('')

onMounted(async () => {
  try {
    const result = await getHealth()
    status.value = result.status
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  }
})
</script>
```

组件只看到 `try/catch`，不知道 fetch 也不关心 HTTP 状态码细节。

## 常见误区

- **"fetch reject 我就 try/catch 了"**——不够。HTTP 4xx 不会 reject，必须 `if (!response.ok) throw`。
- **"DELETE 也能 `response.json()`"**——不能。204 没 body，`.json()` 会抛。
- **"后端返什么错我都 alert"**——精细点：可以 `if (e instanceof ApiError && e.status === 404) ...` 做不同处理。
- **"baseURL 写绝对地址方便跨域"**——dev 期不需要，proxy 已经把跨域解决了；写绝对地址反而会在生产部署时需要全局改。

## 调试位置

- 抓不到错：检查 wrapper 有没有 `if (!response.ok) throw`；测试 500 错误时浏览器 Network 面板和终端日志都要看。
- DELETE 报 `Unexpected end of JSON input`：漏了 204 短路。
- 错误信息是英文默认模板：后端返的不是 `{ error }` 格式（看看 handler 是不是写错了 key）。
