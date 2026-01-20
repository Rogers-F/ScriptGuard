<template>
  <div class="page-container dashboard">
    <div class="page-header">
      <div>
        <h1>Dashboard</h1>
        <p>System overview and performance metrics</p>
      </div>
    </div>

    <!-- Stats Grid -->
    <div class="stats-grid">
      <div class="stat-card glass-panel">
        <div class="stat-header">
           <span class="label">Total Tasks</span>
           <el-icon><List /></el-icon>
        </div>
        <div class="stat-value font-serif">{{ stats.totalTasks }}</div>
        <div class="stat-footer">
          <span class="text-primary">{{ stats.enabledTasks }}</span> active
        </div>
      </div>

      <div class="stat-card glass-panel">
        <div class="stat-header">
           <span class="label">Success Today</span>
           <el-icon class="text-success"><Check /></el-icon>
        </div>
        <div class="stat-value font-serif text-success">{{ stats.todaySuccess }}</div>
        <div class="stat-footer">
          Rate: {{ stats.successRate }}%
        </div>
      </div>

      <div class="stat-card glass-panel">
        <div class="stat-header">
           <span class="label">Failed Today</span>
           <el-icon class="text-danger"><Warning /></el-icon>
        </div>
        <div class="stat-value font-serif" :class="{ 'text-danger': stats.todayFailed > 0 }">
          {{ stats.todayFailed }}
        </div>
        <div class="stat-footer">
          {{ stats.todayFailed > 0 ? 'Requires attention' : 'No issues' }}
        </div>
      </div>

      <div class="stat-card glass-panel">
        <div class="stat-header">
           <span class="label">Avg Duration</span>
           <el-icon><Timer /></el-icon>
        </div>
        <div class="stat-value font-serif">{{ stats.avgDuration }}<span class="unit">s</span></div>
        <div class="stat-footer">
          7-day average
        </div>
      </div>
    </div>

    <!-- Charts Section -->
    <div class="charts-section">
      <div class="card-panel chart-panel">
        <div class="panel-header">
          <h3>Execution Trend</h3>
        </div>
        <v-chart :option="executionTrendOption" autoresize class="chart-container" />
      </div>

      <div class="card-panel chart-panel">
        <div class="panel-header">
          <h3>Task Distribution</h3>
        </div>
        <v-chart :option="taskStatusOption" autoresize class="chart-container" />
      </div>
    </div>

    <!-- Recent Executions -->
    <div class="card-panel recent-section">
      <div class="panel-header">
        <h3>Recent Activity</h3>
        <el-button link type="primary" @click="$router.push('/history')">
          View All History
        </el-button>
      </div>

      <el-table
        :data="recentExecutions"
        style="width: 100%"
        :header-cell-style="{ background: 'transparent', color: '#78716c', fontFamily: 'var(--font-sans)' }"
        :row-style="{ background: 'transparent' }"
      >
        <el-table-column prop="task_name" label="Task" min-width="180">
          <template #default="{ row }">
            <span class="font-medium text-primary">{{ getTaskName(row.task_id) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="start_time" label="Time" width="180">
          <template #default="{ row }">
            <span class="text-secondary">{{ formatTime(row.start_time) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="Status" width="120">
          <template #default="{ row }">
            <div class="status-cell">
               <div class="dot" :class="row.status"></div>
               <span>{{ getStatusText(row.status) }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="duration_ms" label="Duration" width="120">
          <template #default="{ row }">
            <span class="text-secondary font-mono">{{ (row.duration_ms / 1000).toFixed(2) }}s</span>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import { LineChart, PieChart } from 'echarts/charts'
import { GridComponent, TooltipComponent, LegendComponent } from 'echarts/components'
import VChart from 'vue-echarts'
import { List, Check, Warning, Timer } from '@element-plus/icons-vue'
import { useTaskStore } from '@/stores/task'

use([CanvasRenderer, LineChart, PieChart, GridComponent, TooltipComponent, LegendComponent])

const taskStore = useTaskStore()

const stats = reactive({
  totalTasks: 0,
  enabledTasks: 0,
  todaySuccess: 0,
  todayFailed: 0,
  avgDuration: 0,
  successRate: 0
})

const recentExecutions = ref([])

const executionTrendOption = computed(() => ({
  backgroundColor: 'transparent',
  tooltip: {
    trigger: 'axis',
    backgroundColor: '#fff',
    borderColor: '#e7e5e4',
    textStyle: { color: '#1c1917', fontFamily: 'Inter' },
    padding: 12,
    extraCssText: 'box-shadow: 0 4px 12px rgba(0,0,0,0.08); border-radius: 8px;'
  },
  grid: { left: 10, right: 10, bottom: 0, top: 30, containLabel: true },
  xAxis: {
    type: 'category',
    data: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'],
    axisLine: { show: false },
    axisTick: { show: false },
    axisLabel: { color: '#a8a29e', fontFamily: 'Inter' }
  },
  yAxis: {
    type: 'value',
    splitLine: { lineStyle: { color: '#e7e5e4', type: 'dashed' } },
    axisLabel: { color: '#a8a29e', fontFamily: 'Inter' }
  },
  series: [
    {
      name: 'Success',
      type: 'line',
      data: [12, 15, 10, 18, 14, 16, 20],
      smooth: true,
      symbol: 'circle',
      symbolSize: 8,
      lineStyle: { color: '#059669', width: 2 },
      itemStyle: { color: '#fff', borderWidth: 2, borderColor: '#059669' }
    },
    {
      name: 'Failed',
      type: 'line',
      data: [2, 1, 3, 1, 2, 0, 1],
      smooth: true,
      symbol: 'circle',
      symbolSize: 8,
      lineStyle: { color: '#dc2626', width: 2 },
      itemStyle: { color: '#fff', borderWidth: 2, borderColor: '#dc2626' }
    }
  ]
}))

const taskStatusOption = computed(() => ({
  backgroundColor: 'transparent',
  tooltip: {
    trigger: 'item',
    backgroundColor: '#fff',
    borderColor: '#e7e5e4'
  },
  legend: {
    bottom: 0,
    textStyle: { color: '#78716c' },
    itemWidth: 8,
    itemHeight: 8
  },
  series: [
    {
      type: 'pie',
      radius: ['60%', '80%'],
      center: ['50%', '45%'],
      data: [
        { value: stats.enabledTasks, name: 'Active', itemStyle: { color: '#d97706' } },
        { value: stats.totalTasks - stats.enabledTasks, name: 'Stopped', itemStyle: { color: '#e7e5e4' } }
      ],
      label: { show: false },
      itemStyle: { borderColor: '#f4f2ed', borderWidth: 2 }
    }
  ]
}))

onMounted(async () => {
  await loadDashboardData()
})

async function loadDashboardData() {
  await taskStore.loadTasks()
  await taskStore.loadExecutions('', 10)

  stats.totalTasks = taskStore.tasks.length
  stats.enabledTasks = taskStore.tasks.filter(t => t.enabled).length

  const today = new Date().toDateString()
  const todayExecutions = taskStore.executions.filter(e =>
    new Date(e.start_time).toDateString() === today
  )

  stats.todaySuccess = todayExecutions.filter(e => e.status === 'success').length
  stats.todayFailed = todayExecutions.filter(e => e.status === 'failed').length
  stats.successRate = todayExecutions.length > 0
    ? Math.round((stats.todaySuccess / todayExecutions.length) * 100)
    : 0

  const durations = taskStore.executions.filter(e => e.duration_ms > 0).map(e => e.duration_ms)
  stats.avgDuration = durations.length > 0
    ? (durations.reduce((a, b) => a + b, 0) / durations.length / 1000).toFixed(2)
    : 0

  recentExecutions.value = taskStore.executions.slice(0, 5)
}

function getTaskName(taskId) {
  const task = taskStore.tasks.find(t => t.id === taskId)
  return task ? task.name : 'Unknown'
}

function formatTime(time) {
  return new Date(time).toLocaleString('zh-CN', { timeZone: 'Asia/Shanghai' })
}

function getStatusText(status) {
  const map = { success: 'Success', failed: 'Failed', running: 'Running' }
  return map[status] || status
}
</script>

<style lang="scss" scoped>
.dashboard {
  .stats-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 24px;
    margin-bottom: 32px;

    @media (max-width: 1200px) {
      grid-template-columns: repeat(2, 1fr);
    }

    .stat-card {
      padding: 24px;
      display: flex;
      flex-direction: column;
      justify-content: space-between;
      height: 140px;

      .stat-header {
        display: flex;
        justify-content: space-between;
        align-items: flex-start;
        color: var(--text-tertiary);

        .label {
            font-size: 13px;
            font-weight: 500;
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }
      }

      .stat-value {
        font-size: 36px;
        color: var(--text-primary);
        line-height: 1.1;

        .unit { font-size: 16px; color: var(--text-tertiary); margin-left: 4px; font-family: var(--font-sans); }
      }

      .stat-footer {
        font-size: 13px;
        color: var(--text-secondary);
      }
    }
  }

  .charts-section {
    display: grid;
    grid-template-columns: 2fr 1fr;
    gap: 24px;
    margin-bottom: 32px;

    @media (max-width: 1000px) {
      grid-template-columns: 1fr;
    }

    .chart-panel {
      height: 360px;
      display: flex;
      flex-direction: column;

      .panel-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 16px;
        padding-bottom: 16px;
        border-bottom: 1px solid var(--border-light);
      }

      .chart-container {
        flex: 1;
        width: 100%;
      }
    }
  }

  .status-cell {
    display: flex;
    align-items: center;
    gap: 8px;

    .dot {
        width: 6px;
        height: 6px;
        border-radius: 50%;

        &.success { background: var(--color-success); }
        &.failed { background: var(--color-danger); }
        &.running { background: var(--color-warning); }
    }
  }
}
</style>
