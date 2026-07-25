---
id: vue-el-message
title: "Element Plus ElMessage：4 种消息 + showClose + 实战模式"
category: Vue
tags:
  - Vue
  - Element Plus
  - ElMessage
  - 消息提示
summary: "ElMessage.success/error/warning/info 4 种 + showClose 关闭按钮 + duration 自定义时长；删除/保存等异步操作后给用户反馈。"
---

# Element Plus ElMessage 实战

dev-notebook 中所有"操作完成给用户反馈"的提示都用 ElMessage。本文讲 4 种消息 + showClose + 实战模式。

## 一句话结论

> 异步操作（保存 / 删除 / 导入）后用 ElMessage 反馈，4 种类型对应 4 种语义；批量操作按状态码区分（201/207/400）。

## 4 种基础消息

```ts
import { ElMessage, ElMessageBox } from 'element-plus'

// 1. 成功
ElMessage.success('保存成功')

// 2. 错误（操作失败）
ElMessage.error('删除失败：' + error.message)

// 3. 警告（操作成功但需注意）
ElMessage.warning('部分文件导入失败：3/10')

// 4. 普通信息
ElMessage.info('已复制到剪贴板')
```

| 类型 | 颜色 | 语义 | 实战场景 |
| --- | --- | --- | --- |
| `success` | 绿 | 操作完成、没问题 | 保存成功、删除成功、复制成功 |
| `error` | 红 | 操作失败 | 删除失败、API 报错 |
| `warning` | 黄 | 操作有风险/部分成功 | 部分文件导入失败、删完最后一条 |
| `info` | 灰 | 中性通知 | 复制到剪贴板、URL 变化提示 |

## 关键配置

```ts
ElMessage.success({
  message: '保存成功',
  duration: 2000,        // 2 秒后自动消失，默认 3000
  showClose: true,      // 显示关闭按钮
  type: 'success',      // 跟函数名一致时可省略
  center: true,         // 居中显示
  grouping: true,       // 同消息合并（避免重复弹）
})
```

| 配置 | 默认 | 实战建议 |
| --- | --- | --- |
| `duration` | 3000ms | success 用 2000（短）、error 用 0（不自动关，等用户关） |
| `showClose` | false | 失败消息必加 `true`（让用户关） |
| `grouping` | false | 多次相同消息合并显示（防刷屏） |
| `center` | false | 重要消息居中 |

## 实战：dev-notebook 批量导入反馈

```ts
import { ElMessage } from 'element-plus'

async function importNotes(files: File[]) {
  try {
    const result = await apiPost<ImportResult>('/api/notes/import', formData)

    // 201 = 全部成功
    // 207 = 部分成功
    // 400 = 全部失败
    if (result.failed === 0) {
      ElMessage.success(`成功导入 ${result.succeeded} 条笔记`)
    } else if (result.succeeded === 0) {
      ElMessage.error(`导入失败：${result.failed} 条`)
    } else {
      ElMessage.warning(`部分成功：${result.succeeded} 成功，${result.failed} 失败`)
    }
  } catch (e) {
    ElMessage.error(`导入失败：${e instanceof Error ? e.message : String(e)}`)
  }
}
```

## 实战：删除确认（ElMessageBox）

```ts
import { ElMessage, ElMessageBox } from 'element-plus'

async function deleteNote(id: number) {
  try {
    await ElMessageBox.confirm(
      '确定要删除这条笔记吗？此操作不可恢复。',
      '删除确认',
      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
  } catch {
    return  // 用户点取消
  }

  try {
    await deleteNoteAPI(id)
    ElMessage.success('删除成功')
  } catch (e) {
    ElMessage.error(`删除失败：${e instanceof Error ? e.message : String(e)}`)
  }
}
```

**关键**：
- `ElMessageBox.confirm` 抛 promise，用户点取消会进入 catch 块
- 第一段 try/catch 处理"用户取消"（不是错误）
- 第二段 try/catch 处理"API 失败"

## 实战：NoteCard 右上角删除按钮

```vue
<script setup lang="ts">
import { ElMessage, ElMessageBox } from 'element-plus'

const emit = defineEmits<{ deleted: [id: number] }>()

async function onDelete() {
  try {
    await ElMessageBox.confirm('确定要删除这条笔记吗？', '删除确认', {
      type: 'warning',
    })
  } catch {
    return
  }

  try {
    await deleteNoteAPI(props.note.id)
    ElMessage.success('删除成功')
    emit('deleted', props.note.id)  // 让父组件更新列表
  } catch (e) {
    ElMessage.error('删除失败：' + (e instanceof Error ? e.message : String(e)))
  }
}
</script>
```

**emit 模式**：Card 调 API + emit，父组件（NoteListView）更新 state。比 reload 整个列表高效。

## 常见误区

- **"我每个 catch 都 ElMessage.error，但 API 错误堆栈也弹"**——catch 块只 message 字符串，不要把整个 error 对象 stringify
- **"我 duration 设 0 让它一直显示"**——duration: 0 是不自动关，**但用户可能忘了关**——失败消息加 `showClose: true` 让用户主动关
- **"我每次操作都 ElMessage.success，页面到处是提示"**——grouping 合并同消息，或者只在关键操作弹
- **"我在 setup 顶层调 ElMessage"**——SSR 场景会爆（无 document），dev-notebook 是 SPA 不影响
- **"我手动 catch 错误但忘了给用户反馈"**——静默失败最差，至少 ElMessage.error

## 调试位置

- **"ElMessage 不显示"**——Element Plus 没注册 ElMessage / ElMessageBox 组件（按需引入漏了）
- **"ElMessage 类型找不到"**——`npm run dev` 第一次跑才会生成 `auto-imports.d.ts`
- **"ElMessageBox.confirm 没阻塞"**——它返回 promise，**必须 await**，否则点击"确认"后代码继续跑
- **"ElMessage 报错跨域"**——跟 ElMessage 无关，是 fetch 跨域，检查后端 CORS

## 关联知识点

- `vue/004-el-select` — 选完分类后用 ElMessage 反馈
- `vue/002-fetch-wrapper-pattern` — fetch wrapper 把 4xx/5xx 包装成 ApiError，组件用 try/catch + ElMessage
- `vue/003-typescript-generics-and-http-status` — ApiError 配合 ElMessage 精细化提示
