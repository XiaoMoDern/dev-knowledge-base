# 2026-07-22 Phase C 搜索 + 分页 设计文档

## 目标

为已有 notes 列表加：
- **搜索**（按关键字过滤 title + content）
- **分页**（避免一次返几百条，UI 也要翻页器）

让 dev-notebook 跨过"只能看全量"的阶段。

---

## Go 教学点（Phase C 重点）

1. **LIKE 查询**：`WHERE title LIKE ? OR content LIKE ?` + `%keyword%` 占位
2. **FTS5 全文索引**（可选）：倒排索引，性能 10-100x，但要建虚拟表
3. **LIMIT OFFSET 分页**：`LIMIT ? OFFSET ?` + 单独 `COUNT(*)` 算 total
4. **索引与 LIKE 关系**：`LIKE 'Go%'`（前缀）走索引，`LIKE '%Go%'`（包含）全表扫
5. **handler query 参数多参数解析**：`r.URL.Query().Get("q")` / `("page")` / `("pageSize")`
6. **count + page 两次查询**：分页要 total 才能算"共 X 页"——前端 el-pagination 需要
7. **QueryRow vs Query 选择**：count 用 QueryRow + Scan，list 用 Query + rows.Next

---

## 第一步任务（Ray 的，按能力台阶）

1. **我写完整测试** → Ray 跑、读、改
2. 3 个失败测试（TDD Red）：
   - `TestStoreSearchNotesByKeyword`：建 3 个 note，搜 "Go"，返含 "Go" 的
   - `TestStoreListsNotesWithPagination`：建 25 个 note，page=2 pageSize=10，返第 11-20 条
   - `TestStoreSearchAndPaginate`：建 30 个 note，搜 "Go" page=1，返匹配且分页

---

## API 计划

- `GET /api/notes?q=xxx&page=1&pageSize=20` —— 单 endpoint 覆盖搜索 + 分类 + 分页
- `GET /api/notes?categoryId=N` —— 已有
- 组合：`GET /api/notes?q=xxx&categoryId=N&page=1&pageSize=20` —— 4 维过滤
- 返回结构：
  ```json
  {
    "items": [...],
    "total": 42,
    "page": 1,
    "pageSize": 20
  }
  ```

---

## 前端改动

- NoteListView 顶部加 el-input 搜索框（debounce 300ms 触发）
- 底部加 el-pagination 翻页器（layout: total, prev, pager, next）
- URL 同步：搜索关键字 / 当前页同步到 query string（刷新不丢）
- 跟分类筛选互不冲突

---

## 待 Ray 拍板的决策（4 个）

### 决策 1：搜索实现

| 方案 | 优点 | 缺点 | 教学价值 |
| --- | --- | --- | --- |
| **A LIKE**（推荐） | 简单、SQL 基础、好懂 | 全表扫、1k+ 慢 | 学 LIKE / 通配符 / 索引关系 |
| B FTS5 | 性能 10-100x | 概念多（虚拟表 / MATCH / 倒排索引） | 进阶：全文检索基础 |

### 决策 2：分页方式

| 方案 | 优点 | 缺点 | 教学价值 |
| --- | --- | --- | --- |
| **A LIMIT OFFSET**（推荐） | 直观、UI 友好（el-pagination 直接接） | 深页性能差（OFFSET 10000 仍扫 10000 行） | 学基础分页 |
| B cursor-based | 性能恒定（只扫 LIMIT 行） | UI 复杂（不能跳页、只能"加载更多"） | 进阶：游标分页 |

### 决策 3：搜索范围

| 方案 | 优点 | 缺点 |
| --- | --- | --- |
| **A title + content**（推荐） | 覆盖完整 | 命中多、需 LIKE 两次（OR） |
| B only title | 性能稍好 | 用户搜"正文关键词"找不到 |

### 决策 4：分页参数命名

| 方案 | 适用 |
| --- | --- |
| **A `page` + `pageSize`**（推荐） | UI 友好（el-pagination 默认参数名） |
| B `offset` + `limit` | 数据库原生风格、cursor 分页常用 |

---

## 注意事项（铁律重复提醒）

1. **跨边界类型必须有 json tag** —— `PaginatedNotes` / `SearchResult` 加 tag（Phase A 教训）
2. **handler 参数默认值 + 边界**：`page < 1` 走 1，`pageSize > 100` 截到 100
3. **count 查询跟 list 查询在同一事务**（避免分页时 total 跟 items 不一致）
4. **LIKE 参数化**：`?` 占位 + 拼 `%`，不要字符串拼（注入风险）
5. **Element Plus 类型**：第一次 `npm run dev` 才会生成 el-pagination 的 d.ts

---

## Phase C 之后路线

- D：Markdown 导出 / 备份（io.Writer 流式、archive/zip）
- E：测试进阶（表驱动、testify、benchmark、SQLite 夹具）
- F：并发批处理（goroutine、channel、errgroup）—— 可选
