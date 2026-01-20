<template>
  <div class="page-container task-list">
    <div class="page-header">
      <div class="header-left">
        <h1>任务管理</h1>
        <p>配置与调度系统任务</p>
      </div>
      <el-button type="primary" :icon="Plus" @click="showCreateDialog = true">
        新建任务
      </el-button>
    </div>

    <!-- Empty State -->
    <el-empty
      v-if="!taskStore.tasks.length && !taskStore.loading"
      description="暂无任务"
      :image-size="120"
    />

    <!-- Task Grid -->
    <div v-else class="task-grid">
      <div
        v-for="task in taskStore.tasks"
        :key="task.id"
        class="task-card"
        :class="{ inactive: !task.enabled }"
      >
        <div class="card-header">
          <div class="task-identity">
            <h3 class="task-name" :title="task.name">{{ task.name }}</h3>
            <el-tag
              :type="task.enabled ? 'success' : 'info'"
              size="small"
              effect="plain"
              class="status-tag"
            >
              {{ task.enabled ? 'Active' : 'Stopped' }}
            </el-tag>
          </div>

          <div class="action-buttons">
            <el-tooltip content="立即执行" :show-after="500">
              <el-button
                text
                circle
                size="small"
                @click="executeTask(task.id)"
                :loading="executingTasks.has(task.id)"
              >
                <el-icon><VideoPlay /></el-icon>
              </el-button>
            </el-tooltip>
            <el-dropdown trigger="click" @command="(cmd) => handleCommand(cmd, task)">
              <el-button text circle size="small">
                <el-icon><More /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="edit">编辑任务</el-dropdown-item>
                  <el-dropdown-item command="toggle">{{ task.enabled ? '停用任务' : '启用任务' }}</el-dropdown-item>
                  <el-dropdown-item divided command="delete" class="text-danger">删除任务</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>

        <div class="card-body">
          <div class="info-item">
            <span class="label">Schedule</span>
            <span class="value font-mono">{{ task.cron_expr || '-' }}</span>
          </div>
          <div class="info-item">
            <span class="label">Environment</span>
            <span class="value">{{ task.conda_env }}</span>
          </div>
          <div class="info-item full-width">
            <span class="label">Script</span>
            <span class="value path" :title="task.script_path">{{ task.script_path }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Create/Edit Dialog -->
    <el-dialog
      v-model="showCreateDialog"
      :title="editingTask ? '编辑任务' : '新建任务'"
      width="600px"
      class="custom-dialog"
      :close-on-click-modal="false"
      @closed="resetForm"
    >
      <el-form
        ref="taskFormRef"
        :model="taskForm"
        :rules="taskRules"
        label-position="top"
      >
        <el-form-item label="任务名称" prop="name">
          <el-input v-model="taskForm.name" placeholder="如：数据每日备份" />
        </el-form-item>

        <el-form-item label="执行环境 (Conda)" prop="conda_env">
          <el-select
            v-model="taskForm.conda_env"
            :placeholder="getEnvironmentPlaceholder()"
            style="width: 100%"
            filterable
            :loading="taskStore.environmentsLoading"
          >
            <template v-if="taskStore.environmentsError">
              <el-option disabled value="">
                <div style="color: var(--color-danger); text-align: center; padding: 8px 0">
                  {{ taskStore.environmentsError }}
                </div>
              </el-option>
            </template>
            <template v-else-if="!taskStore.environmentsLoading && taskStore.environments.length === 0">
              <el-option disabled value="">
                <div style="color: var(--text-secondary); text-align: center; padding: 8px 0">
                  未找到 Conda 环境，请手动输入或检查 Conda 安装
                </div>
              </el-option>
            </template>
            <template v-else>
              <el-option
                v-for="env in taskStore.environments"
                :key="env.name"
                :label="env.name"
                :value="env.name"
                :disabled="!env.is_valid"
              >
                <div style="display: flex; justify-content: space-between; align-items: center">
                  <span>{{ env.name }}</span>
                  <el-tag v-if="!env.is_valid" type="danger" size="small">无效</el-tag>
                </div>
              </el-option>
            </template>
          </el-select>
        </el-form-item>

        <el-form-item label="脚本路径" prop="script_path">
          <el-input v-model="taskForm.script_path" placeholder="C:\path\to\script.py">
            <template #append>
              <el-button @click="selectScriptFile">浏览</el-button>
            </template>
          </el-input>
        </el-form-item>

        <el-form-item label="执行计划" prop="cron_expr">
          <CronEditor v-model="taskForm.cron_expr" />
        </el-form-item>

        <div class="form-switches">
          <el-form-item class="mb-0">
            <div class="switch-row">
              <span>失败告警通知</span>
              <el-switch v-model="taskForm.notify_on_failure" />
            </div>
          </el-form-item>
          <el-form-item class="mb-0">
            <div class="switch-row">
              <span>立即启用</span>
              <el-switch v-model="taskForm.enabled" />
            </div>
          </el-form-item>
        </div>
      </el-form>

      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="saveTask" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { Plus, VideoPlay, More } from '@element-plus/icons-vue'
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
  name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
  script_path: [{ required: true, message: '请选择脚本路径', trigger: 'blur' }],
  conda_env: [{ required: true, message: '请选择执行环境', trigger: 'change' }],
  cron_expr: [{ required: true, message: '请配置执行计划', trigger: 'change' }]
}

