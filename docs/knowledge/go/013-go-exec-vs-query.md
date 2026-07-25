---
id: go-exec-vs-query
title: "Go database/sql Exec vs Query 边界：改用 Exec、查用 Query"
category: Go
tags:
  - Go
  - database/sql
  - Exec
  - Query
  - QueryRow
  - Scan
summary: "INSERT/UPDATE/DELETE 用 Exec（没 Scan 方法），SELECT 用 Query/QueryRow（有 Scan）；'下一步用跟第一步同款'是错的——每步看 SQL 本质。"
---

# Go Exec vs Query 边界

dev-notebook Phase B/C 实战踩过：Ray 改完 CreateNote 的 INSERT 后，下一步想"用跟第一步同款方法"写 SELECT，结果 SELECT 用 Exec 写——**编译都过不了**（Exec 没 Scan 方法）。

## 一句话结论

> 改 = `Exec`（INSERT/UPDATE/DELETE，无返回值）。查 = `Query` / `QueryRow`（SELECT，有行要 Scan）。每步看 SQL 本质，**不要套用上一步的方法**。

## 三个方法的对比

| 方法 | 适用 SQL | 返回值 | 有 Scan？ | 有 Rows？ |
| --- | --- | --- | --- | --- |
| **`db.Exec(...)`** | INSERT / UPDATE / DELETE | `sql.Result{ LastInsertId, RowsAffected }` | ❌ | ❌ |
| **`db.QueryRow(...)`** | SELECT 单行 | `*sql.Row` | ✅（`.Scan(...)`） | ❌（自动关闭） |
| **`db.Query(...)`** | SELECT 多行 | `*sql.Rows, error` | ✅（`rows.Scan(...)`） | ✅（`rows.Next()` 循环） |

## 错例 1：SELECT 用 Exec

```go
// 错：SELECT 用 Exec
result, err := store.db.Exec("SELECT id, title FROM notes WHERE id = ?", id)
// ↑ 编译不报（Exec 接任意 SQL），但 result.LastInsertId() 是 0、result.RowsAffected() 是 0
// 想拿数据？没 Scan 方法
```

**症状**：result 拿到但没数据，运行时"看起来对、跑就崩"。

## 错例 2：QueryRow 不 Scan

```go
// 错：QueryRow 没 Scan
row := store.db.QueryRow("SELECT id, title FROM notes WHERE id = ?", id)
fmt.Println(row)  // ← &{...}，不是数据
```

**症状**：拿到 `*sql.Row` 对象本身，不是数据。

## 正确模式

### INSERT（Exec）

```go
result, err := store.db.ExecContext(ctx,
    `INSERT INTO notes (workspace_id, title, content) VALUES (?, ?, ?)`,
    workspaceID, title, content,
)
if err != nil {
    return 0, fmt.Errorf("insert note: %w", err)
}
id, err := result.LastInsertId()
if err != nil {
    return 0, fmt.Errorf("last insert id: %w", err)
}
return id, nil
```

### SELECT 单行（QueryRow + Scan）

```go
var note Note
var categoryID sql.NullInt64
err := store.db.QueryRowContext(ctx,
    `SELECT id, category_id, title, content FROM notes WHERE id = ?`,
    id,
).Scan(
    &note.ID,
    &categoryID,
    &note.Title,
    &note.Content,
)
if err == sql.ErrNoRows {
    return Note{}, store.ErrNotFound  // 没找到
}
if err != nil {
    return Note{}, fmt.Errorf("query note: %w", err)
}
if categoryID.Valid {
    note.CategoryID = &categoryID.Int64
}
return note, nil
```

### SELECT 多行（Query + rows.Next + Scan）

```go
rows, err := store.db.QueryContext(ctx,
    `SELECT id, category_id, title FROM notes WHERE workspace_id = ?`,
    workspaceID,
)
if err != nil {
    return nil, fmt.Errorf("query notes: %w", err)
}
defer rows.Close()

notes := make([]Note, 0)
for rows.Next() {
    var note Note
    var categoryID sql.NullInt64
    if err := rows.Scan(&note.ID, &categoryID, &note.Title); err != nil {
        return nil, fmt.Errorf("scan note: %w", err)
    }
    if categoryID.Valid {
        note.CategoryID = &categoryID.Int64
    }
    notes = append(notes, note)
}
if err := rows.Err(); err != nil {
    return nil, fmt.Errorf("rows err: %w", err)
}
return notes, nil
```

## Ray 07-22 实战踩的坑

**场景**：CreateNote 留了 trade-off bug——新建 note 后列表 tag 不显示分类。修法要 SELECT 出 CategoryName 返回给前端。

**Ray 质疑**："INSERT 用 Exec，为什么 SELECT 不用 Exec？"

**正解**：
- INSERT 是改 → Exec
- SELECT 是查 → **必须** QueryRow（单行）或 Query（多行），没第三种
- Exec 没 Scan 方法，编译都过不了（"Exec(...) 没用返回值就 Discard"不算）

**教训**（dev-notebook 教训 10 + 跨项目教训 11）：
- "下一步用跟第一步同款方法"是错的——每步看 SQL 本质
- "质疑比盲信强"是对的，但**质疑要质疑对方向**——"为什么不用 Exec"方向错了，应该是"为什么用 Query"

## 三步自检

1. **看 SQL 关键字**：INSERT/UPDATE/DELETE → Exec；SELECT → Query/QueryRow
2. **看是否要拿数据**：要拿 → 必须有 Scan
3. **看是否多行**：单行 → QueryRow（自动关 rows）；多行 → Query + rows.Next() + defer rows.Close()

## 常见误区

- **"我 QueryRow 不 Scan，直接把 row 当数据用"**——`*sql.Row` 是惰性 Scan 占位符，不 Scan 不会执行
- **"Exec 接 SELECT 也能编译"**——能编译能跑，但拿不到数据（result.LastInsertId 永远 0）
- **"defer rows.Close() 没必要"**——连接池泄漏的常见原因，不 Close 会"看似工作、并发时崩"
- **"QueryRow 自动关 rows，不用 defer"**——对，单行确实不用，但 Query 必须 defer Close

## 调试位置

- **"Exec 没报错但数据没插入"**——SQL 关键字写错（INSERT 写成 INSET 等）
- **"QueryRow.Scan 报 'sql: no Rows in result set'"**——SQL 没匹配行（`sql.ErrNoRows` 特殊处理）
- **"rows.Scan 报 'expected N destination arguments, got M'"**——SELECT 列数跟 Scan 变量数对不上（铁律 4，见 go/011）
- **"连接池 timeout"**——`rows.Close()` 漏了，并发时连接占满

## 关联知识点

- `011-sql-joins-and-nullable-scan.md` — Scan 顺序敏感性、sql.NullX 中转
- `010-if-valid-pattern.md` — if Valid 模式跟 nullable 配合
- `008-delete-request-chain.md` — DELETE 走 Exec + handler 204
