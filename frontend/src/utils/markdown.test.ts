import { describe, expect, it } from 'vitest'
import { renderMarkdown } from './markdown'

// renderMarkdown 的安全契约：marked 输出 + DOMPurify 净化后，危险输入必须被完全中和。
// 验收 3 类危险 + 1 类正常。

describe('renderMarkdown — XSS 净化', () => {
  it('移除 <img onerror> 事件属性', () => {
    const html = renderMarkdown('<img src=x onerror=alert(1)>')
    expect(html).not.toContain('onerror')
  })

  it('移除 <script> 标签', () => {
    const html = renderMarkdown('<script>alert(1)</script>')
    expect(html).not.toContain('<script')
  })

  it('移除 javascript: URL', () => {
    const html = renderMarkdown('[click](javascript:alert(1))')
    // javascript: 大小写或混合大小写都该过滤
    expect(html.toLowerCase()).not.toContain('javascript:')
  })

  it('正常 markdown 输入仍正确渲染（净化不该破坏合法内容）', () => {
    const html = renderMarkdown('# Hello\n\n**bold** and `code`')
    // marked v18 gfm=true + breaks=true 下：h1 + strong + code 三件
    expect(html).toContain('<h1')
    expect(html).toContain('Hello')
    expect(html).toContain('<strong')
    expect(html).toContain('<code')
    // 没误伤：危险内容不在
    expect(html).not.toContain('<script')
  })

  it('代码块里的 HTML 也安全（即使被 escaped 进 <pre><code>，原内容也安全）', () => {
    const html = renderMarkdown('```html\n<script>alert(1)</script>\n```')
    // marked 渲染成 <pre><code>...</code></pre>，内容本身没注入到 DOM
    // 但 sanitize 会进一步保险，确保 <pre> 内不会出现活的 <script>
    expect(html.toLowerCase()).not.toContain('<script>alert(1)</script>')
  })
})
