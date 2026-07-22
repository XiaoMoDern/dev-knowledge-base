<script setup lang="ts">
import { ref, onMounted, watch, onUnmounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { searchNotes, deleteNote } from '../api/notes'
import { listCategories } from '../api/categories'
import type { Note, Category, ImportResult } from '../api/types'
import ImportDialog from '../components/ImportDialog.vue'

// 分类筛选 3 态：'all'（全部） / 'unclassified'（未分类） / number（具体分类 id）
const ALL = 'all'
const UNCLASSIFIED = 'unclassified'
type FilterValue = typeof ALL | typeof UNCLASSIFIED | number

const route = useRoute()
const router = useRouter()

// state
const searchKeyword = ref<string>('')
const filterCategoryId = ref<FilterValue>(ALL)
const page = ref<number>(1)
const pageSize = ref<number>(20)
const total = ref<number>(0)
const categories = ref<Category[]>([])
const notes = ref<Note[]>([])
const loading = ref(false)
const error = ref<string>('')
const importDialogRef = ref<InstanceType<typeof ImportDialog> | null>(null)

// 初始化标志：onMounted 期间 loadFromURL 改 ref 不应触发搜索
// 否则：loadFromURL() 改 ref → watch 触发 → 300ms 后又 load 一次
//      onMounted 又主动调 load() → 双请求 + 视觉闪烁
let isInitializing = true

// debounce 搜索：300ms 内只发一次请求 + 立即触发（Enter / 搜索按钮）
let searchTimer: number | null = null
function triggerSearch(immediate = false) {
  if (searchTimer !== null) clearTimeout(searchTimer)
  const doSearch = () => {
    page.value = 1 // 切搜索时回到第 1 页
    load()
    syncURL()
  }
  if (immediate) {
    doSearch()
  } else {
    searchTimer = window.setTimeout(doSearch, 300)
  }
}
// 监听 v-model 触发（连续打字 debounce）
// 注意：不要用 @input 配合 v-model——会双重触发
// 注意：初始化期间 loadFromURL 改 ref 不应触发搜索（用 isInitializing 拦截）
watch(searchKeyword, () => {
  if (isInitializing) return
  triggerSearch(false)
})
onUnmounted(() => {
  if (searchTimer !== null) clearTimeout(searchTimer)
})

// 同步 state → URL（router.replace 不进 history，避免后退按钮被筛选状态污染）
function syncURL() {
  const query: Record<string, string> = {}
  if (searchKeyword.value) query.q = searchKeyword.value
  if (filterCategoryId.value === UNCLASSIFIED) {
    query.categoryId = '0' // 0 = 未分类（前端约定，跟后端 category_id NULL 对应）
  } else if (typeof filterCategoryId.value === 'number') {
    query.categoryId = String(filterCategoryId.value)
  }
  if (page.value > 1) query.page = String(page.value)
  router.replace({ query })
}

// 从 URL → state（onMounted 初始化 + 浏览器后退触发）
function loadFromURL() {
  const q = route.query
  searchKeyword.value = typeof q.q === 'string' ? q.q : ''
  const cid = typeof q.categoryId === 'string' ? q.categoryId : ''
  if (cid === '0') {
    filterCategoryId.value = UNCLASSIFIED
  } else if (cid) {
    const parsed = Number(cid)
    if (Number.isFinite(parsed) && parsed > 0) {
      filterCategoryId.value = parsed
    } else {
      filterCategoryId.value = ALL
    }
  } else {
    filterCategoryId.value = ALL
  }
  page.value = typeof q.page === 'string' ? Math.max(1, Number(q.page) || 1) : 1
}

async function loadCategories() {
  try {
    categories.value = (await listCategories()).items
  } catch (e) {
    console.error('load categories:', e)
  }
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    const filter = filterCategoryId.value
    const q = searchKeyword.value || undefined // 三个分支都要带 q（搜+分类 AND 关系）
    let result

    if (filter === UNCLASSIFIED) {
      // "未分类"：拉全量 + 本地过滤（后端没暴露"未分类"语义）
      // q 也传给后端减少本地过滤前的全量
      result = await searchNotes({ q, page: page.value, pageSize: pageSize.value })
      notes.value = result.items.filter(n => n.categoryId == null)
      total.value = notes.value.length
    } else if (typeof filter === 'number') {
      // 选具体分类：搜+分类是 AND 关系——后端 SearchNotes 同时应用
      result = await searchNotes({ q, categoryId: filter, page: page.value, pageSize: pageSize.value })
      notes.value = result.items
      total.value = result.total
    } else {
      // 全部：支持搜索
      result = await searchNotes({ q, page: page.value, pageSize: pageSize.value })
      notes.value = result.items
      total.value = result.total
    }
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    ElMessage.error(msg)
  } finally {
    loading.value = false
  }
}

