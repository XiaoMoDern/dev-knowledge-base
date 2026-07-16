---
id: go-short-variable-declaration
title: ":= 短变量声明"
category: Go
tags:
  - 变量
  - 类型推断
  - 基础语法
summary: "Go 中 := 用于在函数内声明新变量，并根据右侧表达式自动推断类型。"
---

# := 短变量声明

## 一句话结论

`:=` 表示“声明一个新变量，并由右侧的值推断类型”。它只能在函数内部使用。

## 最小例子

```go
name := "XiaoMoDern" // Go 自动推断 name 的类型为 string
count := 1            // Go 自动推断 count 的类型为 int
```

这两行等价于：

```go
var name string = "XiaoMoDern"
var count int = 1
```

## 项目中的实际用法

健康检查的测试里有两行：

```go
request := httptest.NewRequest(http.MethodGet, "/api/health", nil)
response := httptest.NewRecorder()
```

这里 Go 会根据函数返回值推断类型：

- `request` 的类型是 `*http.Request`，表示一次 HTTP 请求。
- `response` 的类型是 `*httptest.ResponseRecorder`，用于在测试中接收接口返回值。

后面实现路由时也会写：

```go
router := http.NewServeMux()
```

`router` 的类型由 `http.NewServeMux()` 自动推断，不需要你手写完整类型。

## 使用规则

1. 只能在函数内部使用，不能在包级变量位置使用。
2. 当前作用域中至少要有一个新变量。
3. 已经存在的变量单独重新赋值时，要用 `=`，不能用 `:=`。

```go
name := "first"
name = "second" // 正确：name 已经存在
```

## 前端类比

接近 TypeScript 的：

```ts
const name = "XiaoMoDern"
const count = 1
```

两者都会从右侧推断类型。区别是 Go 的 `:=` 是“声明变量”的语法，变量之后仍可用 `=` 重新赋值；它不是只读声明。

## 容易误解

`:=` 不等于“所有赋值都能用”。下面写法错误：

```go
name := "first"
name := "second"
```

因为第二行没有声明任何新变量。

## 我的练习

在 `main.go` 或独立练习文件中分别用 `:=` 和 `var` 声明一个字符串和整数，然后用 `fmt.Println` 打印它们。
