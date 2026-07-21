import { apiGet, apiPost, apiPut, apiDelete } from './client'
import type { Note, NoteInput, NotesList, ImportNoteInput, ImportResult } from './types'

// GET /api/health —— 仅用于验证前后端打通
export function getHealth(): Promise<{ status: string }> {
  return apiGet<{ status: string }>('/api/health')
}

// GET /api/notes —— 列表
export function listNotes(): Promise<NotesList> {
  return apiGet<NotesList>('/api/notes')
}

// POST /api/notes —— 创建
export function createNote(input: NoteInput): Promise<Note> {
  return apiPost<Note>('/api/notes', input)
}

// PUT /api/notes/:id —— 编辑
export function updateNote(id: number, input: NoteInput): Promise<Note> {
  return apiPut<Note>(`/api/notes/${id}`, input)
}

// DELETE /api/notes/:id —— 删除（204 无 body）
export function deleteNote(id: number): Promise<void> {
  return apiDelete(`/api/notes/${id}`)
}

// POST /api/notes/import —— 批量导入
// 状态码：201 全成功 / 207 部分成功 / 400 全失败
export function importNotes(inputs: ImportNoteInput[]): Promise<ImportResult> {
  return apiPost<ImportResult>('/api/notes/import', { notes: inputs })
}