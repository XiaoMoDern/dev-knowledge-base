<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import CategorySidebar from '../components/CategorySidebar.vue'

const route = useRoute()
const router = useRouter()
const userName = 'Ray'

// 详情/编辑页隐藏侧边栏（单 note 操作不需要分类切换）
const showSidebar = computed(() => !route.meta.hideSidebar)

function goHome() {
  router.push('/')
}
</script>

<template>
  <div class="layout">
    <header class="topbar">
      <div class="topbar-left">
        <div class="logo" @click="goHome">
          <span class="logo-icon">📓</span>
          <span class="logo-text">Dev Notebook</span>
        </div>
      </div>
      <div class="topbar-right">
        <button class="user-menu">
          <span class="user-avatar">{{ userName.charAt(0) }}</span>
          <span class="user-name">{{ userName }}</span>
        </button>
      </div>
    </header>

    <div class="body">
      <aside v-if="showSidebar" class="sidebar">
        <CategorySidebar />
      </aside>

      <main class="content">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<style scoped>
.layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--color-bg);
}

.topbar {
  height: var(--header-height);
  background: var(--color-bg-elevated);
  border-bottom: 1px solid var(--color-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--space-xl);
  position: sticky;
  top: 0;
  z-index: 10;
}

.topbar-left { display: flex; align-items: center; gap: var(--space-md); }

.logo {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  cursor: pointer;
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--color-text);
}
.logo-icon { font-size: 1.5rem; }

.topbar-right { display: flex; align-items: center; gap: var(--space-md); }

.user-menu {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  background: transparent;
  border: none;
  padding: 0.4rem 0.75rem;
  border-radius: var(--radius-md);
  cursor: pointer;
  color: var(--color-text);
  transition: background 0.15s;
}
.user-menu:hover { background: var(--color-bg); }

.user-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 0.875rem;
}

.user-name { font-size: 0.875rem; font-weight: 500; }

.body {
  flex: 1;
  display: flex;
  min-height: 0;
  align-items: flex-start; /* 让 sticky 子元素按内容高度对齐，不被 stretch */
}

.sidebar {
  width: var(--sidebar-width);
  background: var(--color-bg-elevated);
  border-right: 1px solid var(--color-border);
  flex-shrink: 0;
  position: sticky;
  top: var(--header-height);
  height: calc(100vh - var(--header-height));
  overflow-y: auto;
}

.content {
  flex: 1;
  min-width: 0;
  /* 不设 overflow：让 body 滚，sticky 子元素才能粘到视口 */
}
</style>
