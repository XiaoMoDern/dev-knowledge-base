---
id: go-sql-joins-and-nullable-scan
title: "Go SQL JOIN 与 nullable Scan：从单表到联表的 4 个铁律"
category: Go
tags:
  - Go
  - database/sql
  - SQL JOIN
  - nullable
  - Scan
summary: "SQL 别名避免歧义、LEFT JOIN 保留无分类、sql.NullX 中转 NULL、Scan 顺序敏感——4 个铁律实战 dev-notebook Phase B 分类落地。"
---

# Go SQL JOIN 与 nullable Scan

dev-notebook Phase B 把 `notes` 表跟 `categories` 表 JOIN 起来，让前端能一次拿到 `note` + `categoryName`。这一篇文章把踩过的 4 个铁律放一起。

## 铁律 1：SQL 别名避免歧义

```sql
SELECT n.id, n.category_id, c.name AS category_name, ...
FROM notes n
LEFT JOIN categories c ON n.category_id = c.id
WHERE n.workspace_id = ?
```

- `n` 是 `notes` 的别名、`c` 是 `categories` 的别名
- 不用别名就要写全 `notes.id` / `notes.category_id` / `categories.id` / `categories.name`——又长又容易错
- 强别名让"id 到底是谁的 id"这种歧义不存在
- 前端类比：跟 ES6 解构的 rename 同款——`const { id: userId } = user`

## 铁律 2：LEFT JOIN vs INNER JOIN（按 nullable 选）

| JOIN | 语义 | 类比 |
| --- | --- | --- |
| `LEFT JOIN` | 没分类的 note 也保留，`category_name` = NULL | JS `Array.find` 找不到返回 undefined |
| `INNER JOIN` | 没分类的 note 直接被过滤掉 | JS `Array.filter(n => n.category)` |

`Note` 可能没分类（`category_id` 是 nullable）—— **必须 LEFT JOIN**，否则"没分类的笔记"会从列表消失。

## 铁律 3：sql.NullX 中转 NULL

Go database/sql **不能**直接 Scan 到 `*string`（不像 `*int64`）。必须用 `sql.NullString` 中转：

```go
var categoryName sql.NullString
rows.Scan(..., &categoryName, ...)
// ...
if categoryName.Valid {
    note.CategoryName = &categoryName.String
}
```

| 类型 | 中转类型 |
| --- | --- |
| `*int64` | `sql.NullInt64` |
| `*string` | `sql.NullString` |
| `*float64` | `sql.NullFloat64` |
| `*time.Time` | `sql.NullTime` |
| `*bool` | `sql.NullBool` |

**关键陷阱**：`sql.NullInt64` 的 `Int64` 字段**永远有值**（默认 0），但 `Valid: false` 表示"DB 是 NULL"——`omitempty` 不会触发，因为指针非 nil。

**前端类比**：跟 `x ?? defaultValue` 反着——Go 是"如果有效就用值"（主动检查），前端是"如果 null 就用兜底"（被动 fallback）。

## 铁律 4：Scan 顺序敏感性

`rows.Scan(&a, &b, &c, ...)` 按**列出现顺序**赋值（不是列名）。**SQL 加了一列，Scan 必须同步加一个变量**。

```sql
SELECT n.id, n.category_id, c.name AS category_name, n.title, ...
```
```go
rows.Scan(
    &note.ID,        // ← n.id
    &categoryID,     // ← n.category_id
    &categoryName,   // ← c.name AS category_name（必须按这个顺序）
    &note.Title,     // ← n.title
    ...
)
```

顺序**必须**完全一致。错一个 = 全部错位 + 难调试。

**踩坑症状**：`Scan` 时列数对不上 = panic "sql: expected N destination arguments, got M"。

## 实战：dev-notebook Phase B 的 `ListNotes`

```go
rows, err := store.db.Query(`
    SELECT n.id, n.category_id, c.name AS category_name, n.title, n.content, n.visibility, n.created_at, n.updated_at
    FROM notes n
    LEFT JOIN categories c ON n.category_id = c.id
    WHERE n.workspace_id = ?
    ORDER BY n.updated_at DESC, n.id DESC
`, workspaceID)
defer rows.Close()

notes := make([]Note, 0)
for rows.Next() {
    var note Note
    var categoryID sql.NullInt64
    var categoryName sql.NullString  // 铁律 3：中转类型

    if err := rows.Scan(
        &note.ID,
        &categoryID,
        &categoryName,  // 铁律 4：按 SQL 列顺序
        &note.Title,
        &note.Content,
        &note.Visibility,
        &note.CreatedAt,
        &note.UpdatedAt,
    ); err != nil {
        return nil, fmt.Errorf("scan note: %w", err)
    }
    if categoryID.Valid {
        note.CategoryID = &categoryID.Int64
    }
    if categoryName.Valid {
        note.CategoryName = &categoryName.String
    }

    notes = append(notes, note)
}
```

## 检查清单（改 SQL JOIN 时）

- [ ] SQL 用别名 `n` / `c` 了吗？
- [ ] LEFT JOIN（不是 INNER）保留 nullable 行了吗？
- [ ] Scan 变量数 = SELECT 列数（不含字面量）？
- [ ] `sql.NullX` 中转类型正确（`NullString` / `NullInt64` / ...）？
- [ ] `if x.Valid { 赋值给 *T }` 模式两个字段都加了吗？
- [ ] **跑全量测试**（不只是单测）—— 改 SQL 必破 1-2 个测试

## 关联知识点

- `010-if-valid-pattern.md` —— if Valid 模式详细
- `009-pointers.md` —— `*int64` / `*string` 跟 NULL 的关系
- `007-go-slices.md` —— `make([]Note, 0)` 不为 nil 模式
