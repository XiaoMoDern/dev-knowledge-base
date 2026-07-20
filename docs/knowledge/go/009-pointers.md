---
id: go-pointers
title: "Go 指针：值传递、地址与接收者"
category: Go
tags:
  - Go
  - 指针
  - 接收者
  - 依赖注入
summary: "Go 里所有传参都是值传递；想改到原值要传地址（指针）；*Store 是最常见的指针接收者用法。"
---

# Go 指针：值传递、地址与接收者

## 一句话结论

> Go 里**所有传参都是值传递**（复制一份）。想让函数改到原值，就传**地址**（指针）进去，而不是传值。

## 前端类比（关键差异）

在 JavaScript 里你写 `function f(o){o.x=1}` 会改到原对象，因为 JS 的对象是**引用类型**。

Go **没有引用类型**这回事。全部是值传递。同样的代码在 Go 里**改不到原值**——除非用指针：

```javascript
// JavaScript：默认引用传递
function rename(note) { note.title = "新标题" }  // 改到了
```

```go
// Go：默认值传递，必须传指针
func rename(note *Note) { note.Title = "新标题" }  // *Note 是指针
n := Note{Title: "旧"}
rename(&n)  // &n 取 n 的地址传进去
```

## 三个符号别搞混

| 写法 | 是什么 | 例子 |
| --- | --- | --- |
| `*T` | **类型**：T 的指针类型 | `*Counter`、`*Store` |
| `&x` | **操作**：取 x 的地址 | `&c`、`&note` |
| `*p` | **操作**：取出 p 指向的值（解引用） | `*c` |

调用时 Go 会自动取地址 / 解引用，所以你写 `store.db` 看起来像普通字段访问，底层其实是 `(*store).db`——这是**语法糖**，为你省掉显式解引用的写法。

## 为什么 `*Store` 用指针接收者

项目里所有 store 方法都是 `func (store *Store) CreateNote(...)`、`(store *Store) UpdateNote(...)` 等等，那个 `*` 不是装饰，是有实际用途的：

1. **方法要访问 store 内部状态**——比如 `store.db`（数据库连接）。指针保证多个方法调用共享同一个 db，而不是各自拿到一份。
2. **避免每次调用复制整个结构体**——`Store` 结构体里如果将来加配置、连接池等字段，值传递会每次都复制一份。
3. **意图清晰**——Go 社区惯例："会修改接收者 / 或结构体将来可能变大" 就用指针。

## 项目里的真实例子

```go
// 这是项目里 store/note.go 的方法签名
func (store *Store) CreateNote(input CreateNoteInput) (Note, error) {
    workspaceID, err := store.defaultWorkspaceID()  // store.db 等内部字段
    // ...
}

func (store *Store) UpdateNote(noteID int64, input UpdateNoteInput) (Note, error) {
    // 通过 store 访问 SQLite 连接
    result, err := store.db.Exec(`UPDATE notes ...`, ...)
    // ...
}
```

`store.db`、`store.defaultWorkspaceID()` 都要靠 `*Store` 才能正确工作。

## 常见误区

- **"用了 `*Store` 所以改了 store"**——错。`*Store` 本身不"修改"任何东西，它只是保证多个调用共享同一个 Store 实例。要真改字段还得方法里有 `store.field = ...`。
- **"指针一定比值快"**——不一定。对于小结构体，复制一份比追踪指针更快（缓存友好）。Go 的经验法则：**小且不可变用值，大或要变用指针**。
- **"指针和值是同一个东西"**——错。`*Store` 和 `Store` 是**不同类型**。接口 `NotesStore` 要求 `*Store` 实现，那 `Store`（值）就不行。

## 练习

`backend/lesson/pointer.go` 里有 `Counter` 结构体和 `runPointerLesson()` 函数：

1. 运行 `go run ./lesson`，观察值传递（`Count = 0`）和指针传递（`Count = 1`）的差异。
2. 在 `Counter` 上补一个 `Double()` 方法，让 `Count` 翻倍。**接收者必须用 `*Counter`**，否则改的是副本。
3. 取消文件底部三行注释运行验证，期望 `Count = 2`。

```go
// 你的练习答案（写在 lesson/pointer.go 里）
func (c *Counter) Double() {
    c.Count = c.Count * 2
}
```

## 学习顺序

1. 先理解"Go 全是值传递"——这是和其它带引用类型语言最大的差异。
2. 区分三个符号（`*T` / `&x` / `*p`）——它们是不同的东西。
3. 看项目里 `*Store` 怎么用——理解指针接收者的实际意义。
4. 做 `lesson/pointer.go` 练习——亲手验证值传递和指针传递的差异。
5. 以后写新方法时，自己判断该用值接收者还是指针接收者。

## 调试位置

- 方法调用失败时检查接收者类型：`func (c Counter)` vs `func (c *Counter)` 决定能不能改 c。
- 接口未实现错误时（如 `*Store does not implement NotesStore`），看方法签名和接口签名的 `have ... want ...` 对比。
- `panic: nil pointer dereference` 说明你给指针变量赋了 nil 就解引用了——检查变量初始化。
