---
id: go-sqlite-migrations
title: "SQLite 迁移、sqlite_master 与 CREATE TABLE IF NOT EXISTS"
category: Go
tags:
  - SQLite
  - database/sql
  - 事务
  - 迁移
summary: "迁移是在应用启动时以可重复执行的方式创建或升级数据库结构；当前项目首次打开 SQLite 时创建五张核心表。"
---

# SQLite 迁移、sqlite_master 与 CREATE TABLE IF NOT EXISTS

## 一句话结论

数据库迁移不是手动在工具里建表，而是由 Go 程序在启动时执行固定的 SQL，使空数据库自动变成应用需要的表结构。

## 当前项目的调用链

```text
Open(databasePath)
  -> sql.Open("sqlite", databasePath)
  -> Ping() 确认连接
  -> PRAGMA foreign_keys = ON
  -> migrate()
  -> CREATE TABLE IF NOT EXISTS ...
```

因此，无论测试创建临时 `.db` 文件，还是以后服务打开 `backend/data/dev-notes.db`，都会获得同一套表结构。

## CREATE TABLE IF NOT EXISTS

```sql
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    created_at TEXT NOT NULL
)
```

`IF NOT EXISTS` 的作用是让迁移可以重复执行：

- 第一次打开数据库：创建表。
- 之后再次启动：发现表已存在，继续运行而不是报错。

这种“重复执行结果相同”的性质称为幂等性，是迁移代码必须具备的能力。

## 为什么表有创建顺序

`workspaces.owner_user_id` 引用 `users.id`，`notes` 又引用工作空间和分类。因此迁移按以下顺序创建：

```text
users -> workspaces -> workspace_members -> categories -> notes
```

父表先创建、子表后创建，使外键关系清晰且便于维护。

## 外键和 PRAGMA

SQLite 默认可能不主动检查外键关系。项目打开数据库后执行：

```go
database.SetMaxOpenConns(1)
database.Exec(`PRAGMA foreign_keys = ON`)
```

`PRAGMA foreign_keys = ON` 开启外键约束，例如笔记不能指向不存在的工作空间。SQLite 的这个设置绑定到连接本身；第一版把连接数限制为 1，确保每次查询都使用已开启外键检查的连接。

## 为什么用事务

```go
transaction, err := store.db.Begin()
defer transaction.Rollback()

// 执行所有 CREATE TABLE

transaction.Commit()
```

事务保证迁移是一个整体：任意一条建表 SQL 出错，`Rollback()` 会撤销本次未完成的修改，不留下半套数据库结构。提交成功后再调用 `Rollback()` 不会撤销已提交内容，因此可以用 `defer` 兜底。

## sqlite_master 是什么

SQLite 把数据库里的表、索引等结构记录在系统表 `sqlite_master` 中。测试通过它验证表是否真实存在：

```go
database.db.QueryRow(
    `SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?`,
    tableName,
).Scan(&count)
```

## SQL 中的 ? 参数占位符

`?` 不是字符串拼接，而是参数位置。`tableName` 会由数据库驱动单独传入，避免把数据混进 SQL 文本。后续查询笔记、按分类筛选时也必须使用这种方式传递用户输入。
