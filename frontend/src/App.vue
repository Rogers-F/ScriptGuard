<template>
  <div id="app" class="app-layout">
    <el-container class="h-full">
      <el-aside width="260px" class="app-sidebar glass-sidebar">
        <div class="logo-area">
          <div class="logo-icon">
            <span class="initial">S</span>
          </div>
          <span class="app-title">ScriptGuard</span>
        </div>

        <el-menu
          :default-active="$route.path"
          router
          class="nav-menu"
          :collapse-transition="false"
        >
          <div class="menu-section">
            <div class="menu-label">{{ t.menu.analytics }}</div>
            <el-menu-item index="/dashboard">
              <el-icon><DataAnalysis /></el-icon>
              <span>{{ t.menu.dashboard }}</span>
            </el-menu-item>
            <el-menu-item index="/history">
              <el-icon><Clock /></el-icon>
              <span>{{ t.menu.history }}</span>
            </el-menu-item>
          </div>

          <div class="menu-section">
            <div class="menu-label">{{ t.menu.scheduling }}</div>
            <el-menu-item index="/tasks">
              <el-icon><List /></el-icon>
              <span>{{ t.menu.tasks }}</span>
            </el-menu-item>
            <el-menu-item index="/logs">
              <el-icon><Document /></el-icon>
              <span>{{ t.menu.logs }}</span>
            </el-menu-item>
          </div>

          <div class="menu-spacer"></div>

          <div class="menu-section">
            <el-menu-item index="/settings">
              <el-icon><Setting /></el-icon>
              <span>{{ t.menu.settings }}</span>
            </el-menu-item>
          </div>
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
import { computed } from 'vue'
import { Monitor, DataAnalysis, List, Document, Clock, Setting } from '@element-plus/icons-vue'
import { useLanguageStore } from '@/stores/language'

const langStore = useLanguageStore()
const t = computed(() => langStore.t)
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
  display: flex;
  flex-direction: column;
  padding: 24px 16px;

  .logo-area {
    height: 60px;
    display: flex;
    align-items: center;
    padding: 0 16px;
    margin-bottom: 24px;

    .logo-icon {
      width: 32px;
      height: 32px;
      background: var(--color-primary);
      color: white;
      border-radius: 8px;
      display: flex;
      align-items: center;
      justify-content: center;
      margin-right: 12px;
      font-family: var(--font-serif);
      font-weight: bold;
      font-size: 18px;
    }

    .app-title {
      font-family: var(--font-serif);
      font-size: 20px;
      font-weight: 600;
      color: var(--text-primary);
    }
  }

  .nav-menu {
    border-right: none;
    background: transparent;
    flex: 1;

    .menu-section {
      margin-bottom: 24px;
    }

    .menu-label {
      font-size: 11px;
      font-weight: 600;
      color: var(--text-tertiary);
      padding: 0 16px;
      margin-bottom: 8px;
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }

    .el-menu-item {
      height: 40px;
      line-height: 40px;
      border-radius: 8px;
      margin-bottom: 4px;
      color: var(--text-secondary);
      font-weight: 500;
      font-family: var(--font-sans);

      .el-icon {
        color: var(--text-tertiary);
        margin-right: 10px;
      }

      &:hover {
        background-color: rgba(0,0,0,0.03);
        color: var(--text-primary);
        .el-icon { color: var(--text-secondary); }
      }

      &.is-active {
        background-color: rgba(217, 119, 6, 0.08); /* Very subtle orange tint */
        color: var(--color-primary);
        font-weight: 600;

        .el-icon { color: var(--color-primary); }
      }
    }
  }

  .menu-spacer { flex: 1; }
}

.app-main {
  padding: 0;
  overflow-y: auto;
}

.fade-enter-active, .fade-leave-active { transition: opacity 0.2s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
