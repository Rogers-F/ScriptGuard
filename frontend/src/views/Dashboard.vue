<template>
  <div class="page-container dashboard">
    <div class="page-header">
      <div>
        <h1>仪表盘</h1>
        <p>系统概况与运行统计</p>
      </div>
    </div>

    <!-- Stats Grid -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-label">总任务数</div>
        <div class="stat-value">{{ stats.totalTasks }}</div>
        <div class="stat-footer">
          <span class="status-indicator" :class="{ active: stats.enabledTasks > 0 }"></span>
          {{ stats.enabledTasks }} 个任务正在运行
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-label">今日执行成功</div>
        <div class="stat-value text-success">{{ stats.todaySuccess }}</div>
        <div class="stat-footer">
          成功率 {{ stats.successRate }}%
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-label">今日执行失败</div>
        <div class="stat-value" :class="{ 'text-danger': stats.todayFailed > 0 }">
          {{ stats.todayFailed }}
        </div>
        <div class="stat-footer text-tertiary">
          需要关注
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-label">平均耗时</div>
        <div class="stat-value">{{ stats.avgDuration }}<span class="unit">s</span></div>
        <div class="stat-footer">
          近7天平均值
        </div>
      </div>
    </div>

    <!-- Charts Section -->
    <div class="charts-section">
      <div class="card-panel chart-panel">
        <div class="panel-header">
          <h3>执行趋势</h3>
          <el-tag type="info" size="small" effect="plain">示例数据</el-tag>
        </div>
        <v-chart :option="executionTrendOption" autoresize class="chart-container" />
      </div>

      <div class="card-panel chart-panel">
        <div class="panel-header">
          <h3>任务状态分布</h3>
        </div>
        <v-chart :option="taskStatusOption" autoresize class="chart-container" />
      </div>
    </div>

    <!-- Recent Executions -->
    <div class="card-panel recent-section">
      <div class="panel-header">
        <h3>最近执行记录</h3>
        <el-button text bg size="small" @click="$router.push('/history')">
          查看全部
        </el-button>
      </div>

      <el-table
        :data="recentExecutions"
        style="width: 100%"
        :header-cell-style="{ background: 'transparent', color: '#a1a1aa' }"
        :row-style="{ background: 'transparent' }"
      >
        <el-table-column prop="task_name" label="任务名称" min-width="180">
          <template #default="{ row }">
            <span class="font-medium">{{ getTaskName(row.task_id) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="start_time" label="执行时间" width="180">
          <template #default="{ row }">
            <span class="text-secondary">{{ formatTime(row.start_time) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <span :class="['status-dot', row.status]"></span>
            <span class="status-text">{{ getStatusText(row.status) }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="duration_ms" label="耗时" width="120">
          <template #default="{ row }">
            <span class="text-mono">{{ (row.duration_ms / 1000).toFixed(2) }}s</span>
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
    backgroundColor: '#1f1f1f',
    borderColor: '#333',
    textStyle: { color: '#e5e5e5' },
    padding: [8, 12]
  },
  grid: { left: 0, right: 0, bottom: 0, top: 20, containLabel: true },
  xAxis: {
    type: 'category',
    data: ['周一', '周二', '周三', '周四', '周五', '周六', '周日'],
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
      data: [12, 15, 10, 18, 14, 16, 20],
      smooth: true,
      symbol: 'none',
      lineStyle: { color: '#10b981', width: 2 },
      itemStyle: { color: '#10b981' }
    },
    {
      name: '失败',
      type: 'line',
      data: [2, 1, 3, 1, 2, 0, 1],
      smooth: true,
      symbol: 'none',
      lineStyle: { color: '#ef4444', width: 2 },
      itemStyle: { color: '#ef4444' }
    }
  ]
}))

const taskStatusOption = computed(() => ({
  backgroundColor: 'transparent',
  tooltip: {
    trigger: 'item',
    backgroundColor: '#1f1f1f',
    borderColor: '#333',
    textStyle: { color: '#e5e5e5' }
  },
  legend: {
    bottom: 0,
    textStyle: { color: '#a1a1aa' },
    itemWidth: 8,
    itemHeight: 8
  },
  series: [
    {
      type: 'pie',
      radius: ['50%', '70%'],
      center: ['50%', '45%'],
      data: [
        { value: stats.enabledTasks, name: '运行中', itemStyle: { color: '#3b82f6' } },
        { value: stats.totalTasks - stats.enabledTasks, name: '已停止', itemStyle: { color: '#3f3f46' } }
      ],
      label: { show: false }
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
  return task ? task.name : '未知任务'
}

function formatTime(time) {
  return new Date(time).toLocaleString('zh-CN', { timeZone: 'Asia/Shanghai' })
}

function getStatusText(status) {
  const map = { success: '成功', failed: '失败', running: '运行中' }
  return map[status] || status
}
</script>

<style lang="scss" scoped>
.dashboard {
  .stats-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 24px;
    margin-bottom: 24px;

    @media (max-width: 1200px) {
      grid-template-columns: repeat(2, 1fr);
    }

    .stat-card {
      background: var(--bg-secondary);
      border: 1px solid var(--border-light);
      border-radius: 8px;
      padding: 20px;

      .stat-label {
        font-size: 13px;
        color: var(--text-secondary);
        margin-bottom: 8px;
      }

      .stat-value {
        font-size: 28px;
        font-weight: 600;
        color: var(--text-primary);
        letter-spacing: -0.02em;
        margin-bottom: 8px;

        .unit {
          font-size: 14px;
          color: var(--text-secondary);
          margin-left: 4px;
          font-weight: normal;
        }

        &.text-success { color: var(--color-success); }
        &.text-danger { color: var(--color-danger); }
      }

      .stat-footer {
        font-size: 12px;
        color: var(--text-secondary);
        display: flex;
        align-items: center;
        gap: 6px;

        .status-indicator {
          width: 6px;
          height: 6px;
          border-radius: 50%;
          background-color: var(--text-tertiary);

          &.active { background-color: var(--color-success); }
        }
      }
    }
  }

  .charts-section {
    display: grid;
    grid-template-columns: 2fr 1fr;
    gap: 24px;
    margin-bottom: 24px;

    @media (max-width: 1000px) {
      grid-template-columns: 1fr;
    }

    .chart-panel {
      height: 320px;
      display: flex;
      flex-direction: column;

      .panel-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 16px;

        h3 {
          font-size: 16px;
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

  .recent-section {
    .panel-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 16px;

      h3 {
        font-size: 16px;
        font-weight: 500;
        margin: 0;
      }
    }
  }

  .status-dot {
    display: inline-block;
    width: 6px;
    height: 6px;
    border-radius: 50%;
    margin-right: 8px;

    &.success { background-color: var(--color-success); }
    &.failed { background-color: var(--color-danger); }
    &.running { background-color: var(--color-warning); }
  }

  .text-mono {
    font-family: 'JetBrains Mono', 'Fira Code', monospace;
    font-size: 13px;
    color: var(--text-secondary);
  }

  .font-medium {
    font-weight: 500;
  }
}
</style>
