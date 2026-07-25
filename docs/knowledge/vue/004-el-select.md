---
id: vue-el-select
title: "Element Plus el-select：分类下拉的实战模式"
category: Vue
tags:
  - Vue
  - Element Plus
  - el-select
  - 下拉框
summary: "el-select v-model + filterable + clearable + placeholder + change 事件；'未分类'用 number 哨兵值或单独 option。"
---

# Element Plus el-select 实战

dev-notebook Phase B 分类下拉的实战模式。Element Plus 2.14，按需自动引入。

## 一句话结论

> el-select 配 `v-model` + `filterable clearable placeholder` 三件套，change 事件触发数据加载，"未分类"用 `number` 哨兵值（不是 `null/undefined`）。

## 基础用法

```vue
<template>
  <el-select
    v-model="selectedCategoryId"
    placeholder="选择分类"
    filterable
    clearable
    @change="onCategoryChange"
    style="width: 200px"
  >
    <el-option label="未分类" :value="0" />
    <el-option
      v-for="cat in categories"
      :key="cat.id"
      :label="cat.name"
      :value="cat.id"
    />
  </el-select>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { listCategories } from '@/api/categories'
import type { Category } from '@/api/types'

const categories = ref<Category[]>([])
const selectedCategoryId = ref<number | undefined>(undefined)

async function onCategoryChange(newId: number | undefined) {
  if (newId === undefined) {
    // 用户点了 clearable 清除按钮
    selectedCategoryId.value = undefined
  } else if (newId === 0) {
    // "未分类" 哨兵值
    selectedCategoryId.value = 0
  } else {
    selectedCategoryId.value = newId
  }
  await loadNotes()
}
</script>
```

## 关键属性

| 属性 | 作用 | 实战用法 |
| --- | --- | --- |
| `v-model` | 双向绑定当前选中值 | 必填，类型 `number` / `string` 跟 `:value` 一致 |
| `filterable` | 可输入过滤 | 分类多时（10+）必加 |
| `clearable` | 显示清除按钮 | 几乎必加，用户能"取消选择" |
| `placeholder` | 占位符 | "选择分类" 比 "请选择" 友好 |
| `@change` | 选中变化时触发 | 调 `loadNotes()` 重新加载列表 |
| `multiple` | 多选 | 单选场景**不要加**（v-model 类型变 `Array`） |
| `collapse-tags` | 多选时折叠 tag | 配合 `multiple` 用 |

## "未分类"哨兵值设计

dev-notebook 后端 `category_id` 是 `*int64`（nullable），前端怎么表达"没分类"？

**方案 A**：`number` 哨兵值（dev-notebook 采用）

```vue
<el-option label="未分类" :value="0" />  <!-- id=0 表示"未分类" -->
```

```ts
const selectedCategoryId = ref<number | undefined>(undefined)  // 默认"全部"
function onCategoryChange(newId: number | undefined) {
  if (newId === 0) {
    // 调 API：searchNotes({ categoryId: undefined }) → 后端返"全部"
    // 调 API：searchNotes({ categoryId: 0 }) → 后端返"未分类"
  }
}
```

**优点**：URL 同步简单（`?categoryId=0` 直白）
**缺点**：哨兵值 0 跟数据库 id 冲突（如果未来 id 从 0 开始）

**方案 B**：`null` / `undefined` 三态

```ts
const selectedCategoryId = ref<number | null | 'all' | 'uncategorized'>('all')
```

**优点**：语义清晰
**缺点**：类型复杂、URL 同步难（`?categoryId=undefined` 丑）

**dev-notebook 选 A**——简单优先，URL `?categoryId=0` 是约定。

## 跟分类加载联动

```vue
<el-select
  v-model="selectedCategoryId"
  :loading="loadingCategories"
  @visible-change="onDropdownOpen"
>
```

`onDropdownOpen` 第一次展开时拉分类列表：

```ts
async function onDropdownOpen(visible: boolean) {
  if (visible && categories.value.length === 0) {
    loadingCategories.value = true
    try {
      categories.value = await listCategories()
    } finally {
      loadingCategories.value = false
    }
  }
}
```

**优点**：页面加载时不用立即请求分类，展开下拉时才加载（懒加载）。

## 实战：dev-notebook 侧边栏 CategorySidebar

```vue
<template>
  <el-menu
    :default-active="activeCategoryId?.toString() ?? 'all'"
    @select="onSelect"
  >
    <el-menu-item index="all">全部</el-menu-item>
    <el-menu-item index="uncategorized">未分类</el-menu-item>
    <el-menu-item
      v-for="cat in categories"
      :key="cat.id"
      :index="cat.id.toString()"
    >
      {{ cat.name }}
    </el-menu-item>
  </el-menu>
</template>
```

**为什么用 el-menu 不用 el-select？**——侧边栏需要一直显示所有分类（不只是当前选中的），el-select 是下拉框，el-menu 是菜单。

## 常见误区

- **"v-model 直接绑 `string`，类型不安全"**——TS 项目绑 `number` / `Category`，change 时类型自动 narrow
- **"clearable 后 v-model 变成 undefined 没法传给后端"**——后端 API 接 `number | undefined`，不要绑 `number` 强类型
- **"filterable 后输入中文搜索不到"**——el-select 默认按 `label` 过滤，要搜其他字段用 `:filter-method` 自定义
- **"加了 multiple 但 v-model 还是单值"**——v-model 类型是 `Array<number>`，单值会 TS 报错

## 调试位置

- **"v-model 不更新"**——`:value` 类型跟 v-model 类型不一致（比如 v-model 是 number，:value 是 string）
- **"clearable 后调 API 报错"**——change 事件没处理 `undefined` 分支
- **"filterable 输入框卡顿"**——分类 1000+ 时加 `:filter-method` 改远程搜索，不要全量渲染
- **"el-select 的 d.ts 找不到"**——`npm run dev` 第一次跑才会生成 `components.d.ts`

## 关联知识点

- `vue/005-ElMessage` — 选完分类后给用户反馈
- `vue/003-typescript-generics-and-http-status` — 后端 listCategories 返回类型
