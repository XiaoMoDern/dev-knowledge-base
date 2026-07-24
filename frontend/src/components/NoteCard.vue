<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { deleteNote } from '../api/notes'
import type { Note } from '../api/types'

const props = defineProps<{
  note: Note
}>()

const emit = defineEmits<{
  (e: 'deleted', id: number): void
}>()

const router = useRouter()

const excerpt = computed(() => {
  const content = props.note.content || ''
  const plain = content.replace(/[#*`>\-\[\]()]/g, '').replace(/\s+/g, ' ').trim()
  return plain.length > 100 ? plain.slice(0, 100) + '...' : plain
})

const updatedAt = computed(() => {
  const date = new Date(props.note.updatedAt)
  return date.toLocaleDateString('zh-CN', { year: 'numeric', month: 'short', day: 'numeric' })
})

function open() {
  router.push(`/notes/${props.note.id}`)
}

async function onDelete(e: Event) {
  // 阻止冒泡到卡片的 @click（避免跳详情）
  e.stopPropagation()
  e.preventDefault()
  try {
    await ElMessageBox.confirm(`确定要删除「${props.note.title}」吗？`, '删除确认', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消',
    })
  } catch {
    return
  }
  try {
    await deleteNote(props.note.id)
    ElMessage.success('已删除')
    emit('deleted', props.note.id)
  } catch (err) {
    ElMessage.error(err instanceof Error ? err.message : String(err))
  }
}
</script>

<template>
  <article class="note-card" @click="open">
    <button class="card-delete-btn" @click="onDelete" title="删除笔记">✕</button>
    <header class="card-header">
      <h3 class="card-title">{{ note.title }}</h3>
      <span v-if="note.categoryName" class="card-tag">{{ note.categoryName }}</span>
    </header>
    <p v-if="excerpt" class="card-excerpt">{{ excerpt }}</p>
    <footer class="card-footer">
      <time class="card-time">{{ updatedAt }}</time>
    </footer>
  </article>
</template>

<style scoped>
.note-card {
  position: relative;
  background: var(--color-bg-elevated);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--space-lg);
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
  min-height: 140px;
}
.note-card:hover {
  border-color: var(--color-border-strong);
  box-shadow: var(--shadow-md);
  transform: translateY(-2px);
}

.card-delete-btn {
  position: absolute;
  top: 0.5rem;
  right: 0.5rem;
  width: 28px;
  height: 28px;
  border-radius: var(--radius-sm);
  border: none;
  background: transparent;
  color: var(--color-text-muted);
  cursor: pointer;
  opacity: 0;
  transition: all 0.15s;
  font-size: 0.875rem;
  line-height: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1;
}
.note-card:hover .card-delete-btn { opacity: 1; }
.card-delete-btn:hover {
  background: var(--color-danger);
  color: white;
}
.card-delete-btn:focus-visible {
  opacity: 1;
  outline: 2px solid var(--color-accent);
  outline-offset: 2px;
}

.card-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-sm);
  /* 给右上角删除按钮留位置 */
  padding-right: 32px;
}

.card-title {
  margin: 0;
  font-size: 1rem;
  font-weight: 600;
  color: var(--color-text);
  line-height: 1.4;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  flex: 1;
}

.card-tag {
  display: inline-block;
  padding: 0.125rem 0.5rem;
  background: var(--color-accent-light);
  color: var(--color-accent);
  font-size: 0.75rem;
  font-weight: 500;
  border-radius: var(--radius-sm);
  flex-shrink: 0;
  white-space: nowrap;
}

.card-excerpt {
  margin: 0;
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  line-height: 1.6;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  flex: 1;
}

.card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: auto;
}

.card-time {
  font-size: 0.75rem;
  color: var(--color-text-muted);
}
</style>
