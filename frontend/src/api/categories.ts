import { apiGet, apiPost } from './client'
import type { Category, CategoriesList, CreateCategoryInput } from './types'

// GET /api/categories —— 列表
export function listCategories(): Promise<CategoriesList> {
  return apiGet<CategoriesList>('/api/categories')
}

// POST /api/categories —— 创建
export function createCategory(input: CreateCategoryInput): Promise<Category> {
  return apiPost<Category>('/api/categories', input)
}
