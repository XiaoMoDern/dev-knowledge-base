---
id: vue-init-timing
title: "Vue 初始化时序：onMounted 同步改 ref + watch 同步触发 = 双请求/双副作用"
category: Vue
tags:
  - Vue
  - onMounted
  - watch
  - 初始化
  - 时序
summary: "URL 同步场景：onMounted 同步改 ref → watch 同步触发 → 300ms 后又 load() = 双请求。isInitializing flag 拦截是通用解。"
---

# Vue 初始化时序坑

dev-notebook Phase C NoteListView 踩过：URL 同步搜索条件时，`onMounted` 同步改 `searchKeyword` ref → `watch` 同步触发 → 300ms 后又 `load()` = **双请求 + 视觉闪烁**。

## 一句话结论

> `onMounted` 同步改 ref 会让 `watch` 同步触发，产生"初始化副作用"。**`isInitializing` flag 拦截是通用解**——初始化期间 ref 变化不触发 watch。

## 错例（双请求 + 闪烁）

```vue
<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'

const searchKeyword = ref('')

watch(searchKeyword, () => {
  // ↑ 同步触发：onMounted 改 searchKeyword → 立即触发 watch
  triggerSearch(false)  // 300ms 后又 load
})

onMounted(() => {
  searchKeyword.value = route.query.q as string  // 同步改
  loadCategories()  // 立即调
  load()  // 立即调
})
</script>
```

**症状**：
1. 第 1 个请求：onMounted 直接 `load()`（搜当前 URL q 的结果）
2. 第 2 个请求：300ms 后 watch 触发 `load()`（同样 URL q，但走 watch 路径）
3. 视觉：第 1 个 load 完 `v-loading` 消失 → 300ms 后第 2 个 load `v-loading` 又显示 = **闪烁**

## 解法：isInitializing flag

```vue
<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'

const searchKeyword = ref('')
const isInitializing = ref(true)  // ← 关键

watch(searchKeyword, () => {
  if (isInitializing.value) return  // ← 初始化期间不触发
  triggerSearch(false)
})

onMounted(async () => {
  try {
    // 1. 从 URL 读状态
    searchKeyword.value = route.query.q as string
    // ↑ 改 ref → watch 同步触发 → 但 isInitializing=true → return
    
    // 2. 加载分类（同步）
    await loadCategories()
    
    // 3. 加载笔记（同步）
    await load()
    // ↑ 只这一次 load，无 watch 触发
  } finally {
    isInitializing.value = false  // ← 初始化完成
  }
})
</script>
```

**优点**：
- 初始化期间所有 ref 变化都被拦截
- 初始化完成（finally）后，watch 恢复正常
- 单次 load，无双请求

## 为什么 watch 同步触发？

Vue 3 默认 `watch` 是 `flush: 'pre'`（组件更新前），但 ref 的**同步赋值**会同步触发 watch 回调（不是异步）。

```ts
const x = ref(0)
watch(x, (newVal) => {
  console.log('watch:', newVal)
})
x.value = 1
// 立即输出 'watch: 1'（同步触发）
```

**前端类比**：跟 React `useEffect` 不一样——React `useEffect` 异步触发（commit 后），所以 React 没这问题。Vue 的 watch 是同步的。

## 三种 watch flush 模式

```ts
watch(source, callback, {
  flush: 'pre' | 'post' | 'sync'
})
```

| 模式 | 触发时机 | 用途 |
| --- | --- | --- |
| `'pre'`（默认） | 组件更新前，DOM 未更新 | 大多数场景（要在 DOM 更新前改 state） |
| `'post'` | 组件更新后，DOM 已更新 | 要访问更新后的 DOM |
| `'sync'` | ref 变化时**立即**同步触发 | 极少用（要小心同步副作用） |

dev-notebook 默认 `'pre'`，但**对 ref 赋值是同步触发 watch 回调**——跟组件更新时机无关。

## 实战：dev-notebook Phase C NoteListView 完整 onMounted

```ts
async function onMountedInit() {
  isInitializing.value = true
  try {
    // 1. 同步 URL 状态到 ref（不会触发 watch）
    searchKeyword.value = (route.query.q as string) || ''
    selectedCategoryId.value = parseCategoryId(route.query.categoryId)
    currentPage.value = parseInt((route.query.page as string) || '1')
    
    // 2. 加载分类（异步，但 watch 拦截中）
    await loadCategories()
    
    // 3. 加载笔记（异步，单次）
    await load()
  } finally {
    isInitializing.value = false
  }
}

onMounted(onMountedInit)
```

## 实战：URL 变化也要初始化拦截

```ts
watch(() => route.query, (newQuery) => {
  if (isInitializing.value) return
  searchKeyword.value = (newQuery.q as string) || ''
  load()
})
```

**场景**：用户点浏览器后退，URL 变化，watch 重新加载——但要避免"后退瞬间 onMounted 还没跑完就触发 watch"。

## 其他解法（不推荐）

### 解法 B：watchImmediate: false

```ts
watch(searchKeyword, () => triggerSearch(false), { immediate: false })
```

**问题**：默认就是 `immediate: false`，改这个没用——`onMounted` 改 ref 还是会触发 watch。

### 解法 C：watch flush: 'post'

```ts
watch(searchKeyword, () => triggerSearch(false), { flush: 'post' })
```

**问题**：`'post'` 只是延迟到 DOM 更新后，**onMounted 同步赋值还是会同步触发**——不解决根本问题。

### 解法 D：nextTick 包一层

```ts
onMounted(async () => {
  await nextTick()  // 等 watch 触发完
  searchKeyword.value = route.query.q as string
})
</script>
```

**问题**：watch 在 nextTick 之前已经触发；`searchKeyword` 改完后，watch 又触发——还是双请求。

**结论**：`isInitializing` flag 拦截是唯一干净解。

## 常见误区

- **"onMounted 改 ref 不应该触发 watch"**——会触发，Vue 3 watch 默认同步
- **"我加 immediate: false 就行"**——immediate: false 是默认值，改不改都不影响
- **"用 flush: 'post' 延迟"**——不解决同步问题
- **"用 nextTick 等"**——会改两次 ref，watch 触发两次
- **"我直接 off + on 监听"**——可以但比 flag 复杂

## 调试位置

- **"页面刷新有 2 个相同的 API 请求"**——onMounted 改 ref + watch 同步触发
- **"v-loading 闪一下又显示"**——load() 跑两次
- **"后退按钮触发多次 watch"**——route query 变化触发 watch，没 flag 拦截

## 关联知识点

- `vue/006-v-model-double-trigger` — v-model 也触发 watch，跟初始化时序叠加
- `vue/004-el-select` — el-select change 事件也要初始化拦截
- `vue/002-fetch-wrapper-pattern` — load() 用 fetch wrapper，单次 / 双次都好排查
