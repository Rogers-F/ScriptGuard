<template>
  <div class="page-container history-page">
    <div class="page-header">
      <div class="header-left">
        <h1>执行历史</h1>
        <p>查看所有任务的执行记录</p>
      </div>
      <div class="header-right">
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          range-separator="-"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          style="width: 280px"
          @change="loadExecutions"
        />
        <el-select v-model="selectedTask" placeholder="选择任务" style="width: 180px" clearable @change="loadExecutions">
          <el-option label="全部任务" value="" />
          <el-option
            v-for="task in taskStore.tasks"
            :key="task.id"
            :label="task.name"
            :value="task.id"
          />
        </el-select>
      </div>
    </div>

    <!-- Stats Row -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-label">总执行次数</div>
        <div class="stat-value">{{ stats.total }}</div>
      </div>
      <div class="stat-card success">
        <div class="stat-label">成功</div>
        <div class="stat-value">{{ stats.success }}</div>
        <div class="stat-percent">{{ stats.successRate }}%</div>
      </div>
      <div class="stat-card failed">
        <div class="stat-label">失败</div>
        <div class="stat-value">{{ stats.failed }}</div>
        <div class="stat-percent">{{ stats.failureRate }}%</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">平均耗时</div>
        <div class="stat-value">{{ stats.avgDuration }}<span class="unit">s</span></div>
      </div>
    </div>

    <!-- Charts -->
    <div class="charts-row">
      <div class="card-panel chart-panel">
        <div class="panel-header">
          <h3>执行趋势</h3>
        </div>
        <v-chart :option="trendChartOption" autoresize class="chart-container" />
      </div>

      <div class="card-panel chart-panel">
        <div class="panel-header">
          <h3>耗时分布</h3>
        </div>
        <v-chart :option="durationChartOption" autoresize class="chart-container" />
      </div>
    </div>

    <!-- Timeline -->
    <div class="card-panel timeline-section">
      <div class="panel-header">
        <h3>执行记录</h3>
        <el-button text bg size="small" @click="loadExecutions">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>

      <el-timeline v-if="executions.length > 0">
        <el-timeline-item
          v-for="exec in executions"
          :key="exec.id"
          :timestamp="formatTime(exec.start_time)"
          placement="top"
          :color="getTimelineColor(exec.status)"
        >
          <div class="execution-card">
            <div class="execution-header">
              <div class="task-name">{{ getTaskName(exec.task_id) }}</div>
              <el-tag :type="getStatusType(exec.status)" size="small" effect="dark">
                {{ getStatusText(exec.status) }}
              </el-tag>
            </div>

            <div class="execution-details">
              <div class="detail-item">
                <span class="label">ID</span>
                <span class="value mono">{{ exec.id.substring(0, 8) }}...</span>
              </div>
              <div class="detail-item">
                <span class="label">开始</span>
                <span class="value">{{ formatFullTime(exec.start_time) }}</span>
              </div>
              <div class="detail-item">
                <span class="label">耗时</span>
                <span class="value mono">{{ (exec.duration_ms / 1000).toFixed(2) }}s</span>
              </div>
              <div class="detail-item">
                <span class="label">退出码</span>
                <span class="value mono">{{ exec.exit_code }}</span>
              </div>
            </div>

            <div v-if="exec.error_message" class="error-message">
              {{ exec.error_message }}
            </div>

            <div class="execution-actions">
              <el-button size="small" text type="primary" @click="viewLogs(exec.id)">
                查看日志
              </el-button>
            </div>
          </div>
        </el-timeline-item>
      </el-timeline>

      <el-empty v-else description="暂无执行记录" />
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, BarChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import { Refresh } from '@element-plus/icons-vue'
import { useTaskStore } from '@/stores/task'

use([CanvasRenderer, LineChart, BarChart, GridComponent, TooltipComponent, LegendComponent])

const router = useRouter()
const taskStore = useTaskStore()

const selectedTask = ref('')
const dateRange = ref([])
const executions = ref([])

const stats = reactive({
  total: 0,
  success: 0,
  failed: 0,
  successRate: 0,
  failureRate: 0,
  avgDuration: 0
})

