<template>
  <div id="app" class="app-layout">
    <el-container class="h-full">
      <el-aside width="240px" class="app-sidebar">
        <div class="logo-area">
          <div class="logo-icon">
            <el-icon :size="20"><Monitor /></el-icon>
          </div>
          <span class="app-title">ScriptGuard</span>
        </div>

        <el-menu
          :default-active="$route.path"
          router
          class="nav-menu"
          :collapse-transition="false"
        >
          <div class="menu-label">MENU</div>
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

          <div class="menu-spacer"></div>

          <div class="menu-label">SYSTEM</div>
          <el-menu-item index="/settings">
            <el-icon><Setting /></el-icon>
            <span>系统设置</span>
          </el-menu-item>
        </el-menu>
      </el-aside>

      <el-main class="app-main">
        <router-view v-slot="{ Component }">
          <transition name="fade" mode="out-in">
            <component :is="Component" />
          </transition>
        </router-view>
      </el-main>
    </el-container>
  </div>
</template>

<script setup>
import { Monitor, DataAnalysis, List, Document, Clock, Setting } from '@element-plus/icons-vue'
</script>

<style lang="scss">
@import '@/assets/styles/main.scss';

.app-layout {
  height: 100vh;
  background-color: var(--bg-body);
}

.h-full {
  height: 100%;
}

.app-sidebar {
  background-color: var(--bg-secondary);
  border-right: 1px solid var(--border-light);
  display: flex;
  flex-direction: column;

  .logo-area {
    height: 64px;
    display: flex;
    align-items: center;
    padding: 0 24px;
    border-bottom: 1px solid var(--border-light);

    .logo-icon {
      width: 32px;
      height: 32px;
      background: rgba(59, 130, 246, 0.1);
      color: var(--color-primary);
      border-radius: 6px;
      display: flex;
      align-items: center;
      justify-content: center;
      margin-right: 12px;
    }

    .app-title {
      font-size: 16px;
      font-weight: 600;
      color: var(--text-primary);
      letter-spacing: -0.02em;
    }
  }

  .nav-menu {
    border-right: none;
    background: transparent;
    padding: 24px 12px;
    flex: 1;

    .menu-label {
      font-size: 11px;
      font-weight: 600;
      color: var(--text-tertiary);
      padding: 0 12px;
      margin-bottom: 8px;
      margin-top: 8px;
    }

    .el-menu-item {
      height: 40px;
      line-height: 40px;
      border-radius: 6px;
      margin-bottom: 2px;
      color: var(--text-secondary);

      &:hover {
        background-color: var(--bg-tertiary);
        color: var(--text-primary);
      }

      &.is-active {
        background-color: rgba(59, 130, 246, 0.1);
        color: var(--color-primary);
        font-weight: 500;
      }
    }
  }

  .menu-spacer {
    height: 24px;
  }
}

.app-main {
  padding: 0;
  background-color: var(--bg-body);
  overflow-y: auto;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