function onFilterChange(v: FilterValue) {
  filterCategoryId.value = v
  page.value = 1
  load()
  syncURL()
}

function onPageChange(newPage: number) {
  page.value = newPage
  load()
  syncURL()
}

async function onDelete(note: Note) {
  try {
    await ElMessageBox.confirm(`确定要删除「${note.title}」吗？`, '删除确认', {
      type: 'warning', confirmButtonText: '删除', cancelButtonText: '取消',
    })
  } catch { return }
  try {
    await deleteNote(note.id)
    notes.value = notes.value.filter(n => n.id !== note.id)
    total.value = Math.max(0, total.value - 1)
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

// 监听 URL 变化（用户用浏览器后退 / 直接改 URL）
watch(() => route.query, () => { loadFromURL(); load() })

onMounted(async () => {
  loadFromURL()           // 改 ref，被 isInitializing 拦截（不发请求）
  await loadCategories()
  await load()             // 第一次请求
  isInitializing = false  // 初始化完成，后续 watch 才生效
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

    <div style="display: flex; align-items: center; gap: 0.75rem; margin-bottom: 1rem; flex-wrap: wrap;">
      <el-input
        v-model="searchKeyword"
        placeholder="搜索标题或内容..."
        clearable
        style="width: 280px;"
        @keyup.enter="triggerSearch(true)"
      />
      <el-button type="primary" @click="triggerSearch(true)">搜索</el-button>
      <span style="color: #606266; font-size: 0.875rem;">分类：</span>
      <el-select
        :model-value="filterCategoryId"
        @update:model-value="(v: FilterValue) => onFilterChange(v)"
        style="width: 200px;"
      >
        <el-option label="全部分类" :value="ALL" />
        <el-option label="未分类" :value="UNCLASSIFIED" />
        <el-option v-for="cat in categories" :key="cat.id" :label="cat.name" :value="cat.id" />
      </el-select>
    </div>

    <el-alert v-if="error" :title="error" type="error" :closable="false" style="margin-bottom: 1rem;" />

    <el-empty
      v-if="!loading && !error && notes.length === 0"
      :description="
        searchKeyword ? `没有匹配「${searchKeyword}」的笔记`
          : filterCategoryId === UNCLASSIFIED ? '没有未分类的笔记'
          : '暂无笔记，点击右上角新建或导入'"
    />

    <ul v-if="!loading && !error && notes.length > 0" style="list-style: none; padding: 0; margin: 0;">
      <li v-for="note in notes" :key="note.id" style="display: flex; justify-content: space-between; align-items: center; padding: 0.75rem 0; border-bottom: 1px solid #eee;">
        <RouterLink :to="`/notes/${note.id}`" style="flex: 1; color: inherit; text-decoration: none; display: flex; align-items: baseline; gap: 0.5rem; flex-wrap: wrap;">
          <el-link type="primary" :underline="true">{{ note.title }}</el-link>
          <el-tag v-if="note.categoryName" size="small" type="info" effect="plain">{{ note.categoryName }}</el-tag>
          <span style="color: gray; font-size: 0.875rem;">{{ new Date(note.updatedAt).toLocaleString() }}</span>
        </RouterLink>
        <div style="display: flex; gap: 0.5rem;">
          <el-button size="small" @click="$router.push(`/notes/${note.id}/edit`)">编辑</el-button>
          <el-button type="danger" plain size="small" @click="onDelete(note)">删除</el-button>
        </div>
      </li>
    </ul>

    <!-- 分页器：仅在 total > 0 时显示（"未分类" 本地过滤 total 跟 items 一致，也可显示） -->
    <el-pagination
      v-if="total > 0"
      v-model:current-page="page"
      :page-size="pageSize"
      :total="total"
      :page-sizes="[10, 20, 50, 100]"
      layout="total, sizes, prev, pager, next, jumper"
      style="margin-top: 1.5rem; justify-content: center;"
      @current-change="onPageChange"
      @size-change="(s: number) => { pageSize = s; page = 1; load(); syncURL() }"
    />

    <ImportDialog ref="importDialogRef" @success="onImportSuccess" />
  </main>
</template>
