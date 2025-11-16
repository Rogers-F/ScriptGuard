import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

// Wails Runtime 替换插件
function wailsRuntimePlugin() {
  return {
    name: 'wails-runtime-replace',
    transform(code, id) {
      // 替换 bindings 文件中的 @wailsio/runtime 导入
      if (id.includes('bindings') && code.includes('@wailsio/runtime')) {
        // 使用固定相对路径: bindings/scriptguard/backend/app.js -> ../../../src/wails-runtime.js
        const replaced = code.replace(
          /from\s+["']@wailsio\/runtime["']/g,
          'from "../../../src/wails-runtime.js"'
        )

        console.log('🔧 替换 Wails runtime 导入:', id)

        return {
          code: replaced,
          map: null
        }
      }
      return null
    }
  }
}

export default defineConfig({
  plugins: [vue(), wailsRuntimePlugin()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    port: 5173
  },
  build: {
    outDir: 'dist',
    assetsDir: 'assets',
    rollupOptions: {
      output: {
        manualChunks: {
          'element-plus': ['element-plus'],
          'echarts': ['echarts', 'vue-echarts']
        }
      }
    }
  }
})