const trendChartOption = computed(() => {
  const dateMap = new Map()
  executions.value.forEach(exec => {
    const date = new Date(exec.start_time).toLocaleDateString('zh-CN')
    if (!dateMap.has(date)) {
      dateMap.set(date, { success: 0, failed: 0 })
    }
    const stat = dateMap.get(date)
    if (exec.status === 'success') stat.success++
    else if (exec.status === 'failed') stat.failed++
  })

  const dates = Array.from(dateMap.keys()).slice(-7)
  const successData = dates.map(date => dateMap.get(date)?.success || 0)
  const failedData = dates.map(date => dateMap.get(date)?.failed || 0)

  return {
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'axis',
      backgroundColor: '#1f1f1f',
      borderColor: '#333',
      textStyle: { color: '#e5e5e5' }
    },
    grid: { left: 0, right: 0, bottom: 0, top: 20, containLabel: true },
    xAxis: {
      type: 'category',
      data: dates,
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: { color: '#71717a' }
    },
    yAxis: {
      type: 'value',
      splitLine: { lineStyle: { color: '#27272a' } },
      axisLabel: { color: '#71717a' }
    },
    series: [
      {
        name: '成功',
        type: 'line',
        data: successData,
        smooth: true,
        symbol: 'none',
        lineStyle: { color: '#10b981', width: 2 },
        itemStyle: { color: '#10b981' }
      },
      {
        name: '失败',
        type: 'line',
        data: failedData,
        smooth: true,
        symbol: 'none',
        lineStyle: { color: '#ef4444', width: 2 },
        itemStyle: { color: '#ef4444' }
      }
    ]
  }
})

const durationChartOption = computed(() => {
  const ranges = ['0-10s', '10-30s', '30-60s', '60-120s', '>120s']
  const counts = [0, 0, 0, 0, 0]

  executions.value.forEach(exec => {
    const duration = exec.duration_ms / 1000
    if (duration <= 10) counts[0]++
    else if (duration <= 30) counts[1]++
    else if (duration <= 60) counts[2]++
    else if (duration <= 120) counts[3]++
    else counts[4]++
  })

  return {
    backgroundColor: 'transparent',
    tooltip: {
      trigger: 'axis',
      backgroundColor: '#1f1f1f',
      borderColor: '#333',
      textStyle: { color: '#e5e5e5' }
    },
    grid: { left: 0, right: 0, bottom: 0, top: 20, containLabel: true },
    xAxis: {
      type: 'category',
      data: ranges,
      axisLine: { show: false },
      axisTick: { show: false },
      axisLabel: { color: '#71717a' }
    },
    yAxis: {
      type: 'value',
      splitLine: { lineStyle: { color: '#27272a' } },
      axisLabel: { color: '#71717a' }
    },
    series: [
      {
        name: '执行次数',
        type: 'bar',
        data: counts,
        itemStyle: { color: '#3b82f6', borderRadius: [4, 4, 0, 0] }
      }
    ]
  }
})

onMounted(async () => {
  await taskStore.loadTasks()
  await loadExecutions()
})

async function loadExecutions() {
  await taskStore.loadExecutions(selectedTask.value, 100)
  let result = taskStore.executions

  if (dateRange.value && dateRange.value.length === 2) {
    const startDate = new Date(dateRange.value[0])
    startDate.setHours(0, 0, 0, 0)
    const endDate = new Date(dateRange.value[1])
    endDate.setHours(23, 59, 59, 999)

    result = result.filter(e => {
      const execTime = new Date(e.start_time)
      return execTime >= startDate && execTime <= endDate
    })
  }

  executions.value = result

  stats.total = executions.value.length
  stats.success = executions.value.filter(e => e.status === 'success').length
  stats.failed = executions.value.filter(e => e.status === 'failed').length
  stats.successRate = stats.total > 0 ? Math.round((stats.success / stats.total) * 100) : 0
  stats.failureRate = stats.total > 0 ? Math.round((stats.failed / stats.total) * 100) : 0

  const durations = executions.value.filter(e => e.duration_ms > 0).map(e => e.duration_ms)
  stats.avgDuration = durations.length > 0
    ? (durations.reduce((a, b) => a + b, 0) / durations.length / 1000).toFixed(2)
    : 0
}

