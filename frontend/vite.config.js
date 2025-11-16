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
        // 计算从当前文件到 frontend 目录的相对路径深度
        // 例如: frontend/bindings/scriptguard/backend/app.js -> 需要 3 个 ../
        //      frontend/bindings/github.com/.../models.js -> 需要 7 个 ../
        const parts = id.split(/[\\/]/)
        const bindingsIndex = parts.indexOf('bindings')
        if (bindingsIndex === -1) return null

        // 计算从 bindings 后到文件的深度
        const depth = parts.length - bindingsIndex - 1
        const relativePath = '../'.repeat(depth) + 'src/wails-runtime.js'

        const replaced = code.replace(
          /from\s+["']@wailsio\/runtime["']/g,
          `from "${relativePath}"`
        )

        console.log(`🔧 替换 Wails runtime (深度=${depth}):`, id.substring(id.indexOf('bindings')))

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
