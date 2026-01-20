<template>
  <div class="log-viewer">
    <div class="page-header">
      <h1>日志监控</h1>
      <div class="header-actions">
        <el-select v-model="selectedTask" placeholder="选择任务" style="width: 250px" clearable>
          <el-option label="全部任务" value="" />
          <el-option
            v-for="task in taskStore.tasks"
            :key="task.id"
            :label="task.name"
            :value="task.id"
          />
        </el-select>
        <el-input
          v-model="searchText"
          placeholder="搜索日志..."
          style="width: 300px"
          clearable
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
        <el-button :icon="autoScroll ? VideoPause : VideoPlay" @click="autoScroll = !autoScroll">
          {{ autoScroll ? '暂停滚动' : '自动滚动' }}
        </el-button>
        <el-button :icon="Delete" @click="clearLogs">清空</el-button>
      </div>
    </div>

    <!-- 终端风格日志显示 -->
    <div class="terminal-container" ref="terminalRef">
      <div class="terminal-header">
        <div class="terminal-title">
          <span class="dot red"></span>
          <span class="dot yellow"></span>
          <span class="dot green"></span>
          <span class="title-text">ScriptGuard Terminal</span>
        </div>
        <div class="terminal-info">
          <el-tag size="small">{{ filteredLogs.length }} 条日志</el-tag>
        </div>
      </div>

      <div class="terminal-body" ref="logBodyRef">
        <div
          v-for="log in filteredLogs"
          :key="log.id || log.timestamp"
          class="log-line"
          :class="log.level"
        >
          <span class="log-time">{{ formatTime(log.timestamp) }}</span>
          <span class="log-level" :class="log.level">{{ log.level.toUpperCase() }}</span>
          <span class="log-content">{{ log.content }}</span>
        </div>

        <!-- 错误状态 -->
        <div v-if="loadError" class="error-state">
          <el-icon :size="64" color="#ef4444"><WarningFilled /></el-icon>
          <p>{{ loadError }}</p>
          <el-button type="primary" @click="retryLoad">重试</el-button>
          <p class="retry-info" v-if="failureCount > 0">
            已失败 {{ failureCount }} 次，下次重试间隔 {{ Math.round(getPollingInterval() / 1000) }} 秒
          </p>
        </div>

        <div v-else-if="filteredLogs.length === 0" class="empty-state">
          <el-icon :size="64"><Document /></el-icon>
          <p>暂无日志</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { Search, VideoPlay, VideoPause, Delete, Document, WarningFilled } from '@element-plus/icons-vue'
import { useTaskStore } from '@/stores/task'
import api from '@/api'

const route = useRoute()
const taskStore = useTaskStore()

const selectedTask = ref('')
const selectedExecution = ref('')
const searchText = ref('')
const autoScroll = ref(true)
const logs = ref([])
const terminalRef = ref(null)
const logBodyRef = ref(null)
const isLoading = ref(false)
const loadError = ref(null)
const failureCount = ref(0)
const maxRetryInterval = 30000 // 最大退避间隔 30 秒
const baseInterval = 2000 // 基础轮询间隔 2 秒

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

// 加载日志
async function loadLogs() {
  if (isLoading.value) return

  isLoading.value = true
  try {
    // 优先按 execution_id 查询，其次按 task_id
    const result = selectedExecution.value
      ? await api.getLogs(selectedExecution.value, '', 1000)
      : await api.getLogs('', selectedTask.value, 1000)
    logs.value = result || []
    // 成功后重置错误状态和失败计数
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

// 计算退避后的轮询间隔
function getPollingInterval() {
  if (failureCount.value === 0) return baseInterval
  // 指数退避：2^failureCount * baseInterval，最大 maxRetryInterval
  return Math.min(Math.pow(2, failureCount.value) * baseInterval, maxRetryInterval)
}

// 手动重试
function retryLoad() {
  failureCount.value = 0
  loadError.value = null
  loadLogs()
}

// 日志轮询（使用 setTimeout 实现动态间隔）
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
    scheduleNextPoll() // 递归调度下一次轮询
  }, interval)
}

onMounted(async () => {
  // 解析 URL query 参数（从执行历史页跳转过来时带 execution 参数）
  if (route.query.execution) {
    selectedExecution.value = route.query.execution
  }

  await taskStore.loadTasks()
  await loadLogs()

  // 启动带退避机制的轮询
  startPolling()
})

onUnmounted(() => {
  stopPolling()
})

// 监听选中任务变化，重新加载日志
watch(selectedTask, () => {
  // SG-017': 选择任务时清空 execution 过滤，避免冲突
  selectedExecution.value = ''
  loadLogs()
})
</script>

<style lang="scss" scoped>
.log-viewer {
  height: calc(100vh - 100px);
  display: flex;
  flex-direction: column;

  .page-header {
    margin-bottom: 20px;

    h1 {
      font-size: 28px;
      font-weight: 600;
      margin-bottom: 16px;
      background: linear-gradient(90deg, #6366f1, #8b5cf6);
      -webkit-background-clip: text;
      -webkit-text-fill-color: transparent;
    }

    .header-actions {
      display: flex;
      gap: 12px;
      flex-wrap: wrap;
    }
  }

  .terminal-container {
    flex: 1;
    background: #0f172a;
    border: 1px solid #334155;
    border-radius: 12px;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    box-shadow: 0 8px 16px rgba(0, 0, 0, 0.4);

    .terminal-header {
      background: linear-gradient(180deg, #1e293b 0%, #0f172a 100%);
      padding: 12px 20px;
      border-bottom: 1px solid #334155;
      display: flex;
      justify-content: space-between;
      align-items: center;

      .terminal-title {
        display: flex;
        align-items: center;
        gap: 8px;

        .dot {
          width: 12px;
          height: 12px;
          border-radius: 50%;

          &.red { background: #ef4444; }
          &.yellow { background: #f59e0b; }
          &.green { background: #10b981; }
        }

        .title-text {
          margin-left: 12px;
          font-size: 14px;
          font-weight: 500;
          color: #94a3b8;
        }
      }
    }

    .terminal-body {
      flex: 1;
      padding: 16px;
      font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
      font-size: 13px;
      line-height: 1.6;
      overflow-y: auto;
      background: #0f172a;

      .log-line {
        display: flex;
        gap: 12px;
        padding: 4px 0;
        border-bottom: 1px solid rgba(255, 255, 255, 0.03);

        &:hover {
          background: rgba(99, 102, 241, 0.05);
        }

        .log-time {
          color: #64748b;
          min-width: 80px;
        }

        .log-level {
          min-width: 60px;
          font-weight: 600;

          &.info { color: #3b82f6; }
          &.success { color: #10b981; }
          &.warning { color: #f59e0b; }
          &.error, &.stderr { color: #ef4444; }
          &.stdout { color: #94a3b8; }
        }

        .log-content {
          flex: 1;
          color: #e2e8f0;
          word-break: break-all;
        }
      }

      .empty-state,
      .error-state {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        height: 100%;
        color: #64748b;

        p {
          margin-top: 16px;
          font-size: 16px;
        }
      }

      .error-state {
        color: #ef4444;

        .el-button {
          margin-top: 16px;
        }

        .retry-info {
          margin-top: 8px;
          font-size: 12px;
          color: #94a3b8;
        }
      }
    }
  }
}
</style>
