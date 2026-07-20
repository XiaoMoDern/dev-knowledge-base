// fetch wrapper：统一 baseURL + JSON 解析 + 错误处理
// 调用方都用 apiGet/apiPost/apiPut/apiDelete，不要在组件里直接 fetch
const baseURL = ''  // 同源走 Vite proxy；生产再换成绝对地址

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

  // 4xx/5xx：fetch 不会 reject，必须手动抛
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

export function apiGet<T>(path: string): Promise<T> {
  return request<T>(path, { method: 'GET' })
}

export function apiPost<T>(path: string, body: unknown): Promise<T> {
  return request<T>(path, { method: 'POST', body: JSON.stringify(body) })
}

export function apiPut<T>(path: string, body: unknown): Promise<T> {
  return request<T>(path, { method: 'PUT', body: JSON.stringify(body) })
}

export function apiDelete(path: string): Promise<void> {
  return request<void>(path, { method: 'DELETE' })
}
