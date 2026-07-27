<script setup lang="ts">
// 根组件只放 RouterView，路由切换由 vue-router 接管
// 嵌套路由：/ 路径的 component 是 MainLayout，自带顶部 + 侧边栏
//
// Element Plus 国际化：ElConfigProvider 给整棵组件树提供 locale 上下文。
// unplugin-vue-components 按需引入下，ElConfigProvider 也会自动注册，不用全局 install。
// 加载路径用 es/locale/lang/zh-cn（按需），不是 dist/locale（会拉全量语言包）。
import { ElConfigProvider } from 'element-plus'
import zhCn from 'element-plus/es/locale/lang/zh-cn'
</script>

<template>
  <ElConfigProvider :locale="zhCn">
    <RouterView />
  </ElConfigProvider>
</template>
