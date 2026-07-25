---
id: vue-v-model-double-trigger
title: "v-model + @input 双重触发：Vue 最经典的搜索框坑"
category: Vue
tags:
  - Vue
  - v-model
  - @input
  - watch
  - 搜索
summary: "v-model 内部走 update:modelValue，@input 又触发一次 = 双重触发。正确用 watch(ref) 或 @update:model-value，二选一。"
---

# v-model + @input 双重触发

dev-notebook Phase C NoteListView 搜索框踩过：v-model 绑搜索关键字 + @input 触发搜索，**用户每按一个键请求两次接口**。这是 Vue 搜索框的经典坑。

## 一句话结论

> v-model 内部已经走 `update:modelValue`，再加 `@input` 触发 = **双重触发**。正确用 `watch(searchKeyword, ...)` 监听 ref 变化，或者 `@update:model-value="..."` 显式监听，**v-model + @input 二选一**。

## 错例（双重触发）

```vue
<el-input
  v-model="searchKeyword"
  @input="onSearch"
  placeholder="搜索笔记"
/>

<script setup lang="ts">
function onSearch(value: string) {
  // ↑ 每次按键触发两次
  // 第 1 次：el-input 内部 v-model 走 update:modelValue
  // 第 2 次：@input 显式触发
  fetchResults(value)
}
</script>
```

**症状**：
- 用户每按一个键，后端收到 2 个请求
- 即使加 debounce 300ms，也会产生 2 个 setTimeout（清掉一个但浪费调度）
- 网络面板看搜索请求数量 = 2x 实际按键数

## 解法 1：watch 监听 ref（推荐）

```vue
<el-input
  v-model="searchKeyword"
  placeholder="搜索笔记"
/>

<script setup lang="ts">
import { ref, watch } from 'vue'

const searchKeyword = ref('')
let debounceTimer: number | null = null

watch(searchKeyword, (newValue) => {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = window.setTimeout(() => {
    fetchResults(newValue)
  }, 300)
})
</script>
```

**优点**：
- 单次触发（watch 只在 ref 真正变化时触发）
- debounce 干净
- 跟响应式系统对齐（不只是 input 事件）

## 解法 2：@update:model-value 显式监听

```vue
<el-input
  :model-value="searchKeyword"
  @update:model-value="onSearch"
  placeholder="搜索笔记"
/>

<script setup lang="ts">
function onSearch(value: string) {
  searchKeyword.value = value
  fetchResults(value)
}
</script>
```

**优点**：
- 显式 v-model 不参与，单次触发
- 缺点：失去 v-model 双向绑定便利

## 解法 3：v-model + ref + watch + 立即触发按钮（dev-notebook 实际采用）

```vue
<el-input
  v-model="searchKeyword"
  placeholder="搜索笔记"
  @keyup.enter="triggerSearch(true)"
/>
<el-button @click="triggerSearch(true)">查询</el-button>

<script setup lang="ts">
import { ref, watch } from 'vue'

const searchKeyword = ref('')

function triggerSearch(immediate = false) {
  fetchResults(searchKeyword.value, immediate)
}

watch(searchKeyword, () => {
  // 输入时 debounce
  triggerSearch(false)
})
</script>
```

**优点**：
- 输入触发（debounce 300ms）
- 按 Enter / 点按钮立即触发（跳过 debounce）
- 单次触发（只有 watch 触发，没有 @input）

## 实战：dev-notebook Phase C 完整代码

```vue
<template>
  <el-input
    v-model="searchKeyword"
    placeholder="搜索笔记"
    clearable
    @keyup.enter="triggerSearch(true)"
  >
    <template #append>
      <el-button @click="triggerSearch(true)">查询</el-button>
    </template>
  </el-input>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

const searchKeyword = ref('')
let debounceTimer: ReturnType<typeof setTimeout> | null = null

function triggerSearch(immediate: boolean) {
  if (debounceTimer) {
    clearTimeout(debounceTimer)
    debounceTimer = null
  }
  if (immediate) {
    loadNotes()
  } else {
    debounceTimer = setTimeout(() => {
      loadNotes()
    }, 300)
  }
}

watch(searchKeyword, (newVal, oldVal) => {
  // clearable 触发时 newVal 可能为空
  if (newVal !== oldVal) {
    triggerSearch(false)
  }
})
</script>
```

## 为什么 v-model + @input 双重触发？

el-input 内部：
- `v-model="x"` 编译为 `:model-value="x" @update:model-value="x = $event"`
- 用户按键 → 内部 emit `update:modelValue` → v-model 更新 x
- 同时 emit `input`（原 DOM 事件）→ @input 触发

**两个事件都触发 = 双请求**。

## 常见误区

- **"我用 debounce 兜底，双重触发无所谓"**——debounce 只是延后，清掉 1 个 setTimeout 还是浪费调度
- **"v-model 跟 @change 不会双触发"**——@change 在 blur 时触发，不在 input 时，不会双触发（但失去实时搜索）
- **"我用 @input.lazy 延迟到 change"**——失去实时搜索，UX 差
- **"我看 Vue 文档没提这坑"**——文档默认 v-model + @input 不混用，新手不知道

## 调试位置

- **"搜索请求数量 = 2x 实际按键数"**——v-model + @input 双重触发
- **"debounce 不生效，每次按键立即请求"**——@input 在 v-model 外层，watch 还没触发就被 @input 先跑了
- **"按 Enter 也触发 2 次"**——@keyup.enter 跟 watch 都会触发，dev-notebook 解法是 `triggerSearch(true)` 走 immediate 分支

## 关联知识点

- `vue/007-init-timing` — 初始化时 watch 也会触发（双重副作用）
- `vue/005-ElMessage` — debounce 完用 ElMessage 给反馈
- `vue/004-el-select` — el-select 也有 v-model，要避免 @change 双重触发
