<template>
  <div class="history-page">
    <div class="page-header">
      <div class="header-left">
        <h1>执行历史</h1>
        <p>查看所有任务的执行记录和统计分析</p>
      </div>
      <div class="header-right">
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          style="width: 300px"
          @change="loadExecutions"
        />
        <el-select v-model="selectedTask" placeholder="选择任务" style="width: 200px" clearable @change="loadExecutions">
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

    <!-- 统计卡片 -->
    <div class="stats-row">
      <div class="stat-card">
        <div class="stat-label">总执行次数</div>
        <div class="stat-value">{{ stats.total }}</div>
      </div>
      <div class="stat-card success">
        <div class="stat-label">成功次数</div>
        <div class="stat-value">{{ stats.success }}</div>
        <div class="stat-percent">{{ stats.successRate }}%</div>
      </div>
      <div class="stat-card failed">
        <div class="stat-label">失败次数</div>
        <div class="stat-value">{{ stats.failed }}</div>
        <div class="stat-percent">{{ stats.failureRate }}%</div>
      </div>
      <div class="stat-card">
        <div class="stat-label">平均耗时</div>
        <div class="stat-value">{{ stats.avgDuration }}s</div>
      </div>
    </div>

    <!-- 图表区域 -->
    <div class="charts-row">
      <div class="chart-card">
        <div class="card-header">
          <h3>执行趋势（近7天）</h3>
        </div>
        <v-chart :option="trendChartOption" autoresize style="height: 280px" />
      </div>

      <div class="chart-card">
        <div class="card-header">
          <h3>执行耗时分布</h3>
        </div>
        <v-chart :option="durationChartOption" autoresize style="height: 280px" />
      </div>
    </div>

    <!-- 执行记录时间轴 -->
    <div class="timeline-section">
      <div class="section-header">
        <h3>执行记录</h3>
        <el-button text @click="loadExecutions">
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
          <el-card class="execution-card">
            <div class="execution-header">
              <div class="task-name">
                <el-icon><Document /></el-icon>
                {{ getTaskName(exec.task_id) }}
              </div>
              <el-tag :type="getStatusType(exec.status)" size="small">
                {{ getStatusText(exec.status) }}
              </el-tag>
            </div>

            <div class="execution-details">
              <div class="detail-item">
                <span class="label">执行ID:</span>
                <span class="value">{{ exec.id.substring(0, 8) }}...</span>
              </div>
              <div class="detail-item">
                <span class="label">开始时间:</span>
                <span class="value">{{ formatFullTime(exec.start_time) }}</span>
              </div>
              <div class="detail-item">
                <span class="label">耗时:</span>
                <span class="value">{{ (exec.duration_ms / 1000).toFixed(2) }}秒</span>
              </div>
              <div class="detail-item">
                <span class="label">退出码:</span>
                <span class="value">{{ exec.exit_code }}</span>
              </div>
            </div>

            <div v-if="exec.error_message" class="error-message">
              <el-icon><Warning /></el-icon>
              {{ exec.error_message }}
            </div>

            <div class="execution-actions">
              <el-button size="small" text @click="viewLogs(exec.id)">
                <el-icon><View /></el-icon>
                查看日志
              </el-button>
            </div>
          </el-card>
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
import { Refresh, Document, Warning, View } from '@element-plus/icons-vue'
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

// 趋势图配置
const trendChartOption = computed(() => {
  // 按日期分组统计
  const dateMap = new Map()
  executions.value.forEach(exec => {
    const date = new Date(exec.start_time).toLocaleDateString('zh-CN')
    if (!dateMap.has(date)) {
      dateMap.set(date, { success: 0, failed: 0 })
    }
    const stat = dateMap.get(date)
    if (exec.status === 'success') {
      stat.success++
    } else if (exec.status === 'failed') {
      stat.failed++
    }
  })

  const dates = Array.from(dateMap.keys()).slice(-7)
  const successData = dates.map(date => dateMap.get(date).success)
  const failedData = dates.map(date => dateMap.get(date).failed)

  return {
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
      data: dates,
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
        data: successData,
        smooth: true,
        lineStyle: { color: '#10b981', width: 3 },
        areaStyle: { color: 'rgba(16, 185, 129, 0.1)' },
        itemStyle: { color: '#10b981' }
      },
      {
        name: '失败',
        type: 'line',
        data: failedData,
        smooth: true,
        lineStyle: { color: '#ef4444', width: 3 },
        areaStyle: { color: 'rgba(239, 68, 68, 0.1)' },
        itemStyle: { color: '#ef4444' }
      }
    ]
  }
})

