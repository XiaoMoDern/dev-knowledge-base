<script setup lang="ts">
import { ref, onMounted, watch, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { searchNotes } from '../api/notes'
import type { Note, ImportResult } from '../api/types'
import ImportDialog from '../components/ImportDialog.vue'
import NoteCard from '../components/NoteCard.vue'

const route = useRoute()
const router = useRouter()

const searchKeyword = ref<string>('')
const page = ref<number>(1)
const pageSize = ref<number>(12)
const total = ref<number>(0)
const notes = ref<Note[]>([])
const loading = ref(false)
const error = ref<string>('')
const importDialogRef = ref<InstanceType<typeof ImportDialog> | null>(null)

let isInitializing = true

let searchTimer: number | null = null
function triggerSearch(immediate = false) {
  if (searchTimer !== null) clearTimeout(searchTimer)
  const doSearch = () => {
    page.value = 1
    load()
    syncURL()
  }
  if (immediate) doSearch()
  else searchTimer = window.setTimeout(doSearch, 300)
}

watch(searchKeyword, () => {
  if (isInitializing) return
  triggerSearch(false)
})

function syncURL() {
  const query: Record<string, string> = {}
  if (searchKeyword.value) query.q = searchKeyword.value
  if (page.value > 1) query.page = String(page.value)
  router.replace({ query })
}

function loadFromURL() {
  const q = route.query
  searchKeyword.value = typeof q.q === 'string' ? q.q : ''
  page.value = typeof q.page === 'string' ? Math.max(1, Number(q.page) || 1) : 1
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const cid = route.query.categoryId
    let categoryId: number | undefined
    if (typeof cid === 'string' && cid) {
      const parsed = Number(cid)
      if (Number.isFinite(parsed) && parsed >= 0) categoryId = parsed
    }
    const q = searchKeyword.value || undefined
    const result = await searchNotes({ q, categoryId, page: page.value, pageSize: pageSize.value })
    notes.value = result.items
    total.value = result.total
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    ElMessage.error(msg)
  } finally {
    loading.value = false
  }
}

function onPageChange(newPage: number) {
  page.value = newPage
  load()
  syncURL()
}

async function onImportSuccess(_result: ImportResult) {
  await load()
}

async function onCardDeleted(id: number) {
  // 1. 立即从 notes 数组移除（视觉反馈快，不等 reload）
  notes.value = notes.value.filter(n => n.id !== id)
  total.value = Math.max(0, total.value - 1)

  // 2. 重新拉当前页，让后端补位（删了首页某条 → 补下一页第一条）
  await load()

  // 3. 如果当前页空了 + 不是第 1 页 → 跳到前一页（避免空页 + total 跟 items 不一致）
  if (notes.value.length === 0 && page.value > 1) {
    page.value = page.value - 1
    syncURL()
    await load()
  }
}

function openImport() { importDialogRef.value?.open() }
function openNewNote() { router.push('/notes/new') }

// 监听 URL 变化（包括点侧边栏切分类）→ 重新加载
watch(() => route.query, () => {
  if (isInitializing) return
  loadFromURL()
  load()
})

onUnmounted(() => {
  if (searchTimer !== null) clearTimeout(searchTimer)
})

onMounted(async () => {
  loadFromURL()
  await load()
  isInitializing = false
})
</script>

<template>
  <div class="note-list">
    <div class="list-header">
      <h1 class="list-title">我的笔记</h1>
      <div class="list-actions">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索标题或内容..."
          clearable
          style="width: 280px;"
          @keyup.enter="triggerSearch(true)"
        >
          <template #prefix>🔍</template>
        </el-input>
        <el-button type="primary" @click="triggerSearch(true)">搜索</el-button>
        <el-button @click="openImport">导入 .md</el-button>
        <el-button type="primary" @click="openNewNote">+ 新建</el-button>
      </div>
    </div>

    <el-alert v-if="error" :title="error" type="error" :closable="false" style="margin-bottom: var(--space-lg);" />

    <div v-loading="loading" class="card-grid-wrapper">
      <el-empty
        v-if="!loading && notes.length === 0"
        :description="searchKeyword ? `没有匹配「${searchKeyword}」的笔记` : '暂无笔记，点击右上角新建或导入'"
      />
      <div v-else class="card-grid">
        <NoteCard v-for="note in notes" :key="note.id" :note="note" @deleted="onCardDeleted" />
      </div>
    </div>

    <el-pagination
      v-if="total > 0"
      v-model:current-page="page"
      :page-size="pageSize"
      :total="total"
      :page-sizes="[12, 24, 48]"
      layout="total, sizes, prev, pager, next, jumper"
      class="pagination"
      @current-change="onPageChange"
      @size-change="(s: number) => { pageSize = s; page = 1; load() }"
    />

    <ImportDialog ref="importDialogRef" @success="onImportSuccess" />
  </div>
</template>

<style scoped>
.note-list {
  padding: var(--space-xl);
  max-width: 1400px;
  margin: 0 auto;
}

.list-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--space-xl);
  gap: var(--space-md);
  flex-wrap: wrap;
}

.list-title {
  margin: 0;
  font-size: 1.5rem;
  font-weight: 600;
}

.list-actions {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  flex-wrap: wrap;
}

.card-grid-wrapper { min-height: 300px; }

.card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: var(--space-lg);
}

.pagination {
  margin-top: var(--space-2xl);
  justify-content: center;
}
</style>
