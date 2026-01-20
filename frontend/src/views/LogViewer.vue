<template>
  <div class="page-container log-viewer-page">
    <div class="page-header compact">
      <div class="left">
        <h1>日志监控</h1>
      </div>
      <div class="controls">
        <el-select v-model="selectedTask" placeholder="选择任务" size="default" clearable class="w-48">
          <el-option label="全部任务" value="" />
          <el-option v-for="task in taskStore.tasks" :key="task.id" :label="task.name" :value="task.id" />
        </el-select>
        <el-input v-model="searchText" placeholder="搜索日志..." size="default" class="w-64" clearable>
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <el-divider direction="vertical" />
        <el-button size="default" :type="autoScroll ? 'primary' : 'default'" @click="autoScroll = !autoScroll">
          {{ autoScroll ? '暂停滚动' : '自动滚动' }}
        </el-button>
        <el-button size="default" @click="clearLogs">清空</el-button>
        <el-button size="default" type="primary" @click="exportLogs" :loading="isExporting">
          导出日志
        </el-button>
      </div>
    </div>

    <div class="terminal-window">
      <div class="terminal-bar">
        <div class="dots">
          <span></span><span></span><span></span>
        </div>
        <div class="status">{{ filteredLogs.length }} lines</div>
      </div>

      <div class="terminal-content" ref="logBodyRef">
        <!-- Error State -->
        <div v-if="loadError" class="empty-terminal error-state">
          <el-icon :size="48" color="#ef4444"><WarningFilled /></el-icon>
          <p>{{ loadError }}</p>
          <el-button type="primary" size="small" @click="retryLoad">重试</el-button>
          <p class="retry-info" v-if="failureCount > 0">
            已失败 {{ failureCount }} 次，下次重试间隔 {{ Math.round(getPollingInterval() / 1000) }} 秒
          </p>
        </div>

        <div v-else-if="filteredLogs.length === 0" class="empty-terminal">
          No logs to display
        </div>

        <div
          v-else
          v-for="(log, index) in filteredLogs"
          :key="log.id || index"
          class="log-row"
          :class="log.level"
        >
          <span class="ts">{{ formatTime(log.timestamp) }}</span>
          <span class="lvl">{{ (log.level || 'INFO').toUpperCase() }}</span>
          <span class="msg">{{ log.content }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { Search, WarningFilled } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useTaskStore } from '@/stores/task'
import api from '@/api'

const route = useRoute()
const taskStore = useTaskStore()

const selectedTask = ref('')
const selectedExecution = ref('')
const searchText = ref('')
const autoScroll = ref(true)
const logs = ref([])
const logBodyRef = ref(null)
const isLoading = ref(false)
const isExporting = ref(false)
const loadError = ref(null)
const failureCount = ref(0)
const maxRetryInterval = 30000
const baseInterval = 2000

const filteredLogs = computed(() => {
  let result = logs.value

  if (selectedTask.value) {
    result = result.filter(log => log.task_id === selectedTask.value)
  }

  if (searchText.value) {
    const search = searchText.value.toLowerCase()
    result = result.filter(log => log.content.toLowerCase().includes(search))
  }

  return result
})

watch(filteredLogs, async () => {
  if (autoScroll.value) {
    await nextTick()
    scrollToBottom()
  }
})

function scrollToBottom() {
  if (logBodyRef.value) {
    logBodyRef.value.scrollTop = logBodyRef.value.scrollHeight
  }
}

function formatTime(time) {
  return new Date(time).toLocaleTimeString('zh-CN', { hour12: false, timeZone: 'Asia/Shanghai' })
}

function clearLogs() {
  logs.value = []
}

async function loadLogs() {
  if (isLoading.value) return

  isLoading.value = true
  try {
    const result = selectedExecution.value
      ? await api.getLogs(selectedExecution.value, '', 1000)
      : await api.getLogs('', selectedTask.value, 1000)
    logs.value = result || []
    loadError.value = null
    failureCount.value = 0
  } catch (error) {
    console.error('加载日志失败:', error)
    loadError.value = error.message || '加载失败'
    failureCount.value++
  } finally {
    isLoading.value = false
  }
}

