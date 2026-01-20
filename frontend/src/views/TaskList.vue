<template>
  <div class="page-container task-list">
    <div class="page-header">
      <div>
        <h1>{{ t.tasks.title }}</h1>
        <p>{{ t.tasks.subtitle }}</p>
      </div>
      <el-button type="primary" :icon="Plus" class="create-btn" @click="showCreateDialog = true">
        {{ t.tasks.createTask }}
      </el-button>
    </div>

    <!-- Empty State -->
    <el-empty
      v-if="!taskStore.tasks.length && !taskStore.loading"
      :description="t.tasks.noTasks"
      :image-size="140"
    />

    <!-- Task Grid -->
    <div v-else class="task-grid">
      <div
        v-for="task in taskStore.tasks"
        :key="task.id"
        class="task-card glass-panel"
        :class="{ inactive: !task.enabled }"
      >
        <div class="card-header">
          <div class="header-main">
            <h3 class="task-name font-serif" :title="task.name">{{ task.name }}</h3>
            <el-tag
                :type="task.enabled ? 'success' : 'info'"
                size="small"
                class="status-badge"
            >
                {{ task.enabled ? t.common.enabled : t.common.disabled }}
            </el-tag>
          </div>
          <el-dropdown trigger="click" @command="(cmd) => handleCommand(cmd, task)">
            <el-button text circle size="small">
              <el-icon><MoreFilled /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="edit">{{ t.common.edit }}</el-dropdown-item>
                <el-dropdown-item command="toggle">{{ task.enabled ? t.common.disabled : t.common.enabled }}</el-dropdown-item>
                <el-dropdown-item divided command="delete" class="text-danger">{{ t.common.delete }}</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>

        <div class="card-body">
          <div class="info-row">
            <span class="label">{{ t.tasks.environment }}</span>
            <span class="value">{{ task.conda_env }}</span>
          </div>
          <div class="info-row">
             <span class="label">{{ t.tasks.schedule }}</span>
             <span class="value" :title="getScheduleTooltip(task)">{{ formatScheduleText(task) }}</span>
          </div>
          <div class="script-path" :title="task.script_path">
             <el-icon><Folder /></el-icon>
             {{ getFileName(task.script_path) }}
          </div>
        </div>

        <div class="card-footer">
           <el-button
             class="run-btn"
             type="primary"
             size="small"
             :loading="executingTasks.has(task.id)"
             @click="executeTask(task.id)"
           >
             <el-icon><VideoPlay /></el-icon> {{ t.tasks.runNow }}
           </el-button>
        </div>
      </div>
    </div>

    <!-- Create/Edit Dialog -->
    <el-dialog
      v-model="showCreateDialog"
      :title="editingTask ? t.tasks.editTask : t.tasks.createTask"
      width="560px"
      class="custom-dialog"
      :close-on-click-modal="false"
      @closed="resetForm"
    >
      <el-form
        ref="taskFormRef"
        :model="taskForm"
        :rules="taskRules"
        label-position="top"
        class="clean-form"
      >
        <el-form-item :label="t.tasks.taskName" prop="name">
          <el-input v-model="taskForm.name" :placeholder="t.tasks.taskNamePlaceholder" />
        </el-form-item>

        <el-form-item :label="t.tasks.environment" prop="conda_env">
          <el-select
            v-model="taskForm.conda_env"
            :placeholder="getEnvironmentPlaceholder()"
            style="width: 100%"
            filterable
            :loading="taskStore.environmentsLoading"
          >
            <el-option
              v-for="env in taskStore.environments"
              :key="env.name"
              :label="env.name"
              :value="env.name"
              :disabled="!env.is_valid"
            />
          </el-select>
        </el-form-item>

        <el-form-item :label="t.tasks.scriptFile" prop="script_path">
          <el-input v-model="taskForm.script_path" :placeholder="t.tasks.scriptPlaceholder">
            <template #append>
              <el-button @click="selectScriptFile">{{ t.tasks.browse }}</el-button>
            </template>
          </el-input>
        </el-form-item>

        <el-form-item :label="t.tasks.schedule" prop="cron_exprs">
          <CronEditor v-model="taskForm.cron_exprs" />
        </el-form-item>

        <div class="form-switches">
          <div class="switch-row">
            <span>{{ t.tasks.failureAlert }}</span>
            <el-switch v-model="taskForm.notify_on_failure" />
          </div>
          <div class="switch-row">
            <span>{{ t.tasks.enableNow }}</span>
            <el-switch v-model="taskForm.enabled" />
          </div>
        </div>
      </el-form>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showCreateDialog = false">{{ t.common.cancel }}</el-button>
          <el-button type="primary" @click="saveTask" :loading="saving">{{ t.common.save }}</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { Plus, VideoPlay, MoreFilled, Folder } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useTaskStore } from '@/stores/task'
import { useLanguageStore } from '@/stores/language'
import CronEditor from '@/components/CronEditor.vue'
import api from '@/api'

const taskStore = useTaskStore()
const langStore = useLanguageStore()
const t = computed(() => langStore.t)

