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
  <main v-loading="loading" style="padding: 2rem; max-width: 800px; margin: 0 auto;">
    <el-alert v-if="error" :title="error" type="error" :closable="false" style="margin-bottom: 1rem;" />

    <el-empty v-if="!loading && !error && notFound" description="笔记不存在">
      <el-button type="primary" @click="$router.push('/')">返回列表</el-button>
    </el-empty>

    <article v-else-if="note">
      <h1 style="margin: 0;">{{ note.title }}</h1>
      <p v-if="note.categoryName" style="color: #909399; font-size: 0.875rem; margin: 0.25rem 0 0;">
        分类：{{ note.categoryName }}
      </p>
      <p style="color: gray; font-size: 0.875rem; margin: 0.5rem 0 1.5rem;">
        最后更新：{{ new Date(note.updatedAt).toLocaleString() }}
      </p>
      <div class="markdown-body" v-html="renderedContent"></div>
      <div style="display: flex; gap: 0.5rem;">
        <el-button type="primary" @click="$router.push(`/notes/${note.id}/edit`)">编辑</el-button>
        <el-button type="danger" @click="onDelete">删除</el-button>
        <el-button @click="$router.push('/')">返回列表</el-button>
      </div>
    </article>
  </main>
</template>

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
