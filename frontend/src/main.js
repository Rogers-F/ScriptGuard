import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import App from './App.vue'
import router from './router'
import './assets/styles/main.scss'

console.log('🚀 ScriptGuard 正在启动...')

const app = createApp(App)
const pinia = createPinia()

// 注册所有图标
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}

// 全局错误处理
app.config.errorHandler = (err, instance, info) => {
  console.error('❌ 全局错误:', err)
  console.error('📍 错误位置:', info)
  console.error('🔍 组件实例:', instance)
}

// 全局警告处理
app.config.warnHandler = (msg, instance, trace) => {
  console.warn('⚠️ Vue 警告:', msg)
  console.warn('📍 调用栈:', trace)
}

// 路由错误处理
router.onError((error) => {
  console.error('❌ 路由错误:', error)
})

app.use(pinia)
app.use(router)
app.use(ElementPlus, { size: 'default' })

app.mount('#app')

console.log('✅ ScriptGuard 已挂载')
console.log('📍 当前路由:', router.currentRoute.value.path)
