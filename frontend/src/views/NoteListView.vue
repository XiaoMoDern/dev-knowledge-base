<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { listNotes, deleteNote } from '../api/notes'
import type { Note } from '../api/types'

const notes = ref<Note[]>([])
const loading = ref(false)
const error = ref<string>('')

async function load() {
  loading.value = true
  error.value = ''
  try {
    const result = await listNotes()
    notes.value = result.items
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

async function onDelete(note: Note) {
  // ElMessageBox.confirm 用户点取消时 reject，catch 一下静默退出
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
    error.value = e instanceof Error ? e.message : String(e)
  }
}

onMounted(load)
</script>

<template>
  <main v-loading="loading" style="padding: 2rem; max-width: 800px; margin: 0 auto;">
    <header style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem;">
      <h1 style="margin: 0;">我的笔记</h1>
      <RouterLink to="/notes/new" style="text-decoration: none;">
        <el-link type="primary" :underline="false">+ 新建</el-link>
      </RouterLink>
    </header>

    <el-alert v-if="error" :title="error" type="error" :closable="false" style="margin-bottom: 1rem;" />

    <el-empty v-if="!loading && !error && notes.length === 0" description="暂无笔记，点击右上角新建" />

    <ul v-if="!loading && !error && notes.length > 0" style="list-style: none; padding: 0; margin: 0;">
      <li
        v-for="note in notes"
        :key="note.id"
        style="display: flex; justify-content: space-between; align-items: center; padding: 0.75rem 0; border-bottom: 1px solid #eee;"
      >
        <RouterLink
          :to="`/notes/${note.id}`"
          style="flex: 1; color: inherit; text-decoration: none; display: flex; align-items: baseline; gap: 0.5rem;"
        >
          <el-link type="primary" :underline="true">{{ note.title }}</el-link>
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
  </main>
</template>
