<template>
  <div class="page-container log-viewer-page">
    <div class="page-header compact">
      <div class="left">
        <h1>{{ t.logs.title }}</h1>
      </div>
      <div class="controls glass-panel">
        <el-select v-model="selectedTask" :placeholder="t.logs.allTasks" size="default" clearable class="w-48">
          <el-option v-for="task in taskStore.tasks" :key="task.id" :label="task.name" :value="task.id" />
        </el-select>
        <el-input v-model="searchText" :placeholder="t.common.search" size="default" class="w-64" clearable>
          <template #prefix><el-icon><Search /></el-icon></template>
        </el-input>
        <div class="divider"></div>
        <el-button size="default" :type="autoScroll ? 'primary' : 'default'" @click="autoScroll = !autoScroll">
          {{ autoScroll ? t.logs.autoscrollOn : t.logs.autoscrollOff }}
        </el-button>
        <el-button size="default" @click="clearLogs">{{ t.logs.clear }}</el-button>
        <el-button size="default" @click="exportLogs" :loading="isExporting">
          <el-icon><Download /></el-icon>
        </el-button>
      </div>
    </div>

    <div class="terminal-window">
      <div class="terminal-content" ref="logBodyRef">
        <div v-if="loadError" class="empty-terminal error-state">
          <el-icon :size="48" color="#dc2626"><WarningFilled /></el-icon>
          <p>{{ loadError }}</p>
          <el-button link type="primary" @click="retryLoad">{{ t.logs.retry }}</el-button>
        </div>

        <div v-else-if="filteredLogs.length === 0" class="empty-terminal">
          <span>{{ t.logs.waitingForLogs }}</span>
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
import { Search, WarningFilled, Download } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useTaskStore } from '@/stores/task'
import { useLanguageStore } from '@/stores/language'
import api from '@/api'

const route = useRoute()
const taskStore = useTaskStore()
const langStore = useLanguageStore()
const t = computed(() => langStore.t)

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
  if (selectedTask.value) result = result.filter(log => log.task_id === selectedTask.value)
  if (searchText.value) result = result.filter(log => log.content.toLowerCase().includes(searchText.value.toLowerCase()))
  return result
})

watch(filteredLogs, async () => { if (autoScroll.value) { await nextTick(); scrollToBottom() } })
function scrollToBottom() { if (logBodyRef.value) logBodyRef.value.scrollTop = logBodyRef.value.scrollHeight }
function formatTime(time) { return new Date(time).toLocaleTimeString('zh-CN', { hour12: false }) }
function clearLogs() {
  logs.value = []
  stopPolling() // 停止轮询，避免清空后立即重新加载
  autoScroll.value = false
}

async function loadLogs() {
  if (isLoading.value) return
  isLoading.value = true
  try {
    const result = selectedExecution.value ? await api.getLogs(selectedExecution.value, '', 1000) : await api.getLogs('', selectedTask.value, 1000)
    logs.value = result || []
    loadError.value = null
    failureCount.value = 0
  } catch (error) { loadError.value = error.message; failureCount.value++ } finally { isLoading.value = false }
}

function getPollingInterval() {
  if (failureCount.value === 0) return baseInterval
  return Math.min(Math.pow(2, failureCount.value) * baseInterval, maxRetryInterval)
}

function retryLoad() { failureCount.value = 0; loadError.value = null; loadLogs() }

async function exportLogs() {
  isExporting.value = true
  try {
    const savedPath = await api.exportDebugLogs("")
    if (savedPath) ElMessage.success(`${t.value.logs.savedTo} ${savedPath}`)
  } catch (error) { ElMessage.error(error.message) } finally { isExporting.value = false }
}

let pollingTimer = null
function startPolling() { stopPolling(); scheduleNextPoll() }
function stopPolling() { if (pollingTimer) { clearTimeout(pollingTimer); pollingTimer = null } }
function scheduleNextPoll() {
  const interval = getPollingInterval()
  pollingTimer = setTimeout(async () => { await loadLogs(); scheduleNextPoll() }, interval)
}

onMounted(async () => {
  if (route.query.execution) selectedExecution.value = route.query.execution
  await taskStore.loadTasks()
  await loadLogs()
  startPolling()
})
onUnmounted(() => stopPolling())
watch(selectedTask, () => { selectedExecution.value = ''; loadLogs() })
</script>

<style lang="scss" scoped>
.log-viewer-page {
  height: calc(100vh - 40px);
  display: flex;
  flex-direction: column;
}

.page-header.compact {
  margin-bottom: 20px;
  .controls {
    display: flex;
    gap: 12px;
    align-items: center;
    padding: 8px 16px;
    border: none;

    .divider { width: 1px; height: 20px; background: var(--border-light); margin: 0 8px; }
  }
}

.terminal-window {
  flex: 1;
  background-color: #1e1e1e;
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  font-family: var(--font-mono);
  box-shadow: var(--shadow-glass);

  .terminal-content {
    flex: 1;
    overflow-y: auto;
    padding: 24px;
    font-size: 13px;
    line-height: 1.6;

    .log-row {
      display: flex;
      gap: 16px;
      color: #a3a3a3;

      .ts { color: #525252; min-width: 80px; user-select: none; }
      .lvl { min-width: 40px; font-weight: 600; user-select: none; }
      .msg { white-space: pre-wrap; word-break: break-all; color: #d4d4d4; }

      &.info .lvl { color: #60a5fa; }
      &.error .lvl { color: #f87171; }
      &.warning .lvl { color: #fbbf24; }
      &.success .lvl { color: #34d399; }
    }

    .empty-terminal {
      height: 100%;
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      color: #525252;
      gap: 12px;
    }
  }
}

.w-48 { width: 12rem; }
.w-64 { width: 16rem; }
</style>
