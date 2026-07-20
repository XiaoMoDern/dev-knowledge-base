# 编辑笔记设计

## 目标

为现有笔记 API 增加编辑能力，并通过 TDD 理解 SQL UPDATE 语句、RowsAffected 判断记录是否存在，以及指针接收者在真业务里的用途。

## 接口

```text
PUT /api/notes/{id}
```

- 请求体是 JSON，包含 `title` 和 `content`。
- `id` 是路径参数，必须是正整数。
- 编辑成功返回 `200 OK`，响应体是更新后的笔记 JSON。
- `id` 格式错误返回 `400 Bad Request`。
- `title` 为空或空白返回 `400 Bad Request`。
- 笔记不存在返回 `404 Not Found`。
- 数据库错误返回 `500 Internal Server Error`。

## 调用链

```text
HTTP 请求
  -> ServeMux 路由
  -> notesHandler.update
  -> NotesStore.UpdateNote
  -> Store.UpdateNote
  -> SQLite UPDATE
  -> HTTP 状态码响应
```

HTTP 层负责解析路径参数、校验 title、选择状态码和输出 JSON；store 层负责执行 UPDATE、用 RowsAffected 判断记录是否存在、返回更新后的 Note。

## 实现顺序

1. 先为 `Store.UpdateNote` 编写会失败的测试（更新成功 + 更新不存在返回 `sql.ErrNoRows`）。
2. 增加最小的 UPDATE SQL 实现，让存储层测试通过。
3. 在 `NotesStore` 接口加 `UpdateNote`，并给 `fakeNotesStore` 补实现。
4. 为 HTTP PUT handler 增加路由和测试（200 / 400 / 404）。
5. 用 curl 验证成功编辑、格式错误、不存在、空标题四种情况。
6. 更新 Go 学习文档和学习日志，新增指针知识文章。

## 明确不做

本次不做部分更新（PATCH 语义）、categoryId 修改、乐观锁、编辑历史和软删除。title 和 content 必须同时提供。
