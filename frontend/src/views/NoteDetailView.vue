<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { listNotes, deleteNote } from '../api/notes'
import type { Note } from '../api/types'

const route = useRoute()
const router = useRouter()
const note = ref<Note | null>(null)
const loading = ref(false)
const notFound = ref(false)
const error = ref<string>('')

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
    error.value = e instanceof Error ? e.message : String(e)
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
      <p style="color: gray; font-size: 0.875rem; margin: 0.5rem 0 1.5rem;">
        最后更新：{{ new Date(note.updatedAt).toLocaleString() }}
      </p>
      <div style="white-space: pre-wrap; line-height: 1.6; margin-bottom: 2rem;">{{ note.content }}</div>
      <div style="display: flex; gap: 0.5rem;">
        <el-button type="primary" @click="$router.push(`/notes/${note.id}/edit`)">编辑</el-button>
        <el-button type="danger" @click="onDelete">删除</el-button>
        <el-button @click="$router.push('/')">返回列表</el-button>
      </div>
    </article>
  </main>
</template>
