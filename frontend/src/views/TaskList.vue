<template>
  <div class="page-container task-list">
    <div class="page-header">
      <div>
        <h1>Tasks</h1>
        <p>Manage and schedule automation scripts</p>
      </div>
      <el-button type="primary" :icon="Plus" class="create-btn" @click="showCreateDialog = true">
        New Task
      </el-button>
    </div>

    <!-- Empty State -->
    <el-empty
      v-if="!taskStore.tasks.length && !taskStore.loading"
      description="No tasks configured"
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
                {{ task.enabled ? 'Active' : 'Inactive' }}
            </el-tag>
          </div>
          <el-dropdown trigger="click" @command="(cmd) => handleCommand(cmd, task)">
            <el-button text circle size="small">
              <el-icon><MoreFilled /></el-icon>
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="edit">Edit</el-dropdown-item>
                <el-dropdown-item command="toggle">{{ task.enabled ? 'Disable' : 'Enable' }}</el-dropdown-item>
                <el-dropdown-item divided command="delete" class="text-danger">Delete</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>

        <div class="card-body">
          <div class="info-row">
            <span class="label">Env</span>
            <span class="value">{{ task.conda_env }}</span>
          </div>
          <div class="info-row">
             <span class="label">Schedule</span>
             <span class="value font-mono">{{ task.cron_expr || 'Manual Only' }}</span>
          </div>
          <div class="script-path" :title="task.script_path">
             <el-icon><Folder /></el-icon>
             {{ getFileName(task.script_path) }}
          </div>
        </div>

        <div class="card-footer">
           <el-button
             class="run-btn"
             :type="task.enabled ? 'primary' : 'default'"
             plain
             size="small"
             :loading="executingTasks.has(task.id)"
             @click="executeTask(task.id)"
           >
             <el-icon><VideoPlay /></el-icon> Run Now
           </el-button>
        </div>
      </div>
    </div>

    <!-- Create/Edit Dialog -->
    <el-dialog
      v-model="showCreateDialog"
      :title="editingTask ? 'Edit Task' : 'New Task'"
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
        <el-form-item label="Name" prop="name">
          <el-input v-model="taskForm.name" placeholder="e.g. Daily Data Processing" />
        </el-form-item>

        <el-form-item label="Conda Environment" prop="conda_env">
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

        <el-form-item label="Script File" prop="script_path">
          <el-input v-model="taskForm.script_path" placeholder="Path to .py file">
            <template #append>
              <el-button @click="selectScriptFile">Browse</el-button>
            </template>
          </el-input>
        </el-form-item>

        <el-form-item label="Schedule" prop="cron_expr">
          <CronEditor v-model="taskForm.cron_expr" />
        </el-form-item>

        <div class="form-switches">
          <div class="switch-row">
            <span>Failure Notifications</span>
            <el-switch v-model="taskForm.notify_on_failure" />
          </div>
          <div class="switch-row">
            <span>Active Status</span>
            <el-switch v-model="taskForm.enabled" />
          </div>
        </div>
      </el-form>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showCreateDialog = false">Cancel</el-button>
          <el-button type="primary" @click="saveTask" :loading="saving">Save Task</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { Plus, VideoPlay, MoreFilled, Folder } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useTaskStore } from '@/stores/task'
import CronEditor from '@/components/CronEditor.vue'
import api from '@/api'

const taskStore = useTaskStore()
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
  cron_expr: '',
  enabled: true,
  notify_on_failure: true
})

const taskRules = {
  name: [{ required: true, message: 'Required', trigger: 'blur' }],
  script_path: [{ required: true, message: 'Required', trigger: 'blur' }],
  conda_env: [{ required: true, message: 'Required', trigger: 'change' }],
  cron_expr: [{ required: true, message: 'Required', trigger: 'change' }]
}

onMounted(async () => {
  await taskStore.loadTasks()
  await taskStore.loadEnvironments()
})

function getFileName(path) {
  if (!path) return 'No file';
  return path.split(/[\\/]/).pop();
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
      if (editingTask.value) {
        await taskStore.updateTask({ ...editingTask.value, ...taskForm })
        ElMessage.success('Saved')
      } else {
        await taskStore.createTask(taskForm)
        ElMessage.success('Created')
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
  Object.assign(taskForm, { id: '', name: '', script_path: '', conda_env: '', cron_expr: '', enabled: true, notify_on_failure: true })
  editingTask.value = null
  taskFormRef.value?.resetFields()
}

async function deleteTask(task) {
  try {
    await ElMessageBox.confirm(`Delete "${task.name}"?`, 'Confirm', { confirmButtonText: 'Delete', cancelButtonText: 'Cancel', type: 'warning' })
    await taskStore.deleteTask(task.id)
    ElMessage.success('Deleted')
  } catch (error) { if (error !== 'cancel') ElMessage.error(error.message) }
}

async function toggleTask(task) {
  try {
    task.enabled = !task.enabled
    await taskStore.updateTask(task)
    ElMessage.success(task.enabled ? 'Enabled' : 'Disabled')
  } catch (error) {
    task.enabled = !task.enabled
    ElMessage.error(error.message)
  }
}

async function executeTask(taskId) {
  executingTasks.value.add(taskId)
  try {
    await taskStore.executeTask(taskId)
    ElMessage.success('Started')
  } catch (error) {
    ElMessage.error(error.message)
  } finally {
    executingTasks.value.delete(taskId)
  }
}

function getEnvironmentPlaceholder() {
  if (taskStore.environmentsLoading) return 'Loading...'
  return 'Select Environment'
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