onMounted(async () => {
  await taskStore.loadTasks()
  await taskStore.loadEnvironments()
})

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
        ElMessage.success('任务更新成功')
      } else {
        await taskStore.createTask(taskForm)
        ElMessage.success('任务创建成功')
      }
      showCreateDialog.value = false
      resetForm()
    } catch (error) {
      ElMessage.error('操作失败: ' + error.message)
    } finally {
      saving.value = false
    }
  })
}

function resetForm() {
  Object.assign(taskForm, {
    id: '',
    name: '',
    script_path: '',
    conda_env: '',
    cron_expr: '',
    enabled: true,
    notify_on_failure: true
  })
  editingTask.value = null
  taskFormRef.value?.resetFields()
}

async function deleteTask(task) {
  try {
    await ElMessageBox.confirm(
      `确定要删除任务 "${task.name}" 吗？此操作不可恢复。`,
      '删除确认',
      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await taskStore.deleteTask(task.id)
    ElMessage.success('任务已删除')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败: ' + error.message)
    }
  }
}

async function toggleTask(task) {
  try {
    task.enabled = !task.enabled
    await taskStore.updateTask(task)
    ElMessage.success(task.enabled ? '任务已启用' : '任务已停用')
  } catch (error) {
    task.enabled = !task.enabled
    ElMessage.error('操作失败: ' + error.message)
  }
}

async function executeTask(taskId) {
  executingTasks.value.add(taskId)
  try {
    await taskStore.executeTask(taskId)
    ElMessage.success('任务已开始执行，请在日志监控中查看输出')
  } catch (error) {
    ElMessage.error('执行失败: ' + error.message)
  } finally {
    executingTasks.value.delete(taskId)
  }
}

function getEnvironmentPlaceholder() {
  if (taskStore.environmentsLoading) {
    return '正在加载环境列表...'
  } else if (taskStore.environmentsError) {
    return '加载环境失败'
  } else if (taskStore.environments.length === 0) {
    return '未找到环境，可手动输入'
  }
  return '选择Conda环境'
}

async function selectScriptFile() {
  try {
    const selectedPath = await api.selectScriptFile()
    if (selectedPath) {
      taskForm.script_path = selectedPath
    }
  } catch (error) {
    console.error('选择脚本文件失败:', error)
    ElMessage.warning(`选择文件失败: ${error.message || '未知错误'}，请手动输入脚本路径`)
  }
}
</script>

<style lang="scss" scoped>
.task-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
}

.task-card {
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-light);
  border-radius: 8px;
  padding: 20px;
  transition: all 0.2s ease;
  display: flex;
  flex-direction: column;

  &:hover {
    border-color: var(--border-hover);
  }

  &.inactive {
    opacity: 0.6;
  }

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 20px;

    .task-identity {
      .task-name {
        font-size: 16px;
        font-weight: 600;
        margin: 0 0 8px 0;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
        max-width: 200px;
      }
    }

    .action-buttons {
      display: flex;
      gap: 4px;
    }
  }

  .card-body {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 16px;
    font-size: 13px;

    .info-item {
      display: flex;
      flex-direction: column;
      gap: 4px;

      &.full-width {
        grid-column: span 2;
      }

      .label {
        color: var(--text-tertiary);
        font-size: 11px;
        text-transform: uppercase;
        font-weight: 600;
        letter-spacing: 0.05em;
      }

      .value {
        color: var(--text-secondary);

        &.path {
          font-family: monospace;
          background: var(--bg-tertiary);
          padding: 4px 6px;
          border-radius: 4px;
          white-space: nowrap;
          overflow: hidden;
          text-overflow: ellipsis;
        }

        &.font-mono {
          font-family: monospace;
        }
      }
    }
  }
}

.form-switches {
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid var(--border-light);

  .switch-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;
    padding: 8px 0;
  }
}

.text-danger { color: var(--color-danger); }
.mb-0 { margin-bottom: 0; }
</style>
