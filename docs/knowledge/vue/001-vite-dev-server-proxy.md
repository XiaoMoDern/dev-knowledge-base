---
id: vue-vite-dev-server-proxy
title: "Vite 开发服务器代理：开发期跨域解决方案"
category: Vue
tags:
  - Vite
  - 跨域
  - 开发服务器
  - 代理
summary: "Vite dev server 通过 server.proxy 把同源请求转发给后端，绕开浏览器同源策略；这是 dev-only 的临时方案，生产环境必须同源部署。"
---

# Vite 开发服务器代理：开发期跨域解决方案

## 一句话结论

> Vite 在开发期通过 `server.proxy` 把匹配前缀的请求在 dev server 侧转发给真实后端；浏览器看到的还是同源 URL，跨域限制不触发。**这只在 dev 生效**。

## 为什么需要代理

dev 期前端 (`http://127.0.0.1:5173`) 和后端 (`http://127.0.0.1:8181`) 端口不同，浏览器同源策略会直接拦截跨端口请求。三种常见解法：

| 方案 | 适用 | 缺点 |
| --- | --- | --- |
| Vite devServer proxy | dev 期 | 只在 dev 生效 |
| 后端 CORS 中间件 | 跨域 API 服务 | 需要后端代码改动 + 浏览器要正确处理 preflight |
| 同源部署 | 生产环境 | dev 期无法用 |

dev 期最优解是 Vite proxy：不动后端，配置集中在 `vite.config.ts`。

## 配置 shape

```ts
// vite.config.ts
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '/api': 'http://127.0.0.1:8181',
    },
  },
})
```

含义：浏览器请求 `/api/notes` 时，Vite dev server 收到后**代为转发**到 `http://127.0.0.1:8181/api/notes`，再把后端响应原样返回浏览器。浏览器看到的是同源（5173）的响应，没有跨域问题。

## 前端类比

Nginx 的反向代理配置：

```nginx
location /api/ {
    proxy_pass http://backend:8181/;
}
```

Vite devServer proxy 本质就是把 Nginx 的行为搬到了 dev server 里。生产环境用 Nginx / 后端同源服务时，proxy 失效，需要用别的方式提供 `/api`。

## 项目里的真实例子

```ts
// dev-notebook/frontend/vite.config.ts
export default defineConfig({
  plugins: [vue()],
  server: {
    proxy: {
      '/api': 'http://127.0.0.1:8181',
    },
  },
})
```

前端代码直接用相对路径：

```ts
// frontend/src/api/notes.ts
export function getHealth() {
  return apiGet<{ status: string }>('/api/health')  // 实际被 proxy 转到 8181
}
```

## 验证 proxy 生效

最直接的证据是 **DevTools Network 面板**：

| 字段 | 期望值 |
| --- | --- |
| Request URL | `http://localhost:5173/api/health` |
| Status | `200` |
| Remote Address | `127.0.0.1:5173`（不是 8181） |

如果看到 `Request URL` 直接是 8181 或状态码里有 CORS 报错，说明 proxy 没配好。

## 常见误区

- **"proxy 是浏览器侧的"**——错。proxy 在 Vite dev server 侧；浏览器以为请求同源，根本没发 CORS preflight。
- **"proxy 在生产也生效"**——错。`npm run build` 产出的纯静态文件没有 dev server，proxy 配置被丢弃。生产环境必须由后端同源提供 `/api/*`（要么后端加 CORS 中间件，要么后端直接 serve 前端构建产物）。
- **"proxy 路径会改写"**——看写法。`'/api': 'http://backend'` 路径不变；`'/api': { target: 'http://backend', pathRewrite: { '^/api': '' } }` 会去掉 `/api` 前缀。本项目保留 `/api` 前缀，后端路由就不用改。
- **"改 vite.config.ts 不用重启"**——错。Vite 配置改动需要重启 dev server（Ctrl+C 后 `npm run dev`）。

## 调试位置

- 浏览器 CORS 报错：检查 `vite.config.ts` 的 `server.proxy`；确认后端端口对得上。
- 404 但后端 curl 正常：proxy 路径前缀不对；看 Network 面板 `Request URL` 和后端日志的请求路径。
- 502 Bad Gateway：后端没启动，或者 proxy target 端口错。

## 后续衔接

dev-notebook 第一版前端只做 dev 验证；等 CRUD UI 完整闭环后，会在新的设计文档里规划**生产部署**（Vite 构建产物由 Go 同源 serve，或加 Nginx 反代）。届时 proxy 配置会被自然忽略，不再是"必须"。