function getPollingInterval() {
  if (failureCount.value === 0) return baseInterval
  return Math.min(Math.pow(2, failureCount.value) * baseInterval, maxRetryInterval)
}

function retryLoad() {
  failureCount.value = 0
  loadError.value = null
  loadLogs()
}

async function exportLogs() {
  isExporting.value = true
  try {
    const frontendLogs = loadError.value
      ? `前端错误: ${loadError.value}\n失败次数: ${failureCount.value}\n`
      : ''

    const savedPath = await api.exportDebugLogs(frontendLogs)
    if (savedPath) {
      ElMessage.success(`日志已导出到: ${savedPath}`)
    }
  } catch (error) {
    console.error('导出日志失败:', error)
    ElMessage.error(`导出失败: ${error.message || '未知错误'}`)
  } finally {
    isExporting.value = false
  }
}

let pollingTimer = null

function startPolling() {
  stopPolling()
  scheduleNextPoll()
}

function stopPolling() {
  if (pollingTimer) {
    clearTimeout(pollingTimer)
    pollingTimer = null
  }
}

function scheduleNextPoll() {
  const interval = getPollingInterval()
  pollingTimer = setTimeout(async () => {
    await loadLogs()
    scheduleNextPoll()
  }, interval)
}

onMounted(async () => {
  if (route.query.execution) {
    selectedExecution.value = route.query.execution
  }

  await taskStore.loadTasks()
  await loadLogs()
  startPolling()
})

onUnmounted(() => {
  stopPolling()
})

watch(selectedTask, () => {
  selectedExecution.value = ''
  loadLogs()
})
</script>

<style lang="scss" scoped>
.log-viewer-page {
  height: calc(100vh - 48px);
  display: flex;
  flex-direction: column;
}

.page-header.compact {
  margin-bottom: 16px;
  .controls {
    display: flex;
    gap: 12px;
    align-items: center;
  }
}

.terminal-window {
  flex: 1;
  background-color: #0d0d0d;
  border: 1px solid var(--border-light);
  border-radius: 8px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  font-family: 'JetBrains Mono', 'Fira Code', 'Consolas', monospace;

  .terminal-bar {
    height: 36px;
    background-color: var(--bg-secondary);
    border-bottom: 1px solid var(--border-light);
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 16px;

    .dots {
      display: flex;
      gap: 6px;
      span {
        width: 10px;
        height: 10px;
        border-radius: 50%;
        background-color: var(--border-hover);
        &:nth-child(1) { background-color: #ef4444; }
        &:nth-child(2) { background-color: #f59e0b; }
        &:nth-child(3) { background-color: #10b981; }
      }
    }

    .status {
      font-size: 12px;
      color: var(--text-tertiary);
    }
  }

  .terminal-content {
    flex: 1;
    overflow-y: auto;
    padding: 12px;
    font-size: 13px;
    line-height: 1.5;

    .log-row {
      display: flex;
      gap: 12px;
      padding: 2px 0;
      color: #d4d4d8;

      &:hover {
        background-color: rgba(255, 255, 255, 0.03);
      }

      .ts { color: #52525b; min-width: 70px; }
      .lvl {
        min-width: 50px;
        font-weight: bold;
      }
      .msg { white-space: pre-wrap; word-break: break-all; }

      &.info .lvl { color: #3b82f6; }
      &.error .lvl, &.stderr .lvl { color: #ef4444; }
      &.warning .lvl { color: #f59e0b; }
      &.success .lvl { color: #10b981; }
      &.stdout .lvl { color: #a1a1aa; }
    }

    .empty-terminal {
      height: 100%;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      color: var(--text-tertiary);

      &.error-state {
        color: #ef4444;

        p {
          margin: 16px 0 8px;
        }

        .retry-info {
          font-size: 12px;
          color: var(--text-tertiary);
          margin-top: 8px;
        }
      }
    }
  }
}

.w-48 { width: 12rem; }
.w-64 { width: 16rem; }
</style>
