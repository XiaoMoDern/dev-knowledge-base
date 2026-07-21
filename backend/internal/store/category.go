package store

import (
	"fmt"
	"time"
)

// Category 表示工作空间下的一个笔记分类。
// 跟 Note 比只少 visibility 业务字段——CRUD 形态一致。
type Category struct {
	ID          int64  `json:"id"`
	WorkspaceID int64  `json:"workspaceId"`
	Name        string `json:"name"`
	CreatedAt   string `json:"createdAt"`
}

// CreateCategory 在默认工作空间下创建一个分类。
// 教学对照：跟 CreateNote 完全同构——
// 拿 workspace_id → 生成 now → INSERT → LastInsertId → 构造返回。
func (store *Store) CreateCategory(name string) (Category, error) {
	// 1. 拿默认工作空间 ID（跟 note.go 一样，复用"我的知识库"概念）
	workspaceID, err := store.defaultWorkspaceID()
	if err != nil {
		return Category{}, err
	}

	// 2. 生成当前时间（RFC3339Nano 跟 note.go 保持一致，避免刚才修的 bug）
	now := time.Now().UTC().Format(time.RFC3339Nano)

	// 3. INSERT——? 是 SQLite 参数化查询占位符，防 SQL 注入
	result, err := store.db.Exec(`
		INSERT INTO categories (workspace_id, name, created_at)
		VALUES (?, ?, ?)
	`, workspaceID, name, now)
	if err != nil {
		return Category{}, fmt.Errorf("create category: %w", err)
	}

	// 4. 拿刚插入的 ID——SQLite 自带，前端类比 prisma.category.create() 返回的 id
	categoryID, err := result.LastInsertId()
	if err != nil {
		return Category{}, fmt.Errorf("read created category ID: %w", err)
	}

	// 5. 构造返回值——存什么、还什么，不要漏字段
	return Category{
		ID:          categoryID,
		WorkspaceID: workspaceID,
		Name:        name,
		CreatedAt:   now,
	}, nil
}

// ListCategories 列出默认工作空间下的所有分类，按 created_at 倒序（新的在前）。
// 教学对照：跟 ListNotes 完全同构——
// defaultWorkspaceID → Query → rows.Next 循环 + Scan → 累加到 slice。
func (store *Store) ListCategories() ([]Category, error) {
	workspaceID, err := store.defaultWorkspaceID()
	if err != nil {
		return nil, err
	}

	// Query 跟 Exec 的区别：Exec 改数据不返行，Query 查数据返多行
	rows, err := store.db.Query(`
		SELECT id, workspace_id, name, created_at
		FROM categories
		WHERE workspace_id = ?
		ORDER BY created_at DESC, id DESC
	`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	// defer rows.Close() 释放连接——rows 不 Close 会泄漏
	defer rows.Close()

	// make 出空 slice 而非 nil——避免 JSON 序列化时变 null（跟 ImportResult 同款）
	categories := make([]Category, 0)
	for rows.Next() {
		var category Category
		// Scan 按 SELECT 顺序把列值塞进字段
		if err := rows.Scan(
			&category.ID,
			&category.WorkspaceID,
			&category.Name,
			&category.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		categories = append(categories, category)
	}
	// 循环结束后再检查一次 rows.Err()——捕获迭代过程中被中断的错误
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate categories: %w", err)
	}

	return categories, nil
}
