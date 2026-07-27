<script setup lang="ts">
import { Editor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import Image from '@tiptap/extension-image'
import Link from '@tiptap/extension-link'
// @tiptap/extension-table 3.29 只导出 `Table` 命名导出（不再有 default）。
// 子包 table-row/table-cell/table-header 提供 default + 命名 re-export，但 3.29 推荐从主包直导。
import { Table } from '@tiptap/extension-table'
import TableRow from '@tiptap/extension-table-row'
import TableCell from '@tiptap/extension-table-cell'
import TableHeader from '@tiptap/extension-table-header'
import CodeBlockLowlight from '@tiptap/extension-code-block-lowlight'
import { Markdown } from 'tiptap-markdown'
import { common, createLowlight } from 'lowlight'
import { onMounted, onBeforeUnmount, ref, watch } from 'vue'

// 代码高亮：common 包含常见语言（js/ts/python/go/sql/bash 等）
const lowlight = createLowlight(common)

const props = defineProps<{
  modelValue: string
  placeholder?: string
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const editor = ref<Editor | null>(null)

onMounted(() => {
  if (!editor.value) {
    editor.value = new Editor({
      extensions: [
        StarterKit.configure({
          codeBlock: false, // 用 CodeBlockLowlight 替换 StarterKit 的 codeBlock
        }),
        // Markdown 双向桥：setContent(md) 解析、storage.markdown.getMarkdown() 序列化
        Markdown.configure({
          html: false, // 富文本不直接接受 HTML 字符串（防 XSS）
          breaks: true, // 输入 Enter 走 Markdown 段落/换行规则
          linkify: true, // 自动识别 URL
          transformPastedText: true, // 粘贴纯文本按 Markdown 解析
        }),
        Image.configure({
          allowBase64: true, // 允许 base64（用户选本地图片后内嵌）
        }),
        Link.configure({
          openOnClick: false, // 编辑时不打开链接
          autolink: true,
          HTMLAttributes: {
            rel: 'noopener noreferrer',
            target: '_blank',
          },
        }),
        Table.configure({ resizable: false }),
        TableRow,
        TableHeader,
        TableCell,
        CodeBlockLowlight.configure({ lowlight }),
      ],
      content: props.modelValue,
      onUpdate: ({ editor }) => {
        // 取出 Markdown 字符串（不是 HTML）→ 父组件 NoteInput.content 字段
        const markdown = (editor.storage as { markdown: { getMarkdown: () => string } }).markdown.getMarkdown()
        emit('update:modelValue', markdown)
      },
    })
  }
})

// 监听外部 modelValue 变化（如加载已有笔记）→ 同步到 editor
// 仅在外部值跟当前 editor 内部 Markdown 不一致时更新，避免循环
watch(
  () => props.modelValue,
  (newValue) => {
    if (!editor.value) return
    const current = (editor.value.storage as { markdown: { getMarkdown: () => string } }).markdown.getMarkdown()
    if (newValue !== current) {
      editor.value.commands.setContent(newValue, { emitUpdate: false })
    }
  },
)

onBeforeUnmount(() => {
  editor.value?.destroy()
})

// 工具栏 actions
function toggleBold() { editor.value?.chain().focus().toggleBold().run() }
function toggleItalic() { editor.value?.chain().focus().toggleItalic().run() }
function toggleStrike() { editor.value?.chain().focus().toggleStrike().run() }
function toggleHeading(level: 1 | 2 | 3) {
  editor.value?.chain().focus().toggleHeading({ level }).run()
}
function toggleBulletList() { editor.value?.chain().focus().toggleBulletList().run() }
function toggleOrderedList() { editor.value?.chain().focus().toggleOrderedList().run() }
function toggleCodeBlock() { editor.value?.chain().focus().toggleCodeBlock().run() }
function setLink() {
  const previousUrl = (editor.value?.getAttributes('link').href as string) || ''
  const url = window.prompt('请输入链接 URL', previousUrl || 'https://')
  if (url === null) return
  if (url === '') {
    editor.value?.chain().focus().extendMarkRange('link').unsetLink().run()
    return
  }
  editor.value?.chain().focus().extendMarkRange('link').setLink({ href: url }).run()
}
function addImage() {
  const url = window.prompt('请输入图片 URL（base64 也支持）', 'https://')
  if (url === null || url === '') return
  editor.value?.chain().focus().setImage({ src: url }).run()
}
function insertTable() {
  editor.value?.chain().focus().insertTable({ rows: 3, cols: 3, withHeaderRow: true }).run()
}
function undo() { editor.value?.chain().focus().undo().run() }
function redo() { editor.value?.chain().focus().redo().run() }

function isActive(name: string, attrs?: Record<string, unknown>): boolean {
  return editor.value?.isActive(name, attrs) ?? false
}
</script>

<template>
  <div class="rich-text-editor">
    <div v-if="editor" class="toolbar">
      <button type="button" :class="{ active: isActive('bold') }" @click="toggleBold" title="粗体">
        <strong>B</strong>
      </button>
      <button type="button" :class="{ active: isActive('italic') }" @click="toggleItalic" title="斜体">
        <em>I</em>
      </button>
      <button type="button" :class="{ active: isActive('strike') }" @click="toggleStrike" title="删除线">
        <s>S</s>
      </button>
      <span class="divider"></span>
      <button type="button" :class="{ active: isActive('heading', { level: 1 }) }" @click="toggleHeading(1)" title="H1">H1</button>
      <button type="button" :class="{ active: isActive('heading', { level: 2 }) }" @click="toggleHeading(2)" title="H2">H2</button>
      <button type="button" :class="{ active: isActive('heading', { level: 3 }) }" @click="toggleHeading(3)" title="H3">H3</button>
      <span class="divider"></span>
      <button type="button" :class="{ active: isActive('bulletList') }" @click="toggleBulletList" title="无序列表">• List</button>
      <button type="button" :class="{ active: isActive('orderedList') }" @click="toggleOrderedList" title="有序列表">1. List</button>
      <span class="divider"></span>
      <button type="button" :class="{ active: isActive('codeBlock') }" @click="toggleCodeBlock" title="代码块">{}</button>
      <button type="button" :class="{ active: isActive('link') }" @click="setLink" title="链接">🔗</button>
      <button type="button" @click="addImage" title="图片">🖼️</button>
      <button type="button" @click="insertTable" title="表格">⊞</button>
      <span class="divider"></span>
      <button type="button" :disabled="!editor.can().undo()" @click="undo" title="撤销">↶</button>
      <button type="button" :disabled="!editor.can().redo()" @click="redo" title="重做">↷</button>
    </div>
    <EditorContent :editor="editor" class="editor-content" />
  </div>
</template>

<style scoped>
.rich-text-editor {
  border: 1px solid var(--color-border, #dcdfe6);
  border-radius: 4px;
  display: flex;
  flex-direction: column;
  background: white;
}

.toolbar {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.5rem;
  border-bottom: 1px solid var(--color-border, #dcdfe6);
  background: var(--color-bg-elevated, #fafafa);
  flex-wrap: wrap;
}

.toolbar button {
  padding: 0.25rem 0.5rem;
  border: 1px solid transparent;
  border-radius: 4px;
  background: transparent;
  cursor: pointer;
  font-size: 0.875rem;
  color: var(--color-text, #303133);
  transition: all 0.15s;
  min-width: 2rem;
}

.toolbar button:hover:not(:disabled) {
  background: white;
  border-color: var(--color-border, #dcdfe6);
}

.toolbar button.active {
  background: var(--color-accent-light, #ecf5ff);
  border-color: var(--color-accent, #409eff);
  color: var(--color-accent, #409eff);
}

.toolbar button:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.toolbar .divider {
  width: 1px;
  height: 1.25rem;
  background: var(--color-border, #dcdfe6);
  margin: 0 0.25rem;
}

.editor-content {
  padding: 0.75rem 1rem;
  min-height: 300px;
  max-height: 600px;
  overflow-y: auto;
  font-size: 1rem;
  line-height: 1.6;
}

.editor-content :deep(.tiptap) {
  outline: none;
}

.editor-content :deep(h1) {
  font-size: 1.5rem;
  font-weight: 600;
  margin: 1rem 0 0.5rem;
}

.editor-content :deep(h2) {
  font-size: 1.25rem;
  font-weight: 600;
  margin: 1rem 0 0.5rem;
}

.editor-content :deep(h3) {
  font-size: 1.125rem;
  font-weight: 600;
  margin: 0.75rem 0 0.5rem;
}

.editor-content :deep(p) {
  margin: 0.5rem 0;
}

.editor-content :deep(ul),
.editor-content :deep(ol) {
  padding-left: 1.5rem;
  margin: 0.5rem 0;
}

.editor-content :deep(code) {
  background: var(--color-bg, #f5f7fa);
  padding: 0.1rem 0.3rem;
  border-radius: 3px;
  font-family: monospace;
  font-size: 0.875rem;
}

.editor-content :deep(pre) {
  background: #1e1e1e;
  color: #d4d4d4;
  padding: 0.75rem 1rem;
  border-radius: 4px;
  overflow-x: auto;
  font-family: monospace;
  font-size: 0.875rem;
}

.editor-content :deep(pre code) {
  background: transparent;
  padding: 0;
  color: inherit;
}

.editor-content :deep(a) {
  color: var(--color-accent, #409eff);
  text-decoration: underline;
}

.editor-content :deep(img) {
  max-width: 100%;
  height: auto;
  border-radius: 4px;
}

.editor-content :deep(table) {
  border-collapse: collapse;
  margin: 0.5rem 0;
  width: 100%;
}

.editor-content :deep(th),
.editor-content :deep(td) {
  border: 1px solid var(--color-border, #dcdfe6);
  padding: 0.5rem;
  min-width: 50px;
  position: relative;
}

.editor-content :deep(th) {
  background: var(--color-bg-elevated, #fafafa);
  font-weight: 600;
}

/* tiptap-markdown 内部的源码编辑容器（一般不可见） */
.editor-content :deep(.code-block-wrap) {
  position: relative;
}
</style>
