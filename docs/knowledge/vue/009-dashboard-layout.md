---
id: vue-dashboard-layout
title: "Dashboard 布局标准模式：body 滚动 + header sticky + sidebar sticky + content 不 overflow"
category: Vue
tags:
  - CSS
  - dashboard
  - layout
  - sticky
  - flexbox
summary: "dashboard 布局：body 滚（最外层）、header sticky top: 0、sidebar sticky 自己 overflow-y: auto、main 不设 overflow。align-items: flex-start 让 sidebar 自定义高度。"
---

# Dashboard 布局标准模式

dev-notebook UI 重构 07-22 实战：参考 WorkBuddy 成长中心页（顶部 + 侧边栏 + 卡片）的 dashboard 布局，**有 4 个铁律**。

## 一句话结论

> dashboard 布局 = body 滚动 + header `position: sticky; top: 0` + sidebar `position: sticky; top: 64px; overflow-y: auto` + main 不设 overflow。**`align-items: flex-start` 让 sidebar 自定义高度**，否则 sticky 失效。

## 标准结构

```vue
<template>
  <div class="layout">
    <!-- 1. header：粘视口顶部 -->
    <header class="header">
      <div class="logo">dev-notebook</div>
      <div class="user-menu">...</div>
    </header>

    <!-- 2. body：flex 横向布局 -->
    <div class="body">
      <!-- 3. sidebar：粘 header 下面，自己 overflow -->
      <aside class="sidebar">
        <CategorySidebar />
      </aside>

      <!-- 4. main：不设 overflow，body 滚 -->
      <main class="main">
        <RouterView />
      </main>
    </div>
  </div>
</template>
```

## 标准样式

```css
.layout {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background: #f5f5f5;
}

/* 1. header：粘视口顶部 */
.header {
  position: sticky;
  top: 0;
  height: 64px;
  background: white;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
  z-index: 10;
}

/* 2. body：横向 flex */
.body {
  display: flex;
  flex: 1;
  align-items: flex-start;  /* ← 关键 */
}

/* 3. sidebar：粘 header 下面 */
.sidebar {
  position: sticky;
  top: 64px;
  width: 240px;
  height: calc(100vh - 64px);
  overflow-y: auto;  /* ← sidebar 自己滚 */
  background: white;
  border-right: 1px solid #e5e7eb;
}

/* 4. main：不设 overflow */
.main {
  flex: 1;
  padding: 24px 32px;
  /* ← 不设 overflow-y: auto */
}
```

## 4 个铁律

### 铁律 1：body 滚动（最外层）

```css
html, body {
  margin: 0;
  /* 默认 body 滚 */
}
```

**为什么**——sticky 元素粘到视口才有意义；如果某个父容器自己滚，sticky 元素被裁在父容器内，不粘视口。

### 铁律 2：header `position: sticky; top: 0`

```css
.header {
  position: sticky;
  top: 0;  /* 视口顶部 */
  z-index: 10;
}
```

**为什么**——`top: 0` 让 header 滚到视口顶部时停住；`z-index: 10` 防止被内容盖住。

### 铁律 3：sidebar `position: sticky; top: 64px; overflow-y: auto`

```css
.sidebar {
  position: sticky;
  top: 64px;  /* header 高度 64px */
  height: calc(100vh - 64px);  /* 视口减 header */
  overflow-y: auto;  /* sidebar 内容多时自己滚 */
}
```

**为什么**——sidebar 跟着滚但粘在 header 下面；自己 `overflow-y: auto` 让分类列表多时 sidebar 内部滚，不影响 main。

### 铁律 4：main 不设 overflow

```css
.main {
  flex: 1;
  /* ← 不设 overflow */
}
```

**为什么**——main 是 body 滚的一部分，**不能有 overflow**（否则 sticky 在 main 内失效）。

### 隐藏铁律 5：`align-items: flex-start`

```css
.body {
  display: flex;
  align-items: flex-start;  /* ← 关键，不能 stretch */
}
```

**为什么**——flex 默认 `align-items: stretch`，sidebar 被拉伸到跟 main 等高。但 sticky 元素需要"超出父容器"才有效果——sidebar 跟 main 等高，sticky 没意义。

`flex-start` 让 sidebar 自定义高度（`calc(100vh - 64px)`），main 流式高度，sticky 才能"超出 main 时粘住"。

## 实战：dev-notebook MainLayout.vue