function getTaskName(taskId) {
  const task = taskStore.tasks.find(t => t.id === taskId)
  return task ? task.name : '未知任务'
}

function formatTime(time) {
  return new Date(time).toLocaleString('zh-CN', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    timeZone: 'Asia/Shanghai'
  })
}

function formatFullTime(time) {
  return new Date(time).toLocaleString('zh-CN', { timeZone: 'Asia/Shanghai' })
}

function getStatusType(status) {
  const map = { success: 'success', failed: 'danger', running: 'warning' }
  return map[status] || 'info'
}

function getStatusText(status) {
  const map = { success: '成功', failed: '失败', running: '运行中' }
  return map[status] || status
}

function getTimelineColor(status) {
  const map = { success: '#10b981', failed: '#ef4444', running: '#f59e0b' }
  return map[status] || '#64748b'
}

function viewLogs(executionId) {
  router.push(`/logs?execution=${executionId}`)
}
</script>

<style lang="scss" scoped>
.history-page {
  .page-header {
    .header-right {
      display: flex;
      gap: 12px;
    }
  }

  .stats-row {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 20px;
    margin-bottom: 24px;

    @media (max-width: 1000px) {
      grid-template-columns: repeat(2, 1fr);
    }

    .stat-card {
      background: var(--bg-secondary);
      border: 1px solid var(--border-light);
      border-radius: 8px;
      padding: 16px 20px;

      &.success { border-left: 3px solid var(--color-success); }
      &.failed { border-left: 3px solid var(--color-danger); }

      .stat-label {
        font-size: 12px;
        color: var(--text-secondary);
        margin-bottom: 4px;
      }

      .stat-value {
        font-size: 24px;
        font-weight: 600;
        color: var(--text-primary);

        .unit {
          font-size: 14px;
          color: var(--text-secondary);
          margin-left: 2px;
          font-weight: normal;
        }
      }

      .stat-percent {
        font-size: 12px;
        color: var(--text-tertiary);
        margin-top: 2px;
      }
    }
  }

  .charts-row {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 20px;
    margin-bottom: 24px;

    @media (max-width: 900px) {
      grid-template-columns: 1fr;
    }

    .chart-panel {
      height: 280px;
      display: flex;
      flex-direction: column;

      .panel-header {
        margin-bottom: 12px;

        h3 {
          font-size: 15px;
          font-weight: 500;
          margin: 0;
        }
      }

      .chart-container {
        flex: 1;
        width: 100%;
      }
    }
  }

  .timeline-section {
    .panel-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 20px;

      h3 {
        font-size: 15px;
        font-weight: 500;
        margin: 0;
      }
    }

    .execution-card {
      background: var(--bg-tertiary);
      border: 1px solid var(--border-light);
      border-radius: 6px;
      padding: 16px;

      .execution-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 12px;

        .task-name {
          font-weight: 500;
          color: var(--text-primary);
        }
      }

      .execution-details {
        display: grid;
        grid-template-columns: repeat(4, 1fr);
        gap: 12px;
        margin-bottom: 12px;
        font-size: 13px;

        .detail-item {
          .label {
            color: var(--text-tertiary);
            font-size: 11px;
            text-transform: uppercase;
            display: block;
            margin-bottom: 2px;
          }

          .value {
            color: var(--text-secondary);

            &.mono {
              font-family: monospace;
            }
          }
        }
      }

      .error-message {
        background: rgba(239, 68, 68, 0.1);
        border: 1px solid rgba(239, 68, 68, 0.2);
        border-radius: 4px;
        padding: 8px 12px;
        margin-bottom: 12px;
        color: #ef4444;
        font-size: 13px;
      }

      .execution-actions {
        display: flex;
        justify-content: flex-end;
        padding-top: 12px;
        border-top: 1px solid var(--border-light);
      }
    }
  }
}
</style>
