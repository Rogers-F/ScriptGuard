<template>
  <div class="dashboard">
    <h1 class="page-title">仪表盘</h1>

    <!-- 统计卡片 -->
    <div class="stats-grid">
      <div class="stat-card">
        <div class="stat-icon" style="background: linear-gradient(135deg, #6366f1, #8b5cf6)">
          <el-icon :size="32"><List /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-label">总任务数</div>
          <div class="stat-value">{{ stats.totalTasks }}</div>
          <div class="stat-trend">
            <span :class="{ positive: stats.enabledTasks > 0 }">
              {{ stats.enabledTasks }} 个运行中
            </span>
          </div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon" style="background: linear-gradient(135deg, #10b981, #059669)">
          <el-icon :size="32"><Select /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-label">今日成功</div>
          <div class="stat-value">{{ stats.todaySuccess }}</div>
          <div class="stat-trend">
            <span class="positive">成功率 {{ stats.successRate }}%</span>
          </div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon" style="background: linear-gradient(135deg, #ef4444, #dc2626)">
          <el-icon :size="32"><CloseBold /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-label">今日失败</div>
          <div class="stat-value">{{ stats.todayFailed }}</div>
          <div class="stat-trend">
            <span :class="{ negative: stats.todayFailed > 0 }">
              需要关注
            </span>
          </div>
        </div>
      </div>

      <div class="stat-card">
        <div class="stat-icon" style="background: linear-gradient(135deg, #f59e0b, #d97706)">
          <el-icon :size="32"><Clock /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-label">平均耗时</div>
          <div class="stat-value">{{ stats.avgDuration }}s</div>
          <div class="stat-trend">
            <span>近7天平均</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 图表区域 -->
    <div class="charts-row">
      <div class="chart-card">
        <div class="card-header">
          <h3>执行趋势（近7天）</h3>
        </div>
        <v-chart :option="executionTrendOption" autoresize style="height: 300px" />
      </div>

      <div class="chart-card">
        <div class="card-header">
          <h3>任务状态分布</h3>
        </div>
        <v-chart :option="taskStatusOption" autoresize style="height: 300px" />
      </div>
    </div>

    <!-- 最近执行 -->
    <div class="recent-executions">
      <div class="section-header">
        <h3>最近执行</h3>
        <el-button text @click="$router.push('/history')">
          查看全部 <el-icon><ArrowRight /></el-icon>
        </el-button>
      </div>

      <el-table :data="recentExecutions" stripe style="width: 100%">
        <el-table-column prop="task_name" label="任务名称" width="200" />
        <el-table-column prop="start_time" label="执行时间" width="180">
          <template #default="{ row }">
            {{ formatTime(row.start_time) }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" size="small">
              {{ getStatusText(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="duration_ms" label="耗时" width="120">
          <template #default="{ row }">
            {{ (row.duration_ms / 1000).toFixed(2) }}s
          </template>
        </el-table-column>
        <el-table-column prop="error_message" label="错误信息" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="row.error_message" class="error-text">{{ row.error_message }}</span>
            <span v-else class="success-text">执行成功</span>
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
import { List, Select, CloseBold, Clock, ArrowRight } from '@element-plus/icons-vue'
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
    backgroundColor: 'rgba(15, 23, 42, 0.9)',
    borderColor: '#6366f1',
    textStyle: { color: '#e2e8f0' }
  },
  grid: { left: '3%', right: '4%', bottom: '3%', top: '10%', containLabel: true },
  xAxis: {
    type: 'category',
    data: ['周一', '周二', '周三', '周四', '周五', '周六', '周日'],
    axisLine: { lineStyle: { color: '#475569' } },
    axisLabel: { color: '#94a3b8' }
  },
  yAxis: {
    type: 'value',
    axisLine: { lineStyle: { color: '#475569' } },
    axisLabel: { color: '#94a3b8' },
    splitLine: { lineStyle: { color: '#334155' } }
  },
  series: [
    {
      name: '成功',
      type: 'line',
      data: [12, 15, 10, 18, 14, 16, 20],
      smooth: true,
      lineStyle: { color: '#10b981', width: 3 },
      areaStyle: { color: 'rgba(16, 185, 129, 0.1)' },
      itemStyle: { color: '#10b981' }
    },
    {
      name: '失败',
      type: 'line',
      data: [2, 1, 3, 1, 2, 0, 1],
      smooth: true,
      lineStyle: { color: '#ef4444', width: 3 },
      areaStyle: { color: 'rgba(239, 68, 68, 0.1)' },
      itemStyle: { color: '#ef4444' }
    }
  ]
}))

const taskStatusOption = computed(() => ({
  backgroundColor: 'transparent',
  tooltip: {
    trigger: 'item',
    backgroundColor: 'rgba(15, 23, 42, 0.9)',
    borderColor: '#6366f1',
    textStyle: { color: '#e2e8f0' }
  },
  legend: {
    top: '5%',
    left: 'center',
    textStyle: { color: '#94a3b8' }
  },
  series: [
    {
      type: 'pie',
      radius: ['40%', '70%'],
      center: ['50%', '60%'],
      data: [
        { value: stats.enabledTasks, name: '运行中', itemStyle: { color: '#10b981' } },
        { value: stats.totalTasks - stats.enabledTasks, name: '已停止', itemStyle: { color: '#64748b' } }
      ],
      label: {
        color: '#e2e8f0',
        fontSize: 14
      },
      emphasis: {
        itemStyle: {
          shadowBlur: 10,
          shadowOffsetX: 0,
          shadowColor: 'rgba(99, 102, 241, 0.5)'
        }
      }
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

  // 计算今日统计
  const today = new Date().toDateString()
  const todayExecutions = taskStore.executions.filter(e =>
    new Date(e.start_time).toDateString() === today
  )

  stats.todaySuccess = todayExecutions.filter(e => e.status === 'success').length
  stats.todayFailed = todayExecutions.filter(e => e.status === 'failed').length
  stats.successRate = todayExecutions.length > 0
    ? Math.round((stats.todaySuccess / todayExecutions.length) * 100)
    : 0

  // 计算平均耗时
  const durations = taskStore.executions.filter(e => e.duration_ms > 0).map(e => e.duration_ms)
  stats.avgDuration = durations.length > 0
    ? (durations.reduce((a, b) => a + b, 0) / durations.length / 1000).toFixed(2)
    : 0

  recentExecutions.value = taskStore.executions.slice(0, 10)
}

function formatTime(time) {
  return new Date(time).toLocaleString('zh-CN')
}

function getStatusType(status) {
  const map = {
    success: 'success',
    failed: 'danger',
    running: 'warning'
  }
  return map[status] || 'info'
}

function getStatusText(status) {
  const map = {
    success: '成功',
    failed: '失败',
    running: '运行中'
  }
  return map[status] || status
}
</script>

<style lang="scss" scoped>
.dashboard {
  .page-title {
    font-size: 28px;
    font-weight: 600;
    margin-bottom: 24px;
    background: linear-gradient(90deg, #6366f1, #8b5cf6);
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 20px;
    margin-bottom: 24px;

    .stat-card {
      background: linear-gradient(135deg, #1e293b 0%, #0f172a 100%);
      border: 1px solid var(--border-color);
      border-radius: 12px;
      padding: 20px;
      display: flex;
      gap: 16px;
      transition: all 0.3s ease;

      &:hover {
        transform: translateY(-4px);
        box-shadow: 0 8px 16px rgba(99, 102, 241, 0.2);
      }

      .stat-icon {
        width: 64px;
        height: 64px;
        border-radius: 12px;
        display: flex;
        align-items: center;
        justify-content: center;
        color: #fff;
      }

      .stat-content {
        flex: 1;

        .stat-label {
          font-size: 13px;
          color: var(--text-secondary);
          margin-bottom: 8px;
        }

        .stat-value {
          font-size: 32px;
          font-weight: 700;
          color: var(--text-primary);
          margin-bottom: 4px;
        }

        .stat-trend {
          font-size: 12px;

          .positive {
            color: #10b981;
          }

          .negative {
            color: #ef4444;
          }
        }
      }
    }
  }

  .charts-row {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
    gap: 20px;
    margin-bottom: 24px;
  }

  .chart-card,
  .recent-executions {
    background: linear-gradient(135deg, #1e293b 0%, #0f172a 100%);
    border: 1px solid var(--border-color);
    border-radius: 12px;
    padding: 20px;

    .card-header,
    .section-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 16px;

      h3 {
        font-size: 18px;
        font-weight: 600;
      }
    }
  }

  .error-text {
    color: #ef4444;
  }

  .success-text {
    color: #10b981;
  }
}
</style>
