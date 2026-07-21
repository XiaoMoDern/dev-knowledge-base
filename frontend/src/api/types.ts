// 与后端 store.Note 字段名 1:1 对齐（json tag）
export interface Note {
  id: number
  categoryId?: number
  title: string
  content: string
  visibility: 'private' | 'public'
  createdAt: string
  updatedAt: string
}

// 创建/编辑笔记的入参：title + content 同时提供
export interface NoteInput {
  title: string
  content: string
}

// GET /api/notes 列表响应
export interface NotesList {
  items: Note[]
}

// POST /api/notes/import 入参的单条 note；后端会校验 title 非空
export interface ImportNoteInput {
  title: string
  content: string
}

// 后端返回的单条错误明细：index 是请求数组里 0-based 的位置
export interface ImportError {
  index: number
  title: string
  reason: string
}

// POST /api/notes/import 响应
// 状态码语义：201 全成功 / 207 部分成功 / 400 全失败
export interface ImportResult {
  imported: number
  failed: number
  items: Note[]
  errors: ImportError[]
}
