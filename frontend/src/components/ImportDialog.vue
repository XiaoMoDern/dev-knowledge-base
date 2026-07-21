<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElNotification } from 'element-plus'
import type { UploadUserFile } from 'element-plus'
import { importNotes } from '../api/notes'
import { parseMarkdownFile, type ParsedMarkdown } from '../utils/markdown'
import type { ImportResult } from '../api/types'

const emit = defineEmits<{
  success: [ImportResult]
}>()

const visible = ref(false)
const fileList = ref<UploadUserFile[]>([])
const parsed = ref<ParsedMarkdown[]>([])
const submitting = ref(false)

function open() {
  visible.value = true
}
defineExpose({ open })

function reset() {
  fileList.value = []
  parsed.value = []
  submitting.value = false
}

async function onParse() {
  if (fileList.value.length === 0) {
    ElMessage.warning('请先选择 .md 文件')
    return
  }
  const results: ParsedMarkdown[] = []
  for (const f of fileList.value) {
    const raw = f.raw
    if (!raw) {
      results.push({ fileName: f.name, input: null, error: '无法读取文件' })
      continue
    }
    try {
      const text = await raw.text()
      results.push(parseMarkdownFile(f.name, text))
    } catch (e) {
      results.push({
        fileName: f.name,
        input: null,
        error: e instanceof Error ? e.message : String(e),
      })
    }
  }
  parsed.value = results
}

function removeParsed(index: number) {
  parsed.value.splice(index, 1)
}

function onSubmit() {
  const valid = parsed.value.filter(p => p.input !== null)
  if (valid.length === 0) {
    ElMessage.warning('没有可导入的合法笔记')
    return
  }
  submitting.value = true
  importNotes(valid.map(p => p.input!))
    .then(result => {
      if (result.failed === 0) {
        ElMessage.success(`导入成功：${result.imported} 条`)
        visible.value = false
        reset()
        emit('success', result)
      } else if (result.imported === 0) {
        ElMessage.error(`全部失败：${result.failed} 条`)
      } else {
        showPartialResult(result)
        visible.value = false
        reset()
        emit('success', result)
      }
    })
    .catch(e => {
      ElMessage.error(e instanceof Error ? e.message : String(e))
    })
    .finally(() => {
      submitting.value = false
    })
}

// 部分成功用 ElNotification：非阻塞、可以展示多行错误明细
function showPartialResult(result: ImportResult) {
  const head = result.errors.slice(0, 5)
  const more = result.errors.length > 5 ? `\n... 还有 ${result.errors.length - 5} 条` : ''
  const lines = head.map(e => `  · #${e.index}「${e.title || '(空)'}」: ${e.reason}`).join('\n')
  ElNotification({
    title: '部分导入成功',
    message: `成功 ${result.imported} 条 / 失败 ${result.failed} 条\n${lines}${more}`,
    type: 'warning',
    duration: 8000,
  })
}
</script>

<template>
  <el-dialog v-model="visible" title="导入笔记" width="640px" @close="reset">
    <el-upload
      v-model:file-list="fileList"
      :auto-upload="false"
      :multiple="true"
      accept=".md"
      :limit="50"
    >
      <el-button>选择 .md 文件</el-button>
      <template #tip>
        <div style="color: gray; font-size: 0.875rem; margin-top: 0.5rem;">
          支持 front matter（--- 块里的 title 字段）；最多 50 个文件
        </div>
      </template>
    </el-upload>

    <el-table v-if="parsed.length > 0" :data="parsed" style="margin-top: 1rem;" max-height="240">
      <el-table-column prop="fileName" label="文件" width="180" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag v-if="row.input" type="success" size="small">就绪</el-tag>
          <el-tag v-else type="danger" size="small" :title="row.error">
            {{ row.error || '错误' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="标题">
        <template #default="{ row }">
          {{ row.input?.title || '—' }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="80" align="center">
        <template #default="{ $index }">
          <el-button size="small" link type="danger" @click="removeParsed($index)">移除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <template #footer>
      <el-button @click="onParse" :disabled="fileList.length === 0">解析</el-button>
      <el-button
        type="primary"
        :loading="submitting"
        :disabled="parsed.filter(p => p.input).length === 0"
        @click="onSubmit"
      >
        开始导入（{{ parsed.filter(p => p.input).length }}）
      </el-button>
    </template>
  </el-dialog>
</template>
