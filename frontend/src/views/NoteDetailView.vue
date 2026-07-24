<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { listNotes, deleteNote } from '../api/notes'
import type { Note } from '../api/types'
import { renderMarkdown } from '../utils/markdown'

const route = useRoute()
const router = useRouter()
const note = ref<Note | null>(null)
const loading = ref(false)
const notFound = ref(false)
const error = ref<string>('')

const renderedContent = computed(() =>
  note.value ? renderMarkdown(note.value.content) : ''
)

async function load() {
  loading.value = true
  error.value = ''
  notFound.value = false
  try {
    const result = await listNotes()
    const found = result.items.find((n: Note) => n.id === Number(route.params.id))
    if (!found) {
      notFound.value = true
      return
    }
    note.value = found
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

async function onDelete() {
  if (!note.value) return
  try {
    await ElMessageBox.confirm(`确定要删除「${note.value.title}」吗？`, '删除确认', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }
  try {
    await deleteNote(note.value.id)
    ElMessage.success('已删除')
    router.push('/')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    ElMessage.error(msg)
  }
}

onMounted(load)
</script>

<template>
  <main v-loading="loading" class="detail-view">
    <el-alert v-if="error" :title="error" type="error" :closable="false" style="margin-bottom: var(--space-lg);" />

    <el-empty v-if="!loading && !error && notFound" description="笔记不存在">
      <el-button type="primary" @click="$router.push('/')">返回列表</el-button>
    </el-empty>

    <article v-else-if="note">
      <div class="action-bar">
        <div class="action-bar-top">
          <h1 class="article-title">{{ note.title }}</h1>
          <div class="action-buttons">
            <el-button @click="$router.push('/')">返回列表</el-button>
            <el-button type="danger" plain @click="onDelete">删除</el-button>
            <el-button type="primary" @click="$router.push(`/notes/${note.id}/edit`)">编辑</el-button>
          </div>
        </div>
        <div class="action-bar-meta">
          <span v-if="note.categoryName">分类：{{ note.categoryName }}</span>
          <span>最后更新：{{ new Date(note.updatedAt).toLocaleString() }}</span>
        </div>
      </div>
      <div class="markdown-body" v-html="renderedContent"></div>
    </article>
  </main>
</template>

<style scoped>
.detail-view {
  padding: var(--space-xl);
  max-width: 900px;
  margin: 0 auto;
}

.action-bar {
  position: sticky;
  top: calc(var(--header-height) + var(--space-md0));
  background: var(--color-bg-elevated);
  z-index: 5;
  padding: var(--space-lg) var(--space-xl);
  margin: calc(-1 * var(--space-xl)) calc(-1 * var(--space-xl)) var(--space-xl);
  border-radius: var(--radius-lg);
  box-shadow: var(--shadow-md);
  border: 1px solid var(--color-border);
}

.action-bar-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-md);
}

.action-bar-meta {
  display: flex;
  gap: var(--space-lg);
  margin-top: var(--space-sm);
  padding-top: var(--space-sm);
  border-top: 1px solid var(--color-border);
  font-size: 0.8125rem;
  color: var(--color-text-muted);
}

.article-title {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--color-text);
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.action-buttons {
  display: flex;
  gap: var(--space-sm);
  flex-shrink: 0;
}
</style>

<style>
/* GitHub README 风格的 markdown 样式
   非 scoped：因为 v-html 注入的 DOM 不带 scope attribute，
   写 :deep() 反而麻烦；selector 全是 .markdown-body 开头不会污染其他元素 */
.markdown-body { line-height: 1.7; color: #24292e; word-wrap: break-word; }
.markdown-body h1 { font-size: 1.8rem; border-bottom: 1px solid #eaecef; padding-bottom: 0.3rem; margin-top: 1.5rem; }
.markdown-body h2 { font-size: 1.4rem; border-bottom: 1px solid #eaecef; padding-bottom: 0.3rem; margin-top: 1.5rem; }
.markdown-body h3 { font-size: 1.2rem; margin-top: 1.5rem; }
.markdown-body p { margin: 0.8rem 0; }
.markdown-body code {
  background: #f6f8fa; padding: 0.2em 0.4em; border-radius: 3px;
  font-size: 0.9em; font-family: ui-monospace, SFMono-Regular, "SF Mono", Consolas, monospace;
}
.markdown-body pre {
  background: #f6f8fa; padding: 1rem; border-radius: 6px; overflow-x: auto;
  line-height: 1.5;
}
.markdown-body pre code { background: transparent; padding: 0; font-size: 0.9em; }
.markdown-body blockquote {
  border-left: 4px solid #dfe2e5; color: #6a737d; padding: 0 1rem; margin: 1rem 0;
}
.markdown-body ul, .markdown-body ol { padding-left: 2rem; }
.markdown-body a { color: #0366d6; text-decoration: none; }
.markdown-body a:hover { text-decoration: underline; }
.markdown-body table { border-collapse: collapse; margin: 1rem 0; }
.markdown-body table th, .markdown-body table td { border: 1px solid #dfe2e5; padding: 0.4rem 0.8rem; }
.markdown-body hr { border: 0; border-top: 1px solid #eaecef; }
</style>
