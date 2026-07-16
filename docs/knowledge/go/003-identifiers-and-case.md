---
id: go-identifiers-and-case
title: "Go 标识符、大小写与变量名一致性"
category: Go
tags:
  - 变量
  - 函数
  - 命名
summary: "Go 区分大小写；变量和函数必须以完全相同的名称被声明和使用。首字母大写还会影响是否能被外部包导入。"
---

# Go 标识符、大小写与变量名一致性

## 一句话结论

Go 中 `serve` 和 `server` 是两个不同的变量，`NewServer` 和 `newServer` 也是两个不同的函数名。

## 项目中的例子

下面写法会报错：

```go
serve := NewServer()

if server.Addr != "127.0.0.1:8181" {
    // server 没有被声明；声明的是 serve
}
```

应该统一为：

```go
server := newServer()

if server.Addr != "127.0.0.1:8181" {
    // server 与声明处是同一个变量
}
```

## 大小写的额外含义

- `newServer`：首字母小写，只能在同一个 Go 包内使用。
- `NewServer`：首字母大写，可以被其他 Go 包导入和调用。

当前的 `newServer` 只服务 `cmd/server` 内的启动逻辑，不需要让别的包调用，所以使用小写。

## 前端类比

TypeScript 中的 `createServer` 和 `CreateServer` 同样是不同标识符。Go 的额外规则是：首字母大写代表“对其他包公开”。
