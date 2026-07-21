<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { listNotes, createNote, updateNote } from '../api/notes'
import type { Note } from '../api/types'

const route = useRoute()
const router = useRouter()
const isEdit = Boolean(route.params.id)

const title = ref<string>('')
const content = ref<string>('')

const loading = ref(false)
const saving = ref(false)
const error = ref<string>('')
const notFound = ref(false)

// 编辑模式从 listNotes 找 id 复用的笔记；找不到走 notFound
// 不引入后端 getNote 接口，简单项目不值得多一次往返
async function loadNote() {
  if (!isEdit) return
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
    title.value = found.title
    content.value = found.content
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

async function onSubmit() {
  if (!title.value.trim()) {
    ElMessage.error('title 不能为空')
    return
  }
  saving.value = true
  error.value = ''
  try {
    const input = { title: title.value, content: content.value }
    if (isEdit) {
      await updateNote(Number(route.params.id), input)
    } else {
      await createNote(input)
    }
    ElMessage.success(isEdit ? '已保存' : '已创建')
    router.push('/')
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    saving.value = false
  }
}

onMounted(loadNote)
</script>

<template>
  <main v-loading="loading" style="padding: 2rem; max-width: 600px; margin: 0 auto;">
    <el-alert v-if="error" :title="error" type="error" :closable="false" style="margin-bottom: 1rem;" />

    <el-empty v-if="!loading && !error && notFound" description="笔记不存在">
      <el-button type="primary" @click="$router.push('/')">返回列表</el-button>
    </el-empty>

    <template v-else-if="!loading && !error">
      <h1 style="margin: 0 0 1.5rem;">{{ isEdit ? '编辑笔记' : '新建笔记' }}</h1>

      <el-form @submit.prevent="onSubmit" label-position="top">
        <el-form-item label="标题">
          <el-input v-model="title" placeholder="给笔记起个标题" :disabled="saving" maxlength="200" show-word-limit />
        </el-form-item>

        <el-form-item label="内容">
          <el-input v-model="content" type="textarea" :rows="12" placeholder="写点什么..." :disabled="saving" />
        </el-form-item>

        <div style="display: flex; gap: 0.5rem;">
          <el-button type="primary" :loading="saving" native-type="submit">
            {{ isEdit ? '保存' : '创建' }}
          </el-button>
          <el-button @click="$router.push('/')">取消</el-button>
        </div>
      </el-form>
    </template>
  </main>
</template>
