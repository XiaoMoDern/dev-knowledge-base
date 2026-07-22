---
id: go-if-valid-pattern
title: "Go if Valid 模式：sql.NullX 转 *T 的标准转换"
category: Go
tags:
  - Go
  - database/sql
  - nullable
  - 中转类型
summary: "Scan 到 sql.NullInt64/NullString 后，用 if Valid 决定是否赋值给 *int64/*string——这是 Go 处理 DB NULL 的标准模式。"
---

# Go if Valid 模式：sql.NullX 转 *T 的标准转换

## 一句话结论

> `sql.NullInt64` / `sql.NullString` 是 Scan 的"中转类型"（带 `Valid bool` 标记是否 NULL）。用 `if Valid { 字段 = &NullX.Int64 }` 决定是否把值搬到 `*T` 指针——这是 Go 处理 nullable 字段的标准模式。

## 背景：为什么需要中转类型

Go 的 `*int64` 是真正的指针（nil = 没值），但 `database/sql` 的 `rows.Scan` 不能直接 Scan 到 `*int64`——必须 Scan 到 `sql.NullInt64`（带 `Valid` 标记）。

```go
// ❌ Scan 进 *int64 时 NULL 会报错（不能把 NULL 表达为 nil）
rows.Scan(&note.CategoryID)

// ✅ 正确：先 Scan 到 NullInt64
var categoryID sql.NullInt64
rows.Scan(&categoryID)
// 再 if Valid 转成 *int64
if categoryID.Valid {
    note.CategoryID = &categoryID.Int64
}
```

## 标准模式

```go
note := Note{ ID: noteID, Title: input.Title /* 其他字段 */ }

// 只有当 Valid=true 时才设指针；Valid=false 时指针保持 nil
if categoryIDArg.Valid {
    note.CategoryID = &categoryIDArg.Int64
}
// string 同理
if categoryNameArg.Valid {
    note.CategoryName = &categoryNameArg.String
}

return note, nil
```

## 易错点：`Int64` 字段的"永远有值"陷阱

```go
categoryIDArg := sql.NullInt64{}  // 零值：Valid=false, Int64=0

// 即使没赋值，categoryIDArg.Int64 也是 0
fmt.Println(categoryIDArg.Int64)  // 输出 0（不是 nil）
```

**关键**：`Int64` 永远有数字（默认 0），但 `Valid` 才是"DB 是不是 NULL"的真相。

**反面教材**（直接 `&categoryIDArg.Int64` 不 if Valid）：

```go
// ❌ 即使 Valid=false 也会返 0（不是 nil！）
return Note{
    CategoryID: &categoryIDArg.Int64,  // 总是返 *int64 指向 0
}, nil

// 后果：前端拿到 `categoryId: 0`
// - 前端 el-select 把 0 当"选了 id=0 的分类"（但分类 id 从 1 开始）
// - 前端过滤"未分类"用 `n.categoryId == null`——0 不匹配，漏掉所有无分类笔记
```

## 前端类比（关键差异）

| Go | TypeScript | 含义 |
|---|---|---|
| `*int64 == nil` | `x === null \|\| x === undefined` | "没值" |
| `*int64 != nil` | `x !== null` | "有值" |
| `*int64` 指向 0 | `x === 0` | "有值，且是 0" |

`if Valid` 模式 ≈ 前端 `x != null` 守门——但**顺序反着**：

```typescript
// 前端：if 兜底
const value = x ?? defaultValue
```

```go
// Go：if 显式赋值
if categoryIDArg.Valid {
    note.CategoryID = &categoryIDArg.Int64
}
// 否则保持 nil
```

## 完整工作流：INSERT/UPDATE/SELECT 都要走这套

```go
// INSERT：input.CategoryID *int64 → SQL NullInt64
categoryIDArg := sql.NullInt64{}
if input.CategoryID != nil {
    categoryIDArg = sql.NullInt64{ Int64: *input.CategoryID, Valid: true }
}
store.db.Exec(`INSERT INTO notes (..., category_id, ...) VALUES (..., ?, ...)`, ..., categoryIDArg, ...)

// SELECT：DB → NullInt64 → *int64
var categoryID sql.NullInt64
rows.Scan(&categoryID)
if categoryID.Valid {
    note.CategoryID = &categoryID.Int64
}

// UPDATE 同 INSERT
```

## 教学点（迁移价值）

1. **可空字段 = 指针**（不仅是参数，结构体字段也是）
2. **Scan 必须经过 NullX 中转**（不能跳过）
3. **NullX.Int64 永远有值**——必须 if Valid 才能信任
4. **CreateNote 这种"手 return Note{}"特别容易漏字段**——靠人记得所有 nullable 字段；UpdateNote 走 SELECT 自动反序列化所以安全

## 实战教训（dev-notebook Phase B）

- 写 `CreateNote` 时加 `CategoryID: &categoryIDArg.Int64`——**漏了 if Valid** → 无分类时返 0，前端过滤"未分类"漏数据
- 27 个测试全绿也抓不到这个 bug（测试只测"带分类"场景）——要靠 E2E curl 验各种 categoryId 场景才能发现
- 教训：**测 nullable 字段必须测"有/无/null"3 个场景**——单测不全 = silent fail