const showCreateDialog = ref(false)
const editingTask = ref(null)
const saving = ref(false)
const executingTasks = ref(new Set())
const taskFormRef = ref(null)

const taskForm = reactive({
  id: '',
  name: '',
  script_path: '',
  conda_env: '',
  cron_expr: '',       // 兼容字段
  cron_exprs: [],      // 新字段：Cron 数组
  enabled: true,
  notify_on_failure: true
})

const taskRules = computed(() => ({
  name: [{ required: true, message: langStore.isChinese ? '请输入任务名称' : 'Task name is required', trigger: 'blur' }],
  script_path: [{ required: true, message: langStore.isChinese ? '请选择脚本文件' : 'Script file is required', trigger: 'blur' }],
  conda_env: [{ required: true, message: langStore.isChinese ? '请选择执行环境' : 'Environment is required', trigger: 'change' }],
  cron_exprs: [{
    validator: (_, val, cb) => {
      if (!Array.isArray(val) || val.length === 0) return cb(new Error(langStore.isChinese ? '请配置执行计划' : 'Schedule is required'))
      if (val.length > 60) return cb(new Error(langStore.isChinese ? '最多支持 60 个时间点' : 'Up to 60 time points'))
      cb()
    },
    trigger: 'change'
  }]
}))

onMounted(async () => {
  await taskStore.loadTasks()
  await taskStore.loadEnvironments()
})

function getFileName(path) {
  if (!path) return langStore.isChinese ? '未选择文件' : 'No file selected';
  return path.split(/[\\/]/).pop();
}

// 归一化任务的 cron 表达式
function normalizeExprs(task) {
  if (Array.isArray(task.cron_exprs) && task.cron_exprs.length) return task.cron_exprs
  if (task.cron_expr) return [task.cron_expr]
  return []
}

// 从 cron 表达式提取时间
function extractTimeFromCron(expr) {
  const reDaily = /^(\d{1,2}) (\d{1,2}) (\d{1,2}) \* \* \*$/
  const reWeekly = /^(\d{1,2}) (\d{1,2}) (\d{1,2}) \* \* ([0-6](?:,[0-6])*)$/
  const reMonthly = /^(\d{1,2}) (\d{1,2}) (\d{1,2}) (\d{1,2}) \* \*$/

  let m = expr.match(reDaily) || expr.match(reWeekly) || expr.match(reMonthly)
  if (m) {
    const h = String(parseInt(m[3], 10)).padStart(2, '0')
    const min = String(parseInt(m[2], 10)).padStart(2, '0')
    const s = String(parseInt(m[1], 10)).padStart(2, '0')
    return `${h}:${min}:${s}`
  }
  return null
}

// 格式化任务调度文本（用于卡片显示）
function formatScheduleText(task) {
  const exprs = normalizeExprs(task)
  if (!exprs.length) return t.value.tasks.manualOnly

  // interval（单条）
  if (exprs.length === 1) {
    let m = exprs[0].match(/^0 \*\/(\d+) \* \* \* \*$/)
    if (m) return langStore.isChinese ? `每 ${m[1]} 分钟` : `Every ${m[1]} min`
    m = exprs[0].match(/^0 0 \*\/(\d+) \* \* \*$/)
    if (m) return langStore.isChinese ? `每 ${m[1]} 小时` : `Every ${m[1]} hr`
    m = exprs[0].match(/^0 0 0 \*\/(\d+) \* \*$/)
    if (m) return langStore.isChinese ? `每 ${m[1]} 天` : `Every ${m[1]} day`
  }

  // 提取时间点
  const times = exprs.map(e => extractTimeFromCron(e)).filter(Boolean)
  if (times.length === 0) return exprs[0] // 兜底

  const timesStr = times.slice(0, 3).join(langStore.isChinese ? '、' : ', ')
  const suffix = times.length > 3 ? (langStore.isChinese ? ` 等${times.length}个` : ` +${times.length - 3}`) : ''

  // 检测是 weekly 还是 monthly 还是 daily
  const reWeekly = /^(\d{1,2}) (\d{1,2}) (\d{1,2}) \* \* ([0-6](?:,[0-6])*)$/
  const reMonthly = /^(\d{1,2}) (\d{1,2}) (\d{1,2}) (\d{1,2}) \* \*$/

  const weeklyMatch = exprs[0].match(reWeekly)
  if (weeklyMatch) {
    const weekdayNames = langStore.isChinese
      ? ['日', '一', '二', '三', '四', '五', '六']
      : ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
    const days = weeklyMatch[4].split(',').map(d => weekdayNames[parseInt(d, 10)]).join(langStore.isChinese ? '、' : ', ')
    return langStore.isChinese ? `每周${days} ${timesStr}${suffix}` : `Weekly ${days} ${timesStr}${suffix}`
  }

  const monthlyMatch = exprs[0].match(reMonthly)
  if (monthlyMatch) {
    const day = parseInt(monthlyMatch[4], 10)
    return langStore.isChinese ? `每月${day}号 ${timesStr}${suffix}` : `Monthly day ${day} ${timesStr}${suffix}`
  }

  return langStore.isChinese ? `每天 ${timesStr}${suffix}` : `Daily ${timesStr}${suffix}`
}

