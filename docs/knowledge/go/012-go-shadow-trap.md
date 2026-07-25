---
id: go-shadow-trap
title: "Go := shadow 陷阱：if 块里的小心眼"
category: Go
tags:
  - Go
  - :=
  - shadow
  - scope
summary: "if x, y := ... 在 if 块里 shadow 外层变量，外层拿不到 if 块里的赋值；要'重新赋值'用 =。"
---

# Go := Shadow 陷阱

dev-notebook Phase C SearchNotes 实战中踩过：if 块里用 `:=` 声明的变量**只在 if 块内有效**，外层的同名变量**拿不到** if 块里的赋值。这是 Go 程序员最容易踩的"看起来对、跑就崩"的坑之一。

## 一句话结论

> `:=` 在 if / for / switch 块里会**新建作用域**。外层同名变量被 shadow（遮蔽），要"重新赋值"用 `=`，不要 `:=`。

## 错例（shadow 陷阱）

```go
var result PaginatedNotes
var err error

if total, err := store.CountNotes(ctx, opts); err != nil {
    return nil, err  // ← 这个 err 是 if 块里的 err，外层 err 没动
}
// ↑ if 块结束，total 和 err 都消失
// ↓ 外层 err 仍然是 nil（如果 CountNotes 没报错）

rows, err := store.ListNotes(ctx, opts)  // ← 这里才给外层 err 赋值
if err != nil {
    return nil, err
}
result.Items = notes
result.Total = ???  // ← 拿不到 if 块里的 total！
```

**症状**：编译能过、逻辑"看起来对"、跑起来 `result.Total = 0`——因为外层 total 变量从未被赋值。

## 对比例（正确：外层声明 + 块内赋值）

```go
var result PaginatedNotes
var total int64
var err error

// 块内 := 不安全，改为外层声明 + 块内 =
if err = store.CountNotes(ctx, opts, &total); err != nil {
    return nil, err
}

rows, err := store.ListNotes(ctx, opts)
if err != nil {
    return nil, err
}
result.Items = notes
result.Total = total  // ← 拿到了
```

**关键**：
- `err` 在 if 条件里用 `=`（重新赋值），不是 `:=`
- `total` 在 if **外面**声明，if 块里通过 `&total` 写入
- 块结束，外层 `total` 已经有值

## 为什么 Go 这么设计？

Go 故意让 if 条件里的 `:=` **新建作用域**——避免"if 块跑了一半就 panic，外层变量被半初始化的中间状态污染"。代价就是新手容易踩。

**前端类比**：跟 ES6 的 `let` 在 `{}` 块作用域内 shadow 同款——

```js
let x = 1
if (true) {
    let x = 2  // shadow
    console.log(x)  // 2
}
console.log(x)  // 1
```

JS 用 `let` 块作用域，Go 用 `:=` if 条件作用域——陷阱同源。

## 三个快速识别法

1. **看 if 条件里有没有 `:=`**——有就警惕 shadow
2. **if 块结束后还想用这个变量？**——必须外层声明 + 块内 `=`
3. **`go vet` 会报" ineffectual assignment"**——开了 `-vet` 的能自动抓到部分

## 实战：dev-notebook Phase C 的 `SearchNotes`

```go
func (s *Store) SearchNotes(ctx context.Context, opts SearchOptions) (PaginatedNotes, error) {
    var result PaginatedNotes
    var total int64
    var err error

    // 1. count 查询（外层 total + err 块内 =）
    if err = s.db.QueryRowContext(ctx,
        buildCountQuery(opts), buildCountArgs(opts)...,
    ).Scan(&total); err != nil {
        return PaginatedNotes{}, fmt.Errorf("count notes: %w", err)
    }

    // 2. list 查询（外层 rows 块内 := + err 块内 =）
    rows, err := s.db.QueryContext(ctx,
        buildListQuery(opts), buildListArgs(opts, result.PageSize)...,
    )
    if err != nil {
        return PaginatedNotes{}, fmt.Errorf("query notes: %w", err)
    }
    defer rows.Close()

    // 3. 扫描结果
    items := make([]Note, 0)
    for rows.Next() {
        // ... scan
        items = append(items, note)
    }
    if err := rows.Err(); err != nil {
        return PaginatedNotes{}, fmt.Errorf("rows err: %w", err)
    }

    result.Items = items
    result.Total = total  // ← 拿得到（外层声明的）
    return result, nil
}
```

## 常见误区

- **"我在 if 块外声明 err，if 条件里用 `:=` 重新声明一下"**——这是 shadow，外层 err 不会更新
- **"我用 `result, err := ...` 把 err 重新声明"**——err 仍是 shadow
- **"改用 `var err error` + `if err = ...` 啰嗦"**——啰嗦但正确，shadow bug 比啰嗦代价大
- **"Go 应该有更好语法"**——Go 1.22+ 的 if 块作用域不变，**`if x := ...; ...`** 故意这么设计，写法绕不开

## 调试位置

- 变量值"看起来对、跑就 0"——大概率 shadow
- `if` 块外的同名变量值不变——99% 是 `:=` 写成 `=`
- 编译器不报，但 `go vet ./...` 会报" ineffectual assignment"

## 关联知识点

- `011-sql-joins-and-nullable-scan.md` — 同款 silent fail 陷阱
- `006-go-functions-methods-and-multiple-returns.md` — 多返回值跟 err 处理
- `009-pointers.md` — `&total` 通过指针写入外层变量的另一种解法
