import { apiGet, apiPost, apiPut, apiDelete } from './client'
import type { Note, NoteInput, PaginatedNotes, ImportNoteInput, ImportResult } from './types'

// GET /api/health —— 仅用于验证前后端打通
export function getHealth(): Promise<{ status: string }> {
  return apiGet<{ status: string }>('/api/health')
}

// searchNotesParams 是 GET /api/notes 的 4 维 query 参数
export interface searchNotesParams {
  q?: string           // 搜索关键字（title OR content），空 = 不过滤
  categoryId?: number  // 分类过滤，undefined = 不过滤
  page?: number        // 页码，默认 1
  pageSize?: number    // 每页条数，默认 20
}

// GET /api/notes —— 4 维过滤 + 分页统一接口
// 后端 handler 解析 4 维 query，调 store.SearchNotes
export function searchNotes(params: searchNotesParams = {}): Promise<PaginatedNotes> {
  const search = new URLSearchParams()
  if (params.q) search.set('q', params.q)
  if (params.categoryId !== undefined) search.set('categoryId', String(params.categoryId))
  if (params.page) search.set('page', String(params.page))
  if (params.pageSize) search.set('pageSize', String(params.pageSize))
  const query = search.toString()
  return apiGet<PaginatedNotes>(`/api/notes${query ? '?' + query : ''}`)
}

// 兼容旧调用：按分类过滤（NoteDetailView / 老代码可能用）
// 内部转发到 searchNotes({ categoryId })
export function listNotesByCategory(categoryId: number): Promise<PaginatedNotes> {
  return searchNotes({ categoryId })
}

// 兼容旧调用：拉全部分类笔记（NoteEditView 编辑模式复用、NoteDetailView 找引用等）
// 内部转发到 searchNotes() 无参（page=1 pageSize=20）
export function listNotes(): Promise<PaginatedNotes> {
  return searchNotes()
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
