<script setup lang="ts">
// 路由 + 列表/创建/更新 API
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { listNotes, createNote, updateNote } from '../api/notes'
import type { Note } from '../api/types'

// 路由：编辑模式 /notes/:id，新建模式 /notes/new
// isEdit 用 params.id 是否有值来区分；不要靠 URL 字符串匹配（不健壮）
const route = useRoute()
const router = useRouter()
const isEdit = Boolean(route.params.id)

// 表单字段：用 ref + v-model 双向绑定
const title = ref<string>('')
const content = ref<string>('')

// UI 状态：四个独立 ref，分别管不同维度
// - loading: 编辑模式初次加载笔记
// - saving: 提交中（按钮 disable + 文字改 "保存中..."）
// - error: 加载或提交错误
// - notFound: 编辑模式下找不到对应 id 的笔记
const loading = ref(false)
const saving = ref(false)
const error = ref<string>('')
const notFound = ref(false)

/**
 * 编辑模式：从 listNotes 找 id 对应的笔记，找不到就显示 notFound 页面。
 * 设计上不引入后端 getNote 接口（避免多一次往返 + 简单项目不值得），
 * 直接复用 listNotes 的结果找。列表大时不划算，那时再开设计文档加 getNote。
 */
async function loadNote() {
  if (!isEdit) return  // 新建模式不用加载
  loading.value = true
  error.value = ''
  notFound.value = false
  try {
    const result = await listNotes()
    // Number(route.params.id) 把 string|number 收窄成 number
    // 不存在 → find 返回 undefined → notFound = true
    const found = result.items.find((n: Note) => n.id === Number(route.params.id))
    if (!found) {
      notFound.value = true
      return
    }
    // 回填表单：编辑模式用户改完才能保存
    title.value = found.title
    content.value = found.content
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
}

/**
 * 提交：编辑模式 updateNote，新建模式 createNote。
 * 前端先 trim 校验 title 非空（后端也会再校验一次，UI 兜底 = 少一次往返）。
 */
async function onSubmit() {
  if (!title.value.trim()) {
    error.value = 'title 不能为空'
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
    // 提交成功 → 跳回列表（用 push 让浏览器能后退回编辑）
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
  <main style="padding: 2rem;">
    <!-- 状态 1：编辑模式初次加载 -->
    <p v-if="loading" style="color: gray;">加载中...</p>

    <!-- 状态 2：编辑模式找不到对应 id 的笔记（你删了 id=5 又访问 /notes/5 就是这） -->
    <template v-else-if="notFound">
      <h1>笔记不存在</h1>
      <p>id 为 <code>{{ route.params.id }}</code> 的笔记已被删除或从未存在。</p>
      <p>
        <RouterLink to="/" style="color: #0066cc;">← 返回列表</RouterLink>
      </p>
    </template>

    <!-- 状态 3：正常表单（新建或编辑） -->
    <template v-else>
      <h1>{{ isEdit ? '编辑笔记' : '新建笔记' }}</h1>

      <form
        @submit.prevent="onSubmit"
        style="display: flex; flex-direction: column; gap: 1rem; max-width: 600px;"
      >
        <label style="display: flex; flex-direction: column; gap: 0.25rem;">
          <span>标题</span>
          <input
            v-model="title"
            type="text"
            :disabled="saving"
            style="padding: 0.5rem; border: 1px solid #ccc; border-radius: 4px;"
          />
        </label>

        <label style="display: flex; flex-direction: column; gap: 0.25rem;">
          <span>内容</span>
          <textarea
            v-model="content"
            rows="10"
            :disabled="saving"
            style="padding: 0.5rem; border: 1px solid #ccc; border-radius: 4px; font-family: inherit;"
          />
        </label>

        <!-- 错误信息：放在按钮上方而不是 alert()，避免阻塞 UI -->
        <p v-if="error" style="color: red; margin: 0;">{{ error }}</p>

        <div style="display: flex; gap: 0.5rem;">
          <button
            type="submit"
            :disabled="saving"
            style="padding: 0.5rem 1rem; background: #0066cc; color: white; border: none; border-radius: 4px;"
          >
            {{ saving ? '保存中...' : (isEdit ? '保存' : '创建') }}
          </button>
          <RouterLink
            to="/"
            style="padding: 0.5rem 1rem; border: 1px solid #ccc; border-radius: 4px;"
          >
            取消
          </RouterLink>
        </div>
      </form>
    </template>
  </main>
</template>