```vue
<template>
  <div class="layout">
    <header class="header">
      <div class="header-left">
        <h1 class="logo">📓 dev-notebook</h1>
      </div>
      <div class="header-right">
        <el-button @click="goToNew">新建笔记</el-button>
      </div>
    </header>

    <div class="body">
      <aside v-if="!route.meta.hideSidebar" class="sidebar">
        <CategorySidebar />
      </aside>

      <main class="main">
        <RouterView v-slot="{ Component }">
          <component :is="Component" />
        </RouterView>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
const route = useRoute()
</script>

<style scoped>
.layout {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background: var(--color-bg);
}
.header {
  position: sticky;
  top: 0;
  height: 64px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0 32px;
  background: var(--color-bg-elevated);
  box-shadow: var(--shadow-sm);
  z-index: 10;
}
.body {
  display: flex;
  flex: 1;
  align-items: flex-start;
}
.sidebar {
  position: sticky;
  top: 64px;
  width: 240px;
  height: calc(100vh - 64px);
  overflow-y: auto;
  background: var(--color-bg-elevated);
  border-right: 1px solid var(--color-border);
  padding: 16px 0;
}
.main {
  flex: 1;
  padding: 24px 32px;
  min-width: 0;
}
</style>
```

**关键设计**：
- `v-if="!route.meta.hideSidebar"` — 详情/编辑页隐藏侧边栏（meta 控制 layout 行为）
- `--color-bg-elevated` — 比 body 背景白一阶，让 sidebar 浮起来
- `min-width: 0` — 防止 flex 子元素被内容撑大

## Vue Router meta 控制 layout 行为

```ts
// router.ts
{
  path: '/notes/:id',
  component: NoteDetailView,
  meta: { hideSidebar: true }  // ← 详情页隐藏 sidebar
}
```

```vue
<aside v-if="!route.meta.hideSidebar" class="sidebar">
```

**优点**——layout 通过 `route.meta` 读路由级元信息，控制行为。比 prop 传递更解耦（不用每个 view 传 hideSidebar）。

**其他 meta 用途**：
- `meta: { requiresAuth: true }` — 强制登录
- `meta: { hideHeader: true }` — 登录页隐藏 header
- `meta: { title: '首页' }` — 动态 title

## 实战：路由嵌套（避免 App.vue 重复包 layout）

```ts
// router.ts（嵌套路由）
{
  path: '/',
  component: MainLayout,  // 父路由
  children: [
    { path: '', component: NoteListView },
    { path: 'notes/new', component: NoteEditView },
    { path: 'notes/:id', component: NoteDetailView, meta: { hideSidebar: true } },
    { path: 'notes/:id/edit', component: NoteEditView, meta: { hideSidebar: true } },
  ]
}
```

```vue
<!-- App.vue：只放 RouterView -->
<template>
  <RouterView />
</template>

<!-- MainLayout.vue：放 layout + 子 RouterView -->
<template>
  <div class="layout">
    <header>...</header>
    <div class="body">
      <aside>...</aside>
      <main>
        <RouterView />  <!-- 子路由出口 -->
      </main>
    </div>
  </div>
</template>
```

**避免双重 layout**——App.vue 不能再包 MainLayout（嵌套 = 重复 = 双重 header/sidebar）。

## 常见误区

- **"main 设 `overflow-y: auto` 防止溢出"**——会让 main 内 sticky 失效，让 main 流式
- **"sidebar 用 `flex: 1` 跟 main 等宽"**——sidebar 是固定 240px，不要 flex
- **"用 grid 替代 flex 布局"**——grid 适合二维（行列），dashboard 一维（横向）用 flex 更简单
- **"header 用 `position: fixed`"**——fixed 完全脱离文档流，要手动 padding-top 补偿；sticky 保留文档流
- **"App.vue 也包一层 MainLayout"**——双重 layout，嵌套 bug

## 调试位置

- **"header 不粘"**——父容器有 overflow，找到并删掉
- **"sidebar 跟 main 等高"**——`align-items: stretch` 改成 `flex-start`
- **"sidebar 滚到 header 下面就停"**——`top: 64px` 跟 header 高度对齐
- **"内容被 sidebar 盖住"**——sidebar `z-index` 太高或 main 缺 `min-width: 0`
- **"路由切换 layout 闪"**——App.vue 重复包了 MainLayout，删掉一层

## 关联知识点

- `vue/008-sticky-overflow` — sticky 在父容器有 overflow 时失效
- `vue/004-el-select` — sidebar 内的 el-select 懒加载
- 跨项目教训 19 — Vue Router 嵌套路由标准模式