// 获取调度 tooltip（显示所有时间点）
function getScheduleTooltip(task) {
  const exprs = normalizeExprs(task)
  if (exprs.length <= 1) return ''

  const times = exprs.map(e => extractTimeFromCron(e)).filter(Boolean)
  return times.join(', ')
}

function handleCommand(command, task) {
  if (command === 'edit') editTask(task)
  else if (command === 'delete') deleteTask(task)
  else if (command === 'toggle') toggleTask(task)
}

function editTask(task) {
  editingTask.value = task
  Object.assign(taskForm, task)
  showCreateDialog.value = true
}

async function saveTask() {
  if (!taskFormRef.value) return
  await taskFormRef.value.validate(async (valid) => {
    if (!valid) return
    saving.value = true
    try {
      // 归一化：确保 cron_expr = cron_exprs[0]
      const payload = {
        ...taskForm,
        cron_exprs: Array.isArray(taskForm.cron_exprs) ? taskForm.cron_exprs : [],
        cron_expr: Array.isArray(taskForm.cron_exprs) && taskForm.cron_exprs.length > 0 ? taskForm.cron_exprs[0] : ''
      }
      if (editingTask.value) {
        await taskStore.updateTask({ ...editingTask.value, ...payload })
        ElMessage.success(t.value.tasks.saveSuccess)
      } else {
        await taskStore.createTask(payload)
        ElMessage.success(t.value.tasks.createSuccess)
      }
      showCreateDialog.value = false
      resetForm()
    } catch (error) {
      ElMessage.error(error.message)
    } finally {
      saving.value = false
    }
  })
}

function resetForm() {
  Object.assign(taskForm, { id: '', name: '', script_path: '', conda_env: '', cron_expr: '', cron_exprs: [], enabled: true, notify_on_failure: true })
  editingTask.value = null
  taskFormRef.value?.resetFields()
}

async function deleteTask(task) {
  try {
    await ElMessageBox.confirm(
      `${t.value.tasks.deleteConfirm} "${task.name}"?`,
      t.value.tasks.deleteTitle,
      { confirmButtonText: t.value.common.delete, cancelButtonText: t.value.common.cancel, type: 'warning' }
    )
    await taskStore.deleteTask(task.id)
    ElMessage.success(t.value.tasks.deleted)
  } catch (error) { if (error !== 'cancel') ElMessage.error(error.message) }
}

async function toggleTask(task) {
  try {
    task.enabled = !task.enabled
    await taskStore.updateTask(task)
    ElMessage.success(task.enabled ? t.value.common.enabled : t.value.common.disabled)
  } catch (error) {
    task.enabled = !task.enabled
    ElMessage.error(error.message)
  }
}

async function executeTask(taskId) {
  executingTasks.value.add(taskId)
  try {
    await taskStore.executeTask(taskId)
    ElMessage.success(t.value.tasks.taskStarted)
  } catch (error) {
    ElMessage.error(error.message)
  } finally {
    executingTasks.value.delete(taskId)
  }
}

function getEnvironmentPlaceholder() {
  if (taskStore.environmentsLoading) return t.value.tasks.loadingEnv
  return t.value.tasks.selectEnv
}

async function selectScriptFile() {
  try {
    const selectedPath = await api.selectScriptFile()
    if (selectedPath) taskForm.script_path = selectedPath
  } catch (error) { ElMessage.warning(error.message) }
}
</script>

<style lang="scss" scoped>
.task-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 24px;
}

.task-card {
  display: flex;
  flex-direction: column;

  &.inactive {
    opacity: 0.8;
    background: rgba(240, 240, 240, 0.5);
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 20px;

    .header-main {
        flex: 1;
        margin-right: 12px;

        .task-name {
            font-size: 18px;
            color: var(--text-primary);
            margin-bottom: 6px;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }
    }
  }

  .card-body {
    flex: 1;
    font-size: 13px;

    .info-row {
        display: flex;
        justify-content: space-between;
        margin-bottom: 8px;

        .label { color: var(--text-tertiary); }
        .value { color: var(--text-secondary); font-weight: 500; }
    }

    .script-path {
        margin-top: 16px;
        background: rgba(0,0,0,0.03);
        padding: 8px 12px;
        border-radius: 6px;
        color: var(--text-secondary);
        display: flex;
        align-items: center;
        gap: 8px;
        font-family: var(--font-mono);
        font-size: 12px;
    }
  }

  .card-footer {
    margin-top: 20px;
    padding-top: 16px;
    border-top: 1px solid var(--border-light);
    display: flex;
    justify-content: flex-end;

    .run-btn {
      min-width: 100px;
    }
  }
}

.form-switches {
    margin-top: 24px;
    background: var(--bg-hover);
    padding: 16px;
    border-radius: 8px;

    .switch-row {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 12px;
        &:last-child { margin-bottom: 0; }
        color: var(--text-secondary);
    }
}
</style>
