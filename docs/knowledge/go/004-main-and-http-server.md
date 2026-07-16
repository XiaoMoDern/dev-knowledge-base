---
id: go-main-and-http-server
title: "main.go、HTTP Server 与 if err :="
category: Go
tags:
  - main
  - 函数
  - 指针
  - 错误处理
  - HTTP
summary: "main.go 是 Go 程序入口；它创建 HTTP 服务并调用 ListenAndServe 开始监听端口。"
---

# main.go、HTTP Server 与 if err :=

## 一句话结论

`func main()` 是 Go 可执行程序的起点。当前项目的 `main.go` 创建一个 HTTP 服务，监听 `127.0.0.1:8181`，并把请求交给 `httpapi` 路由器处理。

## 前端类比

```text
Vue: src/main.ts -> createApp(App).mount(...)
Go:  cmd/server/main.go -> server.ListenAndServe()
```

两者都是应用的启动入口。

## 当前项目中的服务创建

```go
func newServer() *http.Server {
    return &http.Server{
        Addr:    "127.0.0.1:8181",
        Handler: httpapi.NewHandler(),
    }
}
```

- `func newServer()`：定义一个函数，函数名小写，表示只在当前包内部使用。
- `*http.Server`：函数返回一个 HTTP 服务对象的指针。可以先把指针理解为“指向这个服务对象的引用”。
- `&http.Server{ ... }`：创建 `http.Server` 结构体，并取得它的指针。
- `Addr`：服务监听地址。
- `Handler`：收到 HTTP 请求后交给哪个路由器处理。

## 启动服务

```go
func main() {
    server := newServer()

    if err := server.ListenAndServe(); err != nil {
        log.Fatal(err)
    }
}
```

`if err := ...; err != nil` 是 Go 非常常见的写法：

1. `err := server.ListenAndServe()` 先调用函数并声明 `err`。
2. `;` 后立刻判断 `err` 是否不是 `nil`。
3. 如果发生错误，进入代码块处理错误。

这相当于把“调用函数”和“检查错误”放在同一个 `if` 中，`err` 的作用域也只在这个 `if` 里。

## 为什么访问 health 接口能证明服务工作

```text
浏览器
  -> 127.0.0.1:8181
  -> main.go 创建的 http.Server
  -> httpapi.NewHandler()
  -> GET /api/health
  -> {"status":"ok"}
```

这个链路通了，说明程序入口、端口监听、路由注册和响应写入都正常。
