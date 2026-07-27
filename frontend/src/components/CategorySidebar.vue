<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { listCategories, createCategory } from '../api/categories'
import type { Category } from '../api/types'

const route = useRoute()
const router = useRouter()
const categories = ref<Category[]>([])
const activeCategoryId = ref<number | null>(null)

const showDialog = ref(false)
const newName = ref('')
const creating = ref(false)

async function load() {
  try {
    categories.value = (await listCategories()).items
  } catch (e) {
    console.error('load categories:', e)
  }
}

function isActive(id: number | null): boolean {
  return activeCategoryId.value === id
}

function selectCategory(id: number | null) {
  activeCategoryId.value = id
  const query = { ...route.query }
  if (id === null) {
    delete query.categoryId
  } else {
    query.categoryId = String(id)
  }
  router.push({ query })
}

async function onCreate() {
  const name = newName.value.trim()
  if (!name) {
    ElMessage.error('分类名不能为空')
    return
  }
  creating.value = true
  try {
    const created = await createCategory({ name })
    ElMessage.success(`分类「${created.name}」已创建`)
    newName.value = ''
    showDialog.value = false
    await load()
    selectCategory(created.id)
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : String(e))
  } finally {
    creating.value = false
  }
}

onMounted(() => {
  load()
  const cid = route.query.categoryId
  if (cid && cid !== '0') {
    const parsed = Number(cid)
    if (Number.isFinite(parsed) && parsed > 0) {
      activeCategoryId.value = parsed
    }
  }
})
</script>

<template>
  <div class="sidebar-inner">
    <div class="section">
      <div class="section-title">导航</div>
      <div
        class="menu-item"
        :class="{ active: isActive(null) }"
        @click="selectCategory(null)"
      >
        <span class="menu-icon">📚</span>
        <span>全部分类</span>
      </div>
      <div
        class="menu-item"
        :class="{ active: route.query.categoryId === '0' }"
        @click="selectCategory(0)"
      >
        <span class="menu-icon">📂</span>
        <span>未分类</span>
      </div>
    </div>

    <div class="section">
      <div class="section-title">
        <span>分类</span>
        <button class="add-btn" @click="showDialog = true" title="新建分类">+</button>
      </div>
      <div
        v-for="cat in categories"
        :key="cat.id"
        class="menu-item"
        :class="{ active: isActive(cat.id) }"
        @click="selectCategory(cat.id)"
      >
        <span class="menu-icon">🏷️</span>
        <span class="menu-label">{{ cat.name }}</span>
      </div>
      <div v-if="categories.length === 0" class="empty-hint">
        还没有分类，点击 + 创建
      </div>
    </div>

    <el-dialog v-model="showDialog" title="新建分类" width="380px" append-to-body>
      <el-input
        v-model="newName"
        placeholder="如 Go、Vue、读书笔记"
        maxlength="30"
        show-word-limit
        @keyup.enter="onCreate"
      />
      <template #footer>
        <el-button @click="showDialog = false" :disabled="creating">取消</el-button>
        <el-button type="primary" :loading="creating" @click="onCreate">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.sidebar-inner {
  padding: var(--space-lg) 0;
}

.section {
  margin-bottom: var(--space-xl);
}

.section-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--space-lg);
  margin-bottom: var(--space-sm);
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.add-btn {
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: transparent;
  border: 1px solid var(--color-border-strong);
  color: var(--color-text-secondary);
  font-size: 1rem;
  line-height: 1;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
}
.add-btn:hover {
  background: var(--color-text);
  color: white;
  border-color: var(--color-text);
}

.menu-item {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  padding: 0.5rem var(--space-lg);
  margin: 0 var(--space-sm);
  border-radius: var(--radius-md);
  cursor: pointer;
  font-size: 0.875rem;
  color: var(--color-text);
  transition: all 0.15s;
  position: relative;
}
.menu-item:hover { background: var(--color-bg); }
.menu-item.active {
  background: var(--color-accent-light);
  color: var(--color-accent);
  font-weight: 500;
}
.menu-item.active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 25%;
  bottom: 25%;
  width: 3px;
  background: var(--color-accent);
  border-radius: 0 2px 2px 0;
}

.menu-icon { font-size: 1rem; flex-shrink: 0; }
.menu-label { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.empty-hint {
  padding: var(--space-md) var(--space-lg);
  font-size: 0.8125rem;
  color: var(--color-text-muted);
  font-style: italic;
}
</style>
