---
id: go-functions-methods-and-multiple-returns
title: "Go 函数、方法接收者与多返回值"
category: Go
tags:
  - 函数
  - 方法
  - 指针
  - error
  - HTTP
summary: "Go 用参数定义输入、返回值定义输出；前面的接收者让函数成为某个类型的方法，多返回值常用于同时返回结果和错误。"
---

# Go 函数、方法接收者与多返回值

## 一句话结论

普通函数解决独立问题；方法通过接收者属于某个类型；`(Result, error)` 表示一次返回成功结果和错误结果。

## 1. 普通函数

```go
func add(a int, b int) int {
	return a + b
}
```

- `func`：声明函数。
- `add`：函数名。
- `a int`、`b int`：参数名和参数类型。
- 最后的 `int`：返回值类型。
- `return`：把结果返回给调用方。

调用函数：

```go
result := add(2, 3)
```

这里 `2` 传给 `a`，`3` 传给 `b`，`result` 得到 `5`。

## 2. 方法和接收者

```go
type Store struct {
	db *sql.DB
}

func (store *Store) CreateNote(input CreateNoteInput) (Note, error) {
	// 可以通过 store.db 访问数据库。
	return Note{}, nil
}
```

`(store *Store)` 叫方法接收者：

- `store`：方法内部使用的变量名。
- `*Store`：接收者类型；`*` 表示它指向一个 `Store` 对象。
- 它不是普通参数，而是说明“这个方法属于 Store”。

调用方法：

```go
note, err := database.CreateNote(input)
```

这里 `database` 就是方法内部的 `store`。

如果函数不属于某个对象，也不需要对象状态，就不写接收者：

```go
func add(a int, b int) int {
	return a + b
}
```

## 3. 项目中的 HTTP 函数

```go
func healthHandler(response http.ResponseWriter, request *http.Request) {
	response.Write([]byte(`{"status":"ok"}`))
}
```

这是普通函数，不是 `Store` 的方法，因为它不需要数据库对象。

- `response`：服务器写回浏览器的响应工具。
- `request`：浏览器发来的请求。
- 这种参数形式是 `net/http` 规定的处理器格式，所以可以注册到路由：

```go
router.HandleFunc("GET /api/health", healthHandler)
```

## 4. 多返回值

```go
func divide(a int, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("不能除以 0")
	}

	return a / b, nil
}
```

`(int, error)` 不是前端解构，而是函数声明了两个返回值的类型：

1. 第一个返回值是计算结果，类型为 `int`。
2. 第二个返回值是错误信息，类型为 `error`。

调用时用两个变量接收：

```go
result, err := divide(10, 2)
if err != nil {
	// 处理错误。
}
```

成功时通常返回 `result, nil`；失败时通常返回零值和错误，例如 `0, err`。

## 前端类比

调用多返回值有一点像：

```ts
const [result, err] = divide(10, 2)
```

但 Go 不是把数组解构出来，而是语言原生支持多个返回值，并且函数声明会明确写出每个返回值类型。

方法接收者可以粗略类比 JavaScript 方法中的 `this`，但 Go 通过显式的 `(store *Store)` 写出来，不隐藏对象来源。

## 容易误解

- `*Store` 是指针类型，不是乘法；指针细节可以在理解方法之后再单独学习。
- `(store *Store)` 和 `(input CreateNoteInput)` 都出现在函数签名中，但前者是接收者，后者是普通参数。
- `error` 不是一定会有的错误；成功时它通常是 `nil`，所以调用后要判断 `err != nil`。
- `(Note, error)` 不是“把 Note 和 error 解构出来”，而是声明两个返回值。
- 函数名大写（如 `CreateNote`）表示可以被其他包调用；小写（如 `healthHandler`）只在当前包内使用。

## 我的练习

1. 写一个 `multiply(a int, b int) int`，返回两个整数的乘积。
2. 写一个 `divide(a int, b int) (int, error)`，除数为 0 时返回错误。
3. 用自己的话解释 `func (store *Store) CreateNote(input CreateNoteInput) (Note, error)` 中的接收者、参数和两个返回值。