// 耗时分布图配置
const durationChartOption = computed(() => {
  // 按耗时区间分组
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
      backgroundColor: 'rgba(15, 23, 42, 0.9)',
      borderColor: '#6366f1',
      textStyle: { color: '#e2e8f0' }
    },
    grid: { left: '3%', right: '4%', bottom: '3%', top: '10%', containLabel: true },
    xAxis: {
      type: 'category',
      data: ranges,
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
        name: '执行次数',
        type: 'bar',
        data: counts,
        itemStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: '#6366f1' },
              { offset: 1, color: '#8b5cf6' }
            ]
          },
          borderRadius: [8, 8, 0, 0]
        }
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
  executions.value = taskStore.executions

  // 计算统计数据
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
    minute: '2-digit'
  })
}

function formatFullTime(time) {
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

function getTimelineColor(status) {
  const map = {
    success: '#10b981',
    failed: '#ef4444',
    running: '#f59e0b'
  }
  return map[status] || '#64748b'
}

function viewLogs(executionId) {
  router.push(`/logs?execution=${executionId}`)
}
</script>

<style lang="scss" scoped>
.history-page {
  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24px;
    padding-bottom: 20px;
    border-bottom: 1px solid var(--border-color);

    .header-left {
      h1 {
        font-size: 28px;
        font-weight: 600;
        margin-bottom: 8px;
        background: linear-gradient(90deg, #6366f1, #8b5cf6);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
      }

      p {
        color: var(--text-secondary);
        font-size: 14px;
      }
    }

    .header-right {
      display: flex;
      gap: 12px;
    }
  }

  .stats-row {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 16px;
    margin-bottom: 24px;

    .stat-card {
      background: linear-gradient(135deg, #1e293b 0%, #0f172a 100%);
      border: 1px solid var(--border-color);
      border-radius: 12px;
      padding: 20px;
      text-align: center;

      &.success {
        border-color: rgba(16, 185, 129, 0.3);
      }

      &.failed {
        border-color: rgba(239, 68, 68, 0.3);
      }

      .stat-label {
        font-size: 13px;
        color: var(--text-secondary);
        margin-bottom: 8px;
      }

      .stat-value {
        font-size: 32px;
        font-weight: 700;
        color: var(--text-primary);
      }

      .stat-percent {
        font-size: 14px;
        margin-top: 4px;
        color: var(--text-secondary);
      }
    }
  }

  .charts-row {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
    gap: 20px;
    margin-bottom: 24px;
  }

  .chart-card {
    background: linear-gradient(135deg, #1e293b 0%, #0f172a 100%);
    border: 1px solid var(--border-color);
    border-radius: 12px;
    padding: 20px;

    .card-header {
      margin-bottom: 16px;

      h3 {
        font-size: 18px;
        font-weight: 600;
      }
    }
  }

  .timeline-section {
    background: linear-gradient(135deg, #1e293b 0%, #0f172a 100%);
    border: 1px solid var(--border-color);
    border-radius: 12px;
    padding: 24px;

    .section-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 20px;

      h3 {
        font-size: 18px;
        font-weight: 600;
      }
    }

    .execution-card {
      background: rgba(15, 23, 42, 0.5);
      border: 1px solid rgba(255, 255, 255, 0.1);

      .execution-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 12px;

        .task-name {
          display: flex;
          align-items: center;
          gap: 8px;
          font-size: 16px;
          font-weight: 500;
          color: var(--text-primary);
        }
      }

      .execution-details {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
        gap: 12px;
        margin-bottom: 12px;

        .detail-item {
          font-size: 13px;

          .label {
            color: var(--text-secondary);
            margin-right: 8px;
          }

          .value {
            color: var(--text-primary);
          }
        }
      }

      .error-message {
        background: rgba(239, 68, 68, 0.1);
        border: 1px solid rgba(239, 68, 68, 0.3);
        border-radius: 6px;
        padding: 12px;
        margin-bottom: 12px;
        color: #ef4444;
        font-size: 13px;
        display: flex;
        align-items: center;
        gap: 8px;
      }

      .execution-actions {
        display: flex;
        justify-content: flex-end;
        padding-top: 12px;
        border-top: 1px solid rgba(255, 255, 255, 0.05);
      }
    }
  }
}

:root {
  --text-primary: #e2e8f0;
  --text-secondary: #94a3b8;
  --border-color: #334155;
}
</style>
