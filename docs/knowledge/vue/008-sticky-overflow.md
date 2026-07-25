---
id: vue-sticky-overflow
title: "CSS position: sticky 在父容器有 overflow 时失效"
category: Vue
tags:
  - CSS
  - position: sticky
  - overflow
  - dashboard
  - 布局
summary: "sticky 元素被裁在最近的可滚动父容器边界。父容器有 overflow: auto/scroll/hidden 时，sticky 只在父容器内有效，不粘视口。"
---

# CSS Sticky Overflow 陷阱

dev-notebook UI 重构踩过：详情页 `action-bar` 加 `position: sticky; top: 64px`，**结果跟着 markdown body 一起滚，完全没粘**。根因：父容器 `.content { overflow-y: auto }` 让 sticky 失效。

## 一句话结论

> `position: sticky` 在**父容器有 `overflow`（auto/scroll/hidden）**时，sticky 只在父容器内有效，**不粘视口**。sticky 元素被裁在最近的可滚动父容器边界。

## 错例（sticky 失效）

```vue
<template>
  <div class="content">  <!-- ← 父容器有 overflow-y: auto -->
    <div class="action-bar">  <!-- ← 想 sticky top: 64px -->
      <h1>{{ note.title }}</h1>
      <button>编辑</button>
    </div>
    <div class="markdown-body">
      {{ note.content }}
    </div>
  </div>
</template>

<style>
.content {
  overflow-y: auto;  /* ← 罪魁祸首 */
  height: 100vh;
}
.action-bar {
  position: sticky;
  top: 64px;  /* ← 失效：sticky 在 .content 内滚动到 top 就停，不粘视口 */
}
</style>
```

**症状**：
- 用户滚 markdown body，action-bar 跟着一起滚
- 完全不"粘"——看起来跟普通 div 没区别
- 改 `position: fixed` 倒是粘了，但完全脱离文档流要手动 padding-top 补偿

## sticky 触发条件（3 个全满足才生效）

1. **`position: sticky`**（不是 static/relative）
2. **`top` / `bottom` / `left` / `right` 至少一个**（不能只写 `position: sticky`，必须有方向）
3. **父容器不能有 `overflow: auto/scroll/hidden`**（或 sticky 元素的"最近可滚动祖先"不能有 overflow）

**第 3 条最反直觉**——overflow 本来是"让父容器可滚"，反而让 sticky 失效。

## 修法：body 滚动 + sidebar 自己 overflow

```vue
<template>
  <div class="layout">
    <header class="header">  <!-- sticky top: 0，粘视口 -->
      <h1>dev-notebook</h1>
    </header>
    <div class="body">
      <aside class="sidebar">  <!-- sticky，自己 overflow-y: auto -->
        <CategorySidebar />
      </aside>
      <main class="main">
        <router-view />  <!-- 不设 overflow，body 滚 -->
      </main>
    </div>
  </div>
</template>

<style>
/* 关键：body 滚，content 不设 overflow */
.layout {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}
.header {
  position: sticky;
  top: 0;  /* ← 粘视口，因为没祖先有 overflow */
  height: 64px;
  z-index: 10;
}
.body {
  display: flex;
  flex: 1;
  align-items: flex-start;  /* ← 关键：sidebar 自己定义高度，sticky 需要 */
}
.sidebar {
  position: sticky;
  top: 64px;  /* ← 粘 header 下面 */
  height: calc(100vh - 64px);
  width: 240px;
  overflow-y: auto;  /* ← sidebar 自己滚，不影响 main */
}
.main {
  flex: 1;
  /* ← 不设 overflow，body 滚 */
}
</style>
```

**关键**：
- body 滚动（最外层）
- header `position: sticky; top: 0` 粘视口顶部
- sidebar `position: sticky; top: 64px` + `height: calc(100vh - 64px)` + `overflow-y: auto` 自己滚
- main 不设 overflow
- body `align-items: flex-start`（sidebar 高度自定义，不能 stretch）

## 实战：dev-notebook 详情页 sticky action-bar

```vue
<template>
  <div class="detail-view">
    <div class="action-bar">  <!-- sticky top: 80px（header 64 + 间距 16） -->
      <h1>{{ note.title }}</h1>
      <div class="actions">
        <el-button @click="goBack">返回</el-button>
        <el-button @click="deleteNote" type="danger">删除</el-button>
        <el-button @click="editNote" type="primary">编辑</el-button>
      </div>
    </div>
    <div class="meta">
      <span v-if="note.categoryName">{{ note.categoryName }}</span>
      <span>{{ formatDate(note.updatedAt) }}</span>
    </div>
    <div class="markdown-body" v-html="renderedContent" />
  </div>
</template>

<style scoped>
.detail-view {
  /* 不设 overflow */
}
.action-bar {
  position: sticky;
  top: 80px;  /* header 64 + space-md 16 */
  background: var(--color-bg-elevated);
  box-shadow: var(--shadow-md);
  border-radius: var(--radius-lg);
  border: 1px solid var(--color-border);
  z-index: 5;
}
</style>
```

**为什么能粘视口**——`MainLayout` 的 `.main` 不设 overflow，body 滚，所以 action-bar 粘到 top: 80px 就停。

## sticky 视觉层级（让"飘起来"）

```css
.action-bar {
  position: sticky;
  top: 80px;
  /* 视觉：白底 + 阴影 + 圆角 + 边框 */
  background: white;  /* 跟内容区分 */
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);  /* 浮起 */
  border-radius: 12px;  /* 圆角 */
  border: 1px solid #e5e7eb;  /* 边框 */
  padding: 16px 20px;
  z-index: 5;
}
```

**没有视觉提示，sticky 区跟内容糊一起**——白底 + 阴影 + 圆角 + 边框，让用户明确"这是浮起来的卡片"。

## sticky 元素高度 vs 父容器

```css
.body {
  display: flex;
  align-items: flex-start;  /* ← 关键 */
}
.sidebar {
  position: sticky;
  top: 64px;
  height: calc(100vh - 64px);  /* ← 必须显式高度 */
}
```

**如果 `align-items: stretch`**（默认）——sidebar 跟 main 等高，但 sticky 元素高度被父容器限制，sticky 不需要了——因为 sticky 元素跟父容器一样高，**根本没有"超出"的部分可滚**。

`align-items: flex-start` 让 sidebar 自己定义高度（calc(100vh - 64px)），main 流式高度，sticky 才能"超出 main 时粘住"。

## 常见误区

- **"我加了 `position: sticky` 但完全没效果"**——父容器有 overflow，找到最近的有 overflow 的祖先删掉
- **"用 `fixed` 替代 sticky 就好"**——fixed 完全脱离文档流，要手动 padding-top 补偿
- **"sticky 跟 z-index 无关"**——sticky 元素会被其他内容盖住，必须加 z-index
- **"sticky 元素可以无限滚"**——sticky 只在父容器内有效，滚出父容器就跟普通元素一样

## 调试位置

- **"sticky 不粘视口"**——父容器有 `overflow: auto/scroll/hidden`
- **"sticky 元素消失了"**——z-index 太低被盖住
- **"sticky 抖一下"**——父容器高度变化（比如图片加载完），改成固定高度
- **"sticky 完全没效果像普通 div"**——忘了写 `top` / `bottom`

## 关联知识点

- `vue/009-dashboard-layout` — dashboard 标准布局：body 滚 + header/sidebar sticky + main 不 overflow
- `vue/002-fetch-wrapper-pattern` — sticky 不影响 fetch
