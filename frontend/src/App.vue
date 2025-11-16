<template>
  <div id="app" class="app-container">
    <el-container>
      <el-aside width="240px" class="sidebar">
        <div class="logo">
          <el-icon :size="32"><Monitor /></el-icon>
          <span>ScriptGuard</span>
        </div>
        <el-menu
          :default-active="$route.path"
          router
          class="sidebar-menu"
          background-color="transparent"
        >
          <el-menu-item index="/dashboard">
            <el-icon><DataAnalysis /></el-icon>
            <span>仪表盘</span>
          </el-menu-item>
          <el-menu-item index="/tasks">
            <el-icon><List /></el-icon>
            <span>任务管理</span>
          </el-menu-item>
          <el-menu-item index="/logs">
            <el-icon><Document /></el-icon>
            <span>日志监控</span>
          </el-menu-item>
          <el-menu-item index="/history">
            <el-icon><Clock /></el-icon>
            <span>执行历史</span>
          </el-menu-item>
          <el-menu-item index="/settings">
            <el-icon><Setting /></el-icon>
            <span>系统设置</span>
          </el-menu-item>
        </el-menu>
      </el-aside>

      <el-container>
        <el-main class="main-content">
          <router-view v-slot="{ Component }">
            <transition name="fade" mode="out-in">
              <component :is="Component" />
            </transition>
          </router-view>
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script setup>
import { Monitor, DataAnalysis, List, Document, Clock, Setting } from '@element-plus/icons-vue'
</script>

<style lang="scss">
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  background: #0f172a;
  color: #e2e8f0;
}

.app-container {
  height: 100vh;
  overflow: hidden;
}

.sidebar {
  background: linear-gradient(180deg, #1e293b 0%, #0f172a 100%);
  border-right: 1px solid #334155;

  .logo {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 24px 20px;
    font-size: 20px;
    font-weight: 600;
    color: #6366f1;
    border-bottom: 1px solid #334155;
  }

  .sidebar-menu {
    border: none;
    padding: 12px 8px;

    .el-menu-item {
      border-radius: 8px;
      margin-bottom: 4px;
      color: #94a3b8;

      &:hover {
        background: rgba(99, 102, 241, 0.1);
        color: #6366f1;
      }

      &.is-active {
        background: linear-gradient(90deg, #6366f1 0%, #8b5cf6 100%);
        color: #fff;
      }
    }
  }
}

.main-content {
  background: #0f172a;
  padding: 24px;
  overflow-y: auto;
}

.fade-enter-active, .fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from, .fade-leave-to {
  opacity: 0;
}
</style>
