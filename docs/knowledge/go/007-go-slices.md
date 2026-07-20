---
id: go-slices
title: "Go 切片 []T 与 append"
category: Go
tags:
  - 切片
  - 数组
  - append
  - 数据库查询
summary: "切片是可动态增长的一组同类型元素；[]Note 表示 Note 切片，append 用于追加元素。"
---

# Go 切片 `[]T` 与 `append`

## 一句话结论

`[]Note` 表示可以保存多条 `Note` 的切片；`append(notes, note)` 把一条笔记追加到切片末尾，并需要重新赋值给 `notes`。

## 最小例子

```go
type Note struct {
	Title string
}

notes := make([]Note, 0)
notes = append(notes, Note{Title: "第一篇笔记"})
notes = append(notes, Note{Title: "第二篇笔记"})
```

此时 `notes` 中有两条笔记：

```go
notes[0] // 第一篇笔记
notes[1] // 第二篇笔记
```

`notes` 是切片变量，不是下标；`0`、`1` 才是下标。

## 为什么不是单个 `Note`

查询列表时，数据库可能返回 0 条、1 条或很多条笔记，所以返回类型要使用：

```go
[]Note
```

如果一个函数只返回一条笔记，才使用：

```go
Note
```

## 项目中的实际用法

```go
notes := make([]Note, 0)

for rows.Next() {
	var note Note
	// 把当前数据库行扫描到 note
	notes = append(notes, note)
}

return notes, nil
```

执行过程是：

1. 创建一个暂时为空的 `Note` 切片。
2. 每次循环准备一个 `note`，代表当前数据库行。
3. `append` 把当前笔记放入切片。
4. 循环结束后，`notes` 保存所有查询结果。
5. 返回整个切片，而不是只返回最后一条笔记。

## 前端类比

它接近 TypeScript：

```ts
const notes: Note[] = []
notes.push({ title: "第一篇笔记" })
```

`[]Note` 类似 `Note[]`，`append` 类似数组的 `push`。

## 容易误解

- `[]Note` 不是数组下标，`[]` 在类型前表示“切片类型”。
- `notes` 是切片变量，`notes[0]` 才是通过下标取第一个元素。
- `append` 可能返回新的切片，所以通常必须写 `notes = append(notes, note)`。
- 切片只能保存同一种元素类型；`[]Note` 不能直接追加字符串。
- 空切片和 `nil` 切片都可以追加元素，但在 JSON 序列化时可能表现不同；接口返回列表时通常初始化为空切片，避免返回 `null`。

## 我的练习

创建一个 `[]string` 切片，追加三个学习主题，然后分别打印切片长度和第二个元素：

```go
topics := make([]string, 0)
topics = append(topics, "函数")
topics = append(topics, "结构体")
topics = append(topics, "切片")
```

思考：为什么第二个元素的下标是 `1`，不是 `2`？
