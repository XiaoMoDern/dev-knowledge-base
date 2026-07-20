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
