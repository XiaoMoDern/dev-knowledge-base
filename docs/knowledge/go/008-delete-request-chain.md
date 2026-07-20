# 删除请求与完整调用链

## DELETE 接口的目标

本项目使用真正删除：

```text
DELETE /api/notes/{id}
```

例如 `DELETE /api/notes/1` 表示删除 ID 为 `1` 的笔记。

## 分层职责

```text
前端或 Apifox
 -> ServeMux 路由
 -> HTTP handler
 -> Store 方法
 -> SQLite SQL
 -> Store 返回错误或结果
 -> handler 转成 HTTP 响应
```

- 路由：根据 HTTP 方法和路径找到处理函数。
- handler：解析路径参数、校验输入、选择状态码。
- store：执行数据库查询和删除，不负责 HTTP。
- SQLite：真正保存或删除数据。

## `RowsAffected`

`Exec` 返回的 `sql.Result` 可以调用 `RowsAffected()` 获取受影响的行数。

- `1`：确实删除了一条笔记。
- `0`：没有找到对应 ID，可以转换成 `sql.ErrNoRows`，让 HTTP 层返回 404。

## 状态码

- `204 No Content`：删除成功，响应体为空。
- `400 Bad Request`：路径中的 ID 不是合法正整数。
- `404 Not Found`：合法 ID，但对应笔记不存在。
- `500 Internal Server Error`：数据库等服务端错误。

## 调试位置

1. handler 中查看 `request.PathValue("id")`，确认路径参数。
2. `strconv.ParseInt` 后查看转换结果和错误。
3. store 中查看工作区 ID、SQL 执行错误和 `RowsAffected`。
4. handler 返回前查看最终 HTTP 状态码。

## 学习顺序

先写存储层失败测试，再写 `Store.DeleteNote`；然后扩展 `NotesStore`、路由和 HTTP 测试，最后用 Apifox 验证真实服务。
