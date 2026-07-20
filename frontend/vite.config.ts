import { defineConfig } from 'vite'
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
    proxy: {
      '/api': 'http://127.0.0.1:8181',
    },
  },
})
