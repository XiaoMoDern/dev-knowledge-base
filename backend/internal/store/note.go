package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Note struct {
	ID           int64   `json:"id"`
	CategoryID   *int64  `json:"categoryId,omitempty"`
	CategoryName *string `json:"categoryName,omitempty"`
	Title        string  `json:"title"`
	Content      string  `json:"content"`
	Visibility   string  `json:"visibility"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
}

type CreateNoteInput struct {
	Title      string
	Content    string
	CategoryID *int64
}

// 编辑笔记的输入：标题和正文必须同时提供。
type UpdateNoteInput struct {
	Title      string
	Content    string
	CategoryID *int64
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
	now := time.Now().UTC().Format(time.RFC3339Nano)
	categoryIDArg := sql.NullInt64{}
	if input.CategoryID != nil {
		categoryIDArg = sql.NullInt64{Int64: *input.CategoryID, Valid: true}
	}

	//执行sql插入
	result, err := store.db.Exec(`
		INSERT INTO notes (workspace_id, category_id, title, content, visibility, created_at, updated_at)
		VALUES (?, ?, ?, ?, 'private', ?, ?)
	`, workspaceID, categoryIDArg, input.Title, input.Content, now, now)
	if err != nil {
		return Note{}, fmt.Errorf("create note: %w", err)
	}

	noteID, err := result.LastInsertId()
	if err != nil {
		return Note{}, fmt.Errorf("read created note ID: %w", err)
	}

	var note Note
	var categoryID sql.NullInt64
	var categoryName sql.NullString

	err = store.db.QueryRow(`
    SELECT n.id, n.category_id, c.name AS category_name, n.title, n.content, n.visibility, n.created_at, n.updated_at
    FROM notes n
    LEFT JOIN categories c ON n.category_id = c.id
    WHERE n.id = ? AND n.workspace_id = ?
	`, noteID, workspaceID).Scan(
		&note.ID,
		&categoryID,
		&categoryName,
		&note.Title,
		&note.Content,
		&note.Visibility,
		&note.CreatedAt,
		&note.UpdatedAt,
	)
	if err != nil {
		return Note{}, fmt.Errorf("read created note: %w", err)
	}
	if categoryID.Valid {
		note.CategoryID = &categoryID.Int64
	}
	if categoryName.Valid {
		note.CategoryName = &categoryName.String
	}

	return note, nil
}

// 查询笔记列表 ListNotes
func (store *Store) ListNotes() ([]Note, error) {
	workspaceID, err := store.defaultWorkspaceID()
	if err != nil {
		return nil, err
	}

	// 单表查询
	// SELECT id, category_id, title, content, visibility, created_at, updated_at
	// FROM notes
	// WHERE workspace_id = ?
	// ORDER BY updated_at DESC, id DESC

	// 联表查询 （后端把数据拼好，前端只负责展示）
	rows, err := store.db.Query(`
		SELECT n.id, n.category_id, c.name AS category_name, n.title, n.content, n.visibility, n.created_at, n.updated_at
		FROM notes n
		LEFT JOIN categories c ON n.category_id = c.id
		WHERE n.workspace_id = ?
		ORDER BY n.updated_at DESC, n.id DESC
	`, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("list notes: %w", err)
	}
	defer rows.Close()

	notes := make([]Note, 0)
	for rows.Next() {
		var note Note
		var categoryID sql.NullInt64
		var categoryName sql.NullString

		if err := rows.Scan(
			&note.ID,
			&categoryID,
			&categoryName,
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

	//Exec  →  Execute（执行命令，不期望返回数据）
	//Query →  查询（期望返回数据）
	//增删改 → Exec
	//查     → Query
	workspaceID, err := store.defaultWorkspaceID()
	if err != nil {
		return Note{}, err
	}

	// 每次更新都重新生成 updated_at
	now := time.Now().UTC().Format(time.RFC3339Nano)
	var result sql.Result

	if input.CategoryID != nil {
		result, err = store.db.Exec(`
			UPDATE notes
			SET title = ?, content = ?, category_id = ?, updated_at = ?
			WHERE id = ? AND workspace_id = ?
		`, input.Title, input.Content, *input.CategoryID, now, noteID, workspaceID)
	} else {
		result, err = store.db.Exec(`
			UPDATE notes
			SET title = ?, content = ?, updated_at = ?
			WHERE id = ? AND workspace_id = ?
		`, input.Title, input.Content, now, noteID, workspaceID)
	}

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
	var categoryName sql.NullString

	err = store.db.QueryRow(`
		SELECT n.id, n.category_id, c.name AS category_name, n.title, n.content, n.visibility, n.created_at, n.updated_at
	    FROM notes n
		LEFT JOIN categories c ON n.category_id = c.id
		WHERE n.id = ? AND n.workspace_id = ?
		ORDER BY n.updated_at DESC ,n.id DESC 
	`, noteID, workspaceID).Scan(
		&note.ID,
		&categoryID,
		&categoryName,
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
	if categoryName.Valid {
		note.CategoryName = &categoryName.String
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

	now := time.Now().UTC().Format(time.RFC3339Nano)
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

func (store *Store) ListNotesByCategory(categoryID int64) ([]Note, error) {
	workspaceID, err := store.defaultWorkspaceID()
	if err != nil {
		return nil, err
	}

	rows, err := store.db.Query(`
		SELECT n.id, n.category_id, c.name AS category_name, n.title, n.content, n.visibility, n.created_at, n.updated_at
		FROM notes n
		LEFT JOIN categories c ON n.category_id = c.id
		WHERE n.workspace_id = ? AND n.category_id = ?
		ORDER BY n.updated_at DESC, n.id DESC
	`, workspaceID, categoryID)
	if err != nil {
		return nil, fmt.Errorf("list notes by category: %w", err)
	}
	defer rows.Close()

	notes := make([]Note, 0)
	for rows.Next() {
		// 参数 categoryID 跟这里同名——Go 的 shadow（内层覆盖外层）
		var note Note
		var categoryID sql.NullInt64
		var categoryName sql.NullString

		if err := rows.Scan(
			&note.ID,
			&categoryID,
			&categoryName,
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
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notes: %w", err)
	}

	return notes, nil
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

// SearchOptions 是 SearchNotes 的输入参数：4 维过滤
type SearchOptions struct {
	Query      string // 搜索关键字（title OR content），空 = 不过滤
	CategoryID *int64 // 分类 ID，nil = 不过滤
	Page       int    // 页码（从 1 开始），1 = 第一页
	PageSize   int    // 每页条数
}

// PaginatedNotes 是 SearchNotes 的返回结果：分页 items + total + page/pageSize 元数据
type PaginatedNotes struct {
	Items    []Note `json:"items"`
	Total    int    `json:"total"`
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
}

// SearchNotes 按 SearchOptions 多维过滤 + 分页查询 notes
// 当前还没实现 —— Red
func (store *Store) SearchNotes(opts SearchOptions) (PaginatedNotes, error) {
	workspaceID, err := store.defaultWorkspaceID()
	if err != nil {
		return PaginatedNotes{}, err
	}

	whereClause := []string{"n.workspace_id = ?"}
	args := []any{workspaceID}

	if opts.Query != "" {
		// %keyword% 包含匹配：注意用 fmt.Sprintf 拼 %，不要拼到 SQL 字符串
		whereClause = append(whereClause, fmt.Sprintf("n.title LIKE ? OR n.content LIKE ?"))
		pattern := "%" + opts.Query + "%"
		args = append(args, pattern, pattern)
	}

	if opts.CategoryID != nil {
		whereClause = append(whereClause, "n.category_id = ?")
		args = append(args, *opts.CategoryID)
	}

	whereSQL := strings.Join(whereClause, " AND ")

	// 第二步：count 查询（不带 JOIN，只算 notes 数量）
	var total int
	countSQL := "SELECT COUNT(*) FROM notes n WHERE " + whereSQL
	if err := store.db.QueryRow(countSQL, args...).Scan(&total); err != nil {
		return PaginatedNotes{}, fmt.Errorf("count notes: %w", err)
	}

	// 第三步：page/pageSize 边界处理
	page := opts.Page
	if page < 1 {
		page = 1
	}
	pageSize := opts.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	offset := (page - 1) * pageSize

	// 第四步：list 查询（带 JOIN 拿 CategoryName + LIMIT OFFSET）
	listSQL := `
		SELECT n.id, n.category_id, c.name AS category_name, n.title, n.content, n.visibility, n.created_at, n.updated_at
		FROM notes n
		LEFT JOIN categories c ON n.category_id = c.id
		WHERE ` + whereSQL + `
		ORDER BY n.updated_at DESC, n.id DESC
		LIMIT ? OFFSET ?
	`
	listArgs := append(args, pageSize, offset)

	rows, err := store.db.Query(listSQL, listArgs...)
	if err != nil {
		return PaginatedNotes{}, fmt.Errorf("list notes: %w", err)
	}
	defer rows.Close()

	items := make([]Note, 0)
	for rows.Next() {
		var note Note
		var categoryID sql.NullInt64
		var categoryName sql.NullString

		if err := rows.Scan(
			&note.ID,
			&categoryID,
			&categoryName,
			&note.Title,
			&note.Content,
			&note.Visibility,
			&note.CreatedAt,
			&note.UpdatedAt,
		); err != nil {
			return PaginatedNotes{}, fmt.Errorf("scan note: %w", err)
		}

		if categoryID.Valid {
			note.CategoryID = &categoryID.Int64
		}
		if categoryName.Valid {
			note.CategoryName = &categoryName.String
		}

		items = append(items, note)
	}

	if err := rows.Err(); err != nil {
		return PaginatedNotes{}, fmt.Errorf("iterate notes: %w", err)
	}

	return PaginatedNotes{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil

}

// GetNoteByID 按 ID 精确查单条 note，找不到返 sql.ErrNoRows
//
// 为什么需要这个方法：
//   之前详情页/编辑页调 listNotes() 拿全表（第 1 页 20 条）再在内存 find(id)，
//   第 21 条 + 之后全部 "not found"。本方法让前端按 ID 精确查，绕过内存 find。
//
// 为什么 err 不 wrap：
//   handler 用 errors.Is(err, sql.ErrNoRows) 判 404，
//   如果 wrap（%w 包一层），errors.Is 仍能匹配，但增加一层没必要。
//   跟 line 106（CreateNote 内的回读）不同——那里是内部环节，wrap 排错有信息量；
//   这里是"接力"给 handler，原样返最干净。

func (store *Store) GetNoteByID(noteID int64) (Note, error) {
	workspaceID, err := store.defaultWorkspaceID()
	if err != nil {
		return Note{}, err
	}

	var note Note
	var categoryID sql.NullInt64
	var categoryName sql.NullString

	err = store.db.QueryRow(`
        SELECT n.id, n.category_id, c.name AS category_name, n.title, n.content, n.visibility, n.created_at, n.updated_at
        FROM notes n
        LEFT JOIN categories c ON n.category_id = c.id
        WHERE n.id = ? AND n.workspace_id = ?
    `, noteID, workspaceID).Scan(
		&note.ID,
		&categoryID,
		&categoryName,
		&note.Title,
		&note.Content,
		&note.Visibility,
		&note.CreatedAt,
		&note.UpdatedAt,
	)

	if err != nil {
		return Note{}, err
	}

	if categoryID.Valid {
		note.CategoryID = &categoryID.Int64
	}
	if categoryName.Valid {
		note.CategoryName = &categoryName.String
	}

	return note, nil
}
