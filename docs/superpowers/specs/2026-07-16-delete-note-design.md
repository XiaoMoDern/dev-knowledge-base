# 删除笔记设计

## 目标

为现有笔记 API 增加真正删除笔记的能力，并通过一步一步的测试和实现理解完整调用链。

## 接口

```text
DELETE /api/notes/{id}
```

- `id` 是路径参数，必须是正整数。
- 删除成功返回 `204 No Content`，响应体为空。
- `id` 格式错误返回 `400 Bad Request`。
- 笔记不存在返回 `404 Not Found`。
- 数据库错误返回 `500 Internal Server Error`。

## 调用链

```text
HTTP 请求
 -> ServeMux 路由
 -> notesHandler.delete
 -> NotesStore.DeleteNote
 -> Store.DeleteNote
 -> SQLite DELETE
 -> HTTP 状态码响应
```

HTTP 层负责解析路径参数、选择状态码和输出 JSON；store 层负责查询工作区、执行 SQL 和返回数据库错误。

## 实现顺序

1. 先为 `Store.DeleteNote` 编写会失败的测试。
2. 增加最小的删除 SQL 实现，让存储层测试通过。
3. 为 HTTP DELETE handler 增加路由和测试。
4. 用 Apifox 验证成功删除、格式错误和不存在三种情况。
5. 更新 Go 学习文档和学习日志。

## 明确不做

本次不做软删除、恢复、批量删除、权限系统和前端页面。
