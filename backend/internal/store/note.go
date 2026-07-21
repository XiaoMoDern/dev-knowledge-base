package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Note struct {
	ID         int64  `json:"id"`
	CategoryID *int64 `json:"categoryId,omitempty"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Visibility string `json:"visibility"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

type CreateNoteInput struct {
	Title   string
	Content string
}

// 编辑笔记的输入：标题和正文必须同时提供。
type UpdateNoteInput struct {
	Title   string
	Content string
}

// ImportNoteInput 是批量导入的单条输入
type ImportNoteInput struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// ImportError 记录一条导入失败原因
// Index 是请求数组里的位置（0-based），Title 是当时尝试的 title（可能为空）
type ImportError struct {
	Index  int    `json:"index"`
	Title  string `json:"title"`
	Reason string `json:"reason"`
}

// ImportResult 是批量导入的返回结果
// Imported/Failed 是计数；Items 是成功插入的完整 Note（已分配 ID）；Errors 是失败明细
type ImportResult struct {
	Imported int           `json:"imported"`
	Failed   int           `json:"failed"`
	Items    []Note        `json:"items"`
	Errors   []ImportError `json:"errors"`
}

// 创建笔记 CreateNote
func (store *Store) CreateNote(input CreateNoteInput) (Note, error) {
	//获取工作区ID,调用defaultWorkspaceID找到默认工作区
	workspaceID, err := store.defaultWorkspaceID()
	if err != nil {
		return Note{}, err
	}

	//生成时间戳
	now := time.Now().UTC().Format(time.RFC3339)
	//执行sql插入
	result, err := store.db.Exec(`
		INSERT INTO notes (workspace_id, title, content, visibility, created_at, updated_at)
		VALUES (?, ?, ?, 'private', ?, ?)
	`, workspaceID, input.Title, input.Content, now, now)
	if err != nil {
		return Note{}, fmt.Errorf("create note: %w", err)
	}

	noteID, err := result.LastInsertId()
	if err != nil {
		return Note{}, fmt.Errorf("read created note ID: %w", err)
	}

	return Note{
		ID:         noteID,
		Title:      input.Title,
		Content:    input.Content,
		Visibility: "private",
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

// 查询笔记列表 ListNotes
func (store *Store) ListNotes() ([]Note, error) {
	workspaceID, err := store.defaultWorkspaceID()
	if err != nil {
		return nil, err
	}

	rows, err := store.db.Query(`
		SELECT id, category_id, title, content, visibility, created_at, updated_at
		FROM notes
		WHERE workspace_id = ?
		ORDER BY updated_at DESC, id DESC
	`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}
	defer rows.Close()

	notes := make([]Note, 0)
	for rows.Next() {
		var note Note
		var categoryID sql.NullInt64

		if err := rows.Scan(
			&note.ID,
			&categoryID,
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

		notes = append(notes, note)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notes: %w", err)
	}

	return notes, nil
}

// 删除的笔记的实现
func (store *Store) DeleteNote(noteID int64) error {
	workspaceID, err := store.defaultWorkspaceID()
	if err != nil {
		return fmt.Errorf("find default workspace: %w", err)
	}

	result, err := store.db.Exec(`
		DELETE FROM notes
		WHERE id = ? AND workspace_id = ?
	`, noteID, workspaceID)

	if err != nil {
		return fmt.Errorf("delete note: %w", err)
	}

	affectedRows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read affected rows: %w", err)
	}

	if affectedRows == 0 {
		return sql.ErrNoRows
	}

	return nil

}

// 更新笔记
func (store *Store) UpdateNote(noteID int64, input UpdateNoteInput) (Note, error) {

	workspaceID, err := store.defaultWorkspaceID()
	if err != nil {
		return Note{}, err
	}

	// 每次更新都重新生成 updated_at
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := store.db.Exec(`
		UPDATE notes
		SET title = ?, content = ?, updated_at = ?
		WHERE id = ? AND workspace_id = ?
	`, input.Title, input.Content, now, noteID, workspaceID)
	if err != nil {
		return Note{}, fmt.Errorf("update note: %w", err)
	}

	// UPDATE 不会返回行内容，只返回"影响了几行"。
	affectedRows, err := result.RowsAffected()
	if err != nil {
		return Note{}, fmt.Errorf("read affected rows: %w", err)
	}
	if affectedRows == 0 {
		// 受影响行数为 0：要么 id 不存在，要么不属于当前工作空间。
		// 两种都按"找不到"处理，handler 转 404。
		return Note{}, sql.ErrNoRows
	}

	// UPDATE 只改了字段，没把"完整行"返回给 Go。
	// 要把 created_at、visibility、category_id 这些没被改的字段也带回给 handler，
	// 就再 SELECT 一次。
	var note Note
	var categoryID sql.NullInt64
	err = store.db.QueryRow(`
		SELECT id, category_id, title, content, visibility, created_at, updated_at
		FROM notes
		WHERE id = ? AND workspace_id = ?
	`, noteID, workspaceID).Scan(
		&note.ID,
		&categoryID,
		&note.Title,
		&note.Content,
		&note.Visibility,
		&note.CreatedAt,
		&note.UpdatedAt,
	)
	if err != nil {
		return Note{}, fmt.Errorf("read updated note: %w", err)
	}
	if categoryID.Valid {
		note.CategoryID = &categoryID.Int64
	}

	return note, nil
}

// ImportNotes 批量导入 notes：
//   - 业务校验：title 为空跳过，记到 Errors（不进事务）
//   - 剩下的 validNotes 用一个事务批量插入
//   - 全成功才 commit；任何 SQL 错就 rollback
//   - 返回 ImportResult 含计数 + 明细
func (store *Store) ImportNotes(notes []ImportNoteInput) (ImportResult, error) {
	// 初始化 result：Items 和 Errors 必须 make 出空 slice 而不是 nil
	// （nil slice 的 len() 是 0 但 append 会工作；空 slice 表达"没有"更明确）
	result := ImportResult{
		Items:  make([]Note, 0),
		Errors: make([]ImportError, 0),
	}

	workspaceID, err := store.defaultWorkspaceID()
	if err != nil {
		return result, fmt.Errorf("find default workspace: %w", err)
	}

	// 第一步：业务校验——把合法 note 收集到 validNotes，不合法的记 Errors
	validNotes := make([]ImportNoteInput, 0, len(notes))
	for i, note := range notes {
		trimmed := strings.TrimSpace(note.Title)
		if trimmed == "" {
			result.Failed++
			result.Errors = append(result.Errors, ImportError{
				Index:  i,
				Title:  note.Title,
				Reason: "title 不能为空",
			})
			continue
		}
		note.Title = trimmed
		validNotes = append(validNotes, note)
	}

	// 全部不合法：直接返回（不开事务，避免无意义的 SQL）
	if len(validNotes) == 0 {
		return result, nil
	}

	// 第二步：开事务批量插入
	// tx.Commit() 成功后 tx.Rollback() 是 no-op，所以 defer 写在前面最稳

	//return ImportResult{}, nil
	tx, err := store.db.Begin()
	if err != nil {
		return result, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now().UTC().Format(time.RFC3339)
	for _, note := range validNotes {
		execResult, err := tx.Exec(`
            INSERT INTO notes (workspace_id, title, content, visibility, created_at, updated_at)
            VALUES (?, ?, ?, 'private', ?, ?)
        `, workspaceID, note.Title, note.Content, now, now)
		if err != nil {
			return result, fmt.Errorf("insert note: %w", err)
		}
		// 拿刚插入的 ID
		noteID, err := execResult.LastInsertId()
		if err != nil {
			return result, fmt.Errorf("read note ID: %w", err)
		}

		// 构造 Note 写进 Items
		// 这里不二次 SELECT（UpdateNote 会 SELECT 是因为 UPDATE 不返回行内容；
		// INSERT 我们自己知道 Title/Content/visibility 是什么，直接构造省一次往返）
		result.Imported++
		result.Items = append(result.Items, Note{
			ID:         noteID,
			Title:      note.Title,
			Content:    note.Content,
			Visibility: "private",
			CreatedAt:  now,
			UpdatedAt:  now,
		})
	}
	if err := tx.Commit(); err != nil {
		return result, fmt.Errorf("commit transaction: %w", err)
	}

	return result, nil
}

// 获取默认工作区ID -> 这个方法通过用户名和工作区名查找工作区ID
func (store *Store) defaultWorkspaceID() (int64, error) {
	var workspaceID int64
	err := store.db.QueryRow(`
		SELECT w.id
		FROM workspaces w
		JOIN users u ON u.id = w.owner_user_id
		WHERE u.username = ? AND w.name = ?
		LIMIT 1
	`, defaultUsername, defaultWorkspaceName).Scan(&workspaceID)
	if err != nil {
		return 0, fmt.Errorf("find default workspace: %w", err)
	}

	return workspaceID, nil
}
