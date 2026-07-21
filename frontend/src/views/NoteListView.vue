<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listNotes, listNotesByCategory, deleteNote } from '../api/notes'
import { listCategories } from '../api/categories'
import type { Note, Category, ImportResult } from '../api/types'
import ImportDialog from '../components/ImportDialog.vue'

// 筛选状态用 string 哨兵 + 数字 id 三态
// 不用 0 当"未分类"哨兵——避免和真实 id 冲突
const ALL = 'all'
const UNCLASSIFIED = 'unclassified'
type FilterValue = typeof ALL | typeof UNCLASSIFIED | number
const filterCategoryId = ref<FilterValue>(ALL)

const categories = ref<Category[]>([])
const notes = ref<Note[]>([])
const loading = ref(false)
const error = ref<string>('')
const importDialogRef = ref<InstanceType<typeof ImportDialog> | null>(null)

async function loadCategories() {
  try {
    const result = await listCategories()
    categories.value = result.items
  } catch (e) {
    // 筛选下拉加载失败不阻塞列表——只是筛选用不了
    console.error('load categories:', e)
  }
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const filter = filterCategoryId.value
    if (filter === UNCLASSIFIED) {
      // "未分类"：拉全量 + 本地过滤（后端没暴露"未分类"语义）
      const result = await listNotes()
      notes.value = result.items.filter(n => n.categoryId == null)
    } else if (typeof filter === 'number') {
      // 选某分类：走后端 query
      const result = await listNotesByCategory(filter)
      notes.value = result.items
    } else {
      // 全部
      const result = await listNotes()
      notes.value = result.items
    }
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    ElMessage.error(msg)
  } finally {
    loading.value = false
  }
}

function onFilterChange() {
  // 切筛选时重新加载
  load()
}

async function onDelete(note: Note) {
  try {
    await ElMessageBox.confirm(`确定要删除「${note.title}」吗？`, '删除确认', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }
  try {
    await deleteNote(note.id)
    notes.value = notes.value.filter(n => n.id !== note.id)
    ElMessage.success('已删除')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    ElMessage.error(msg)
  }
}

async function onImportSuccess(_result: ImportResult) {
  await load()
}

function openImport() {
  importDialogRef.value?.open()
}

onMounted(() => {
  loadCategories()
  load()
})
</script>

<template>
  <main v-loading="loading" style="padding: 2rem; max-width: 800px; margin: 0 auto;">
    <header style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem;">
      <h1 style="margin: 0;">我的笔记</h1>
      <div style="display: flex; gap: 1rem; align-items: center;">
        <el-button @click="openImport">导入 .md</el-button>
        <RouterLink to="/notes/new" style="text-decoration: none;">
          <el-link type="primary" :underline="false">+ 新建</el-link>
        </RouterLink>
      </div>
    </header>

    <div style="display: flex; align-items: center; gap: 0.75rem; margin-bottom: 1rem;">
      <span style="color: #606266; font-size: 0.875rem;">分类筛选：</span>
      <el-select
        :model-value="filterCategoryId"
        @update:model-value="(v: FilterValue) => { filterCategoryId = v; onFilterChange() }"
        style="width: 200px;"
      >
        <el-option label="全部分类" :value="ALL" />
        <el-option label="未分类" :value="UNCLASSIFIED" />
        <el-option
          v-for="cat in categories"
          :key="cat.id"
          :label="cat.name"
          :value="cat.id"
        />
      </el-select>
    </div>

    <el-alert v-if="error" :title="error" type="error" :closable="false" style="margin-bottom: 1rem;" />

    <el-empty
      v-if="!loading && !error && notes.length === 0"
      :description="filterCategoryId === UNCLASSIFIED ? '没有未分类的笔记' : '暂无笔记，点击右上角新建或导入'"
    />

    <ul v-if="!loading && !error && notes.length > 0" style="list-style: none; padding: 0; margin: 0;">
      <li
        v-for="note in notes"
        :key="note.id"
        style="display: flex; justify-content: space-between; align-items: center; padding: 0.75rem 0; border-bottom: 1px solid #eee;"
      >
        <RouterLink
          :to="`/notes/${note.id}`"
          style="flex: 1; color: inherit; text-decoration: none; display: flex; align-items: baseline; gap: 0.5rem; flex-wrap: wrap;"
        >
          <el-link type="primary" :underline="true">{{ note.title }}</el-link>
          <el-tag
            v-if="note.categoryName"
            size="small"
            type="info"
            effect="plain"
          >
            {{ note.categoryName }}
          </el-tag>
          <span style="color: gray; font-size: 0.875rem;">
            {{ new Date(note.updatedAt).toLocaleString() }}
          </span>
        </RouterLink>
        <div style="display: flex; gap: 0.5rem;">
          <el-button size="small" @click="$router.push(`/notes/${note.id}/edit`)">编辑</el-button>
          <el-button type="danger" plain size="small" @click="onDelete(note)">删除</el-button>
        </div>
      </li>
    </ul>

    <ImportDialog ref="importDialogRef" @success="onImportSuccess" />
  </main>
</template>
