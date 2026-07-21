import { apiGet, apiPost, apiPut, apiDelete } from './client'
import type { Note, NoteInput, NotesList, ImportNoteInput, ImportResult } from './types'

// GET /api/health —— 仅用于验证前后端打通
export function getHealth(): Promise<{ status: string }> {
  return apiGet<{ status: string }>('/api/health')
}

// GET /api/notes —— 列表（全部分类）
export function listNotes(): Promise<NotesList> {
  return apiGet<NotesList>('/api/notes')
}

// GET /api/notes?categoryId=N —— 按分类过滤
// 后端约定：categoryId 必须是正整数（>0），0 / 不传 = 全部
// "未分类" 单独走 listNotes() + 前端本地过滤（后端没暴露"未分类"语义）
export function listNotesByCategory(categoryId: number): Promise<NotesList> {
  return apiGet<NotesList>(`/api/notes?categoryId=${categoryId}`)
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