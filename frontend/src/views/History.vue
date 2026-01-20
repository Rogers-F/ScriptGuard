<template>
  <div class="page-container history-page">
    <div class="page-header">
      <div>
        <h1>Execution History</h1>
        <p>Audit trail of all task runs</p>
      </div>
      <div class="header-right">
        <el-date-picker
          v-model="dateRange"
          type="daterange"
          range-separator="-"
          start-placeholder="Start"
          end-placeholder="End"
          style="width: 240px"
          @change="loadExecutions"
        />
        <el-select v-model="selectedTask" placeholder="Filter Task" style="width: 180px" clearable @change="loadExecutions">
          <el-option v-for="task in taskStore.tasks" :key="task.id" :label="task.name" :value="task.id" />
        </el-select>
      </div>
    </div>

    <div class="card-panel timeline-section">
      <div class="panel-header">
        <h3>Recent Activity</h3>
        <el-button :icon="Refresh" circle size="small" @click="loadExecutions" />
      </div>

      <el-timeline v-if="executions.length > 0">
        <el-timeline-item
          v-for="exec in executions"
          :key="exec.id"
          :timestamp="formatTime(exec.start_time)"
          placement="top"
          :color="getTimelineColor(exec.status)"
          :hollow="true"
        >
          <div class="execution-item">
            <div class="item-header">
                <span class="task-name">{{ getTaskName(exec.task_id) }}</span>
                <span class="duration">{{ (exec.duration_ms / 1000).toFixed(2) }}s</span>
            </div>

            <div class="item-status">
               <el-tag :type="getStatusType(exec.status)" size="small" effect="light">
                  {{ exec.status.toUpperCase() }}
               </el-tag>
               <span class="exit-code" v-if="exec.exit_code !== 0">Exit: {{ exec.exit_code }}</span>
            </div>

            <div v-if="exec.error_message" class="error-msg">
              {{ exec.error_message }}
            </div>

            <el-button size="small" link type="primary" class="view-log-btn" @click="viewLogs(exec.id)">
               View Logs
            </el-button>
          </div>
        </el-timeline-item>
      </el-timeline>

      <el-empty v-else description="No records" />
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { Refresh } from '@element-plus/icons-vue'
import { useTaskStore } from '@/stores/task'

const router = useRouter()
const taskStore = useTaskStore()
const selectedTask = ref('')
const dateRange = ref([])
const executions = ref([])

onMounted(async () => {
  await taskStore.loadTasks()
  await loadExecutions()
})

async function loadExecutions() {
  await taskStore.loadExecutions(selectedTask.value, 50)
  let result = taskStore.executions
  if (dateRange.value && dateRange.value.length === 2) {
    const startDate = new Date(dateRange.value[0]); startDate.setHours(0, 0, 0, 0)
    const endDate = new Date(dateRange.value[1]); endDate.setHours(23, 59, 59, 999)
    result = result.filter(e => { const t = new Date(e.start_time); return t >= startDate && t <= endDate })
  }
  executions.value = result
}

function getTaskName(taskId) { const t = taskStore.tasks.find(x => x.id === taskId); return t ? t.name : 'Unknown' }
function formatTime(time) { return new Date(time).toLocaleString('zh-CN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' }) }
function getStatusType(s) { return { success: 'success', failed: 'danger', running: 'warning' }[s] || 'info' }
function getTimelineColor(s) { return { success: '#059669', failed: '#dc2626', running: '#d97706' }[s] || '#a8a29e' }
function viewLogs(id) { router.push(`/logs?execution=${id}`) }
</script>

<style lang="scss" scoped>
.history-page {
  .header-right { display: flex; gap: 12px; }

  .timeline-section {
    .panel-header { display: flex; justify-content: space-between; margin-bottom: 24px; }

    .execution-item {
      background: rgba(255,255,255,0.5);
      border: 1px solid var(--border-light);
      border-radius: 8px;
      padding: 12px 16px;
      margin-bottom: 8px;

      .item-header {
         display: flex; justify-content: space-between; margin-bottom: 8px;
         .task-name { font-weight: 600; color: var(--text-primary); }
         .duration { font-family: var(--font-mono); color: var(--text-tertiary); font-size: 12px; }
      }

      .item-status {
          display: flex; align-items: center; gap: 8px;
          .exit-code { font-size: 12px; color: var(--color-danger); }
      }

      .error-msg {
          margin-top: 8px; font-size: 12px; color: var(--color-danger);
          background: #fef2f2; padding: 4px 8px; border-radius: 4px;
      }

      .view-log-btn { margin-top: 8px; padding: 0; }
    }
  }
}
</style>
