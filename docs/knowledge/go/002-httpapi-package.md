---
id: go-httpapi-package
title: "httpapi 包和 HTTP 处理器"
category: Go
tags:
  - package
  - HTTP
  - net/http
summary: "httpapi 是项目内部的 Go 包，专门负责把 HTTP 请求分发到接口处理函数。"
---

# httpapi 包和 HTTP 处理器

## 一句话结论

`internal/httpapi` 是后端处理 HTTP 请求的代码目录；文件第一行的 `package httpapi` 表示这些文件属于同一个 Go 包。

## 目录为什么这样分

```text
backend/
  cmd/server/        # 程序入口，负责启动 HTTP 服务
  internal/httpapi/  # 路由和接口处理器
```

前端可以理解为：

```text
src/main.ts          # 类似 cmd/server/main.go
src/api/ 或 router/  # 类似 internal/httpapi/
```

区别是 Go 的 `httpapi` 不请求后端，它本身就是后端接收请求的一层。

## 当前测试在验证什么

```go
NewHandler().ServeHTTP(response, request)
```

这句话的意思是：创建项目的总路由器，然后把一个模拟的 `GET /api/health` 请求交给它处理。测试再检查返回结果是不是 HTTP 200 和 `{"status":"ok"}`。

## 当前项目中的代码

```go
router := http.NewServeMux()
router.HandleFunc("GET /api/health", healthHandler)
```

- `http.NewServeMux()`：创建路由器。
- `router`：用 `:=` 声明的局部变量。
- `HandleFunc`：注册路径和处理函数。
- `healthHandler`：当浏览器请求 `/api/health` 时执行的函数。

处理函数当前返回固定 JSON：

```go
response.Header().Set("Content-Type", "application/json")
response.WriteHeader(http.StatusOK)
_, _ = response.Write([]byte(`{"status":"ok"}`))
```

- `response` 是 HTTP 响应对象。
- `Header().Set` 设置浏览器看到的响应类型。
- `WriteHeader(http.StatusOK)` 返回 HTTP 200，表示请求成功。
- `Write` 将 JSON 内容写入响应体。

## 为什么目录叫 internal

Go 规定：`internal` 目录下的包只能被同一项目范围内的代码导入。它相当于“应用私有实现”，避免未来其他项目依赖你还可能调整的内部代码。
