import type { ImportNoteInput } from '../api/types'
import { marked } from 'marked'

// 单个文件解析结果：input=null 表示解析失败，error 字段说明原因
export interface ParsedMarkdown {
  fileName: string
  input: ImportNoteInput | null
  error: string
}

// 极简 front matter 解析：只识别 `key: value` 单行，不依赖 js-yaml
// 规则：必须从文件首行开始，--- 单独成行；只支持字符串值，数组/嵌套对象不支持
function parseFrontMatter(text: string): { meta: Record<string, string>; body: string } {
  const match = text.match(/^---\r?\n([\s\S]*?)\r?\n---\r?\n?([\s\S]*)$/)
  if (!match) {
    return { meta: {}, body: text }
  }
  const [, yaml, body] = match
  const meta: Record<string, string> = {}
  for (const line of yaml.split(/\r?\n/)) {
    const trimmed = line.trim()
    if (!trimmed || trimmed.startsWith('#')) continue
    const colon = trimmed.indexOf(':')
    if (colon < 0) continue
    const key = trimmed.slice(0, colon).trim()
    // 去首尾引号（单/双），支持 title: "xxx" 或 title: 'xxx'
    const value = trimmed.slice(colon + 1).trim().replace(/^["']|["']$/g, '')
    meta[key] = value
  }
  return { meta, body }
}

// 文件名 fallback：去扩展名 + 把 -_ 替成空格
// "go-pointers.md" -> "go pointers"
function fileNameToTitle(name: string): string {
  return name.replace(/\.md$/i, '').replace(/[-_]+/g, ' ').trim()
}

// 把 .md 文件名 + 文本转成 ImportNoteInput
// title 优先级：front matter 的 title > 文件名 fallback
export function parseMarkdownFile(fileName: string, text: string): ParsedMarkdown {
  const { meta, body } = parseFrontMatter(text)
  const title = (meta.title ?? '').trim() || fileNameToTitle(fileName)
  return { fileName, input: { title, content: body.trim() }, error: '' }
}

// marked 配置：GFM（表格/删除线/任务列表）+ 单换行变 <br>（更接近日常编辑习惯）
// 用 module-level 配置，marked 实例单例，整个应用共享
marked.use({ gfm: true, breaks: true })

/**
 * markdown 文本 → HTML 字符串，给 v-html 用
 * 安全性：marked 默认 HTML escape（用户写的 <script> 不会被执行）
 * 信任源假设：单机本地应用，content 来自用户自己导入的 .md 文件，不加 DOMPurify
 */
export function renderMarkdown(text: string): string {
  // marked v18 parse 默认返回 Promise<string | undefined>，强制同步需要 async:false
  return marked.parse(text, { async: false }) as string
}
