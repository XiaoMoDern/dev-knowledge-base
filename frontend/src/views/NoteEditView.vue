<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getNote, createNote, updateNote } from '../api/notes'
import { listCategories, createCategory } from '../api/categories'
import type { Category } from '../api/types'

const route = useRoute()
const router = useRouter()
const isEdit = Boolean(route.params.id)

const title = ref<string>('')
const content = ref<string>('')
const categoryId = ref<number | null>(null)

const categories = ref<Category[]>([])
const loading = ref(false)
const saving = ref(false)
const error = ref<string>('')

const notFound = ref(false)

// 弹窗：新建分类
const showNewCategory = ref(false)
const newCategoryName = ref('')
const creatingCategory = ref(false)

async function loadCategories() {
  try {
    const result = await listCategories()
    categories.value = result.items
  } catch (e) {
    // 分类加载失败不阻塞编辑——只是没分类可选
    console.error('load categories:', e)
  }
}

// 编辑模式按 ID 精确查单条 note（替代之前的 listNotes() + 内存 find）
async function loadNote() {
  if (!isEdit) return
  loading.value = true
  error.value = ''
  notFound.value = false
  try {
    const found = await getNote(Number(route.params.id))
    title.value = found.title
    content.value = found.content
    categoryId.value = found.categoryId ?? null
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    if (msg.includes('404')) {
      notFound.value = true
    } else {
      error.value = msg
    }
  } finally {
    loading.value = false
  }
}

async function onCreateCategory() {
  const name = newCategoryName.value.trim()
  if (!name) {
    ElMessage.error('分类名不能为空')
    return
  }
  creatingCategory.value = true
  try {
    const created = await createCategory({ name })
    ElMessage.success(`分类「${created.name}」已创建`)
    // 重新加载分类列表 + 自动选新分类
    await loadCategories()
    categoryId.value = created.id
    // 关闭弹窗 + 清空输入
    newCategoryName.value = ''
    showNewCategory.value = false
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : String(e))
  } finally {
    creatingCategory.value = false
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
    const input: { title: string; content: string; categoryId: number | null } = {
      title: title.value,
      content: content.value,
      categoryId: categoryId.value,
    }
    if (isEdit) {
      await updateNote(Number(route.params.id), input)
    } else {
      await createNote(input)
    }
    ElMessage.success(isEdit ? '已保存' : '已创建')
    router.push('/')
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    error.value = msg
    ElMessage.error(msg)
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadCategories()
  loadNote()
})
</script>

<style scoped>
.edit-view {
  padding: var(--space-xl);
  max-width: 800px;
  margin: 0 auto;
}
</style>

<template>
  <main v-loading="loading" class="edit-view">
    <el-alert v-if="error" :title="error" type="error" :closable="false" style="margin-bottom: var(--space-lg);" />

    <el-empty v-if="!loading && !error && notFound" description="笔记不存在">
      <el-button type="primary" @click="$router.push('/')">返回列表</el-button>
    </el-empty>

    <template v-else-if="!loading && !error">
      <h1 style="margin: 0 0 1.5rem;">{{ isEdit ? '编辑笔记' : '新建笔记' }}</h1>

      <el-form @submit.prevent="onSubmit" label-position="top">
        <el-form-item label="标题">
          <el-input v-model="title" placeholder="给笔记起个标题" :disabled="saving" maxlength="200" show-word-limit />
        </el-form-item>

        <el-form-item label="分类">
          <div style="display: flex; gap: 0.5rem; align-items: center;width: 100%;">
            <el-select
              v-model="categoryId"
              placeholder="选择分类（可选）"
              clearable
              :disabled="saving"
              style="width: 100%; flex: 1;"
            >
              <el-option
                v-for="cat in categories"
                :key="cat.id"
                :label="cat.name"
                :value="cat.id"
              />
            </el-select>
            <el-button :disabled="saving" @click="showNewCategory = true">+ 新建</el-button>
          </div>
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

    <el-dialog v-model="showNewCategory" title="新建分类" width="400px">
      <el-form @submit.prevent="onCreateCategory">
        <el-form-item label="分类名">
          <el-input
            v-model="newCategoryName"
            placeholder="如 Go、Vue、读书笔记"
            :disabled="creatingCategory"
            maxlength="30"
            show-word-limit
            ref="newCategoryInput"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showNewCategory = false" :disabled="creatingCategory">取消</el-button>
        <el-button type="primary" :loading="creatingCategory" @click="onCreateCategory">创建</el-button>
      </template>
    </el-dialog>
  </main>
</template>
