# Dev Notebook 数据模型

## 目标

第一版只运行一个本地用户，但数据关联从一开始支持未来的公开阅读、用户注册和工作空间协作。

## 实体关系

```text
User
  -> owns Workspace
  -> joins Workspace through WorkspaceMember

Workspace
  -> contains Category
  -> contains Note

Category
  -> groups Note

Note
  -> has visibility: private or public
  -> may have one Category
```

## 表设计

### users

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | INTEGER | 主键 |
| username | TEXT | 用户名，当前初始化为本地用户 |
| created_at | TEXT | 创建时间 |

第一版不实现登录，但保留用户实体，避免以后给已有笔记补所有者关系。

### workspaces

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | INTEGER | 主键 |
| name | TEXT | 工作空间名称，例如“我的知识库” |
| owner_user_id | INTEGER | 创建者用户 ID |
| created_at | TEXT | 创建时间 |

### workspace_members

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| workspace_id | INTEGER | 工作空间 ID |
| user_id | INTEGER | 用户 ID |
| role | TEXT | `owner`、`editor` 或 `viewer` |

第一版只初始化一条 owner 关系；后续再实现邀请成员的界面和接口。

### categories

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | INTEGER | 主键 |
| workspace_id | INTEGER | 所属工作空间 |
| name | TEXT | 分类名称，例如 Go、Vue |
| created_at | TEXT | 创建时间 |

同一工作空间内分类名唯一，不同工作空间可使用相同分类名。

### notes

| 字段 | 类型 | 说明 |
| --- | --- | --- |
| id | INTEGER | 主键 |
| workspace_id | INTEGER | 所属工作空间 |
| category_id | INTEGER，可为空 | 所属分类 |
| title | TEXT | 笔记标题 |
| content | TEXT | Markdown 正文 |
| visibility | TEXT | `private` 或 `public` |
| created_at | TEXT | 创建时间 |
| updated_at | TEXT | 最后修改时间 |

## 第一版初始化数据

数据库第一次启动时创建：

1. 一个本地用户。
2. 一个名为“我的知识库”的工作空间。
3. 该用户作为工作空间 `owner` 的成员关系。

应用界面暂时不显示用户、工作空间切换或权限管理，但所有笔记会属于默认工作空间。

## 暂不实现

- 密码、登录、注册和会话。
- 文件附件和上传。
- 公开笔记页面。
- 工作空间成员邀请。

这些能力会直接复用上述关系，不需要改变笔记和分类的归属结构。
