import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    // 按需自动引入 Element Plus 组件 + API
    // - 组件：写 <el-button> 不用手 import
    // - API：ElMessage / ElMessageBox 等直接用
    // - dts 生成在 src/，被 tsconfig.app.json 的 include 自动覆盖
    AutoImport({
      resolvers: [ElementPlusResolver()],
      dts: 'src/auto-imports.d.ts',
    }),
    Components({
      resolvers: [ElementPlusResolver()],
      dts: 'src/components.d.ts',
    }),
  ],
  server: {
    host: '0.0.0.0', // 监听所有 IPv4 网卡，让局域网设备能通过 http://172.18.4.121:5173 访问
    port: 5173,
    proxy: {
      '/api': 'http://127.0.0.1:8181',
    },
  },
  // Vitest 配置：DOMPurify 依赖浏览器 DOM，需要 jsdom 环境
  test: {
    environment: 'jsdom',
    include: ['src/**/*.test.ts'],
  },
})
