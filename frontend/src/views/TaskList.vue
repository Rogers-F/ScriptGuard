<template>
  <div class="task-list-page">
    <div class="page-header">
      <div class="header-left">
        <h1>任务管理</h1>
        <p>管理所有Python脚本定时任务</p>
      </div>
      <div class="header-right">
        <el-button type="primary" :icon="Plus" @click="showCreateDialog = true">
          新建任务
        </el-button>
      </div>
    </div>

    <!-- 任务列表 -->
    <div class="task-cards">
      <el-empty v-if="!taskStore.tasks.length && !taskStore.loading" description="暂无任务，点击右上角新建任务" />

      <TransitionGroup name="list" tag="div" class="cards-grid">
        <div
          v-for="task in taskStore.tasks"
          :key="task.id"
          class="task-card"
          :class="{ disabled: !task.enabled }"
        >
          <div class="card-header">
            <div class="task-info">
              <h3>{{ task.name }}</h3>
              <el-tag :type="task.enabled ? 'success' : 'info'" size="small">
                {{ task.enabled ? '运行中' : '已停止' }}
              </el-tag>
            </div>
            <div class="card-actions">
              <el-tooltip content="立即执行">
                <el-button
                  :icon="VideoPlay"
                  circle
                  size="small"
                  @click="executeTask(task.id)"
                  :loading="executingTasks.has(task.id)"
                />
              </el-tooltip>
              <el-tooltip content="编辑">
                <el-button
                  :icon="Edit"
                  circle
                  size="small"
                  @click="editTask(task)"
                />
              </el-tooltip>
              <el-tooltip content="删除">
                <el-button
                  :icon="Delete"
                  circle
                  size="small"
                  type="danger"
                  @click="deleteTask(task)"
                />
              </el-tooltip>
            </div>
          </div>

          <div class="card-body">
            <div class="info-row">
              <el-icon><Document /></el-icon>
              <span class="label">脚本路径:</span>
              <span class="value">{{ task.script_path }}</span>
            </div>
            <div class="info-row">
              <el-icon><Box /></el-icon>
              <span class="label">执行环境:</span>
              <el-tag size="small" type="warning">{{ task.conda_env }}</el-tag>
            </div>
            <div class="info-row">
              <el-icon><Clock /></el-icon>
              <span class="label">执行计划:</span>
              <span class="value schedule">{{ getScheduleText(task.cron_expr) }}</span>
            </div>
            <div class="info-row">
              <el-icon><Bell /></el-icon>
              <span class="label">失败告警:</span>
              <el-tag :type="task.notify_on_failure ? 'success' : 'info'" size="small">
                {{ task.notify_on_failure ? '已开启' : '已关闭' }}
              </el-tag>
            </div>
          </div>

          <div class="card-footer">
            <el-switch
              v-model="task.enabled"
              @change="toggleTask(task)"
              active-text="启用"
              inactive-text="停用"
            />
          </div>
        </div>
      </TransitionGroup>
    </div>

    <!-- 创建/编辑任务对话框 -->
    <el-dialog
      v-model="showCreateDialog"
      :title="editingTask ? '编辑任务' : '新建任务'"
      width="700px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="taskFormRef"
        :model="taskForm"
        :rules="taskRules"
        label-width="100px"
      >
        <el-form-item label="任务名称" prop="name">
          <el-input
            v-model="taskForm.name"
            placeholder="请输入任务名称"
            clearable
          />
        </el-form-item>

        <el-form-item label="脚本路径" prop="script_path">
          <el-input
            v-model="taskForm.script_path"
            placeholder="C:\path\to\script.py"
            clearable
          >
            <template #append>
              <el-button @click="selectScriptFile">浏览</el-button>
            </template>
          </el-input>
        </el-form-item>

        <el-form-item label="执行环境" prop="conda_env">
          <el-select
            v-model="taskForm.conda_env"
            :placeholder="getEnvironmentPlaceholder()"
            style="width: 100%"
            filterable
            :loading="taskStore.environmentsLoading"
          >
            <template v-if="taskStore.environmentsError">
              <el-option disabled value="">
                <div style="color: var(--el-color-danger); text-align: center; padding: 8px 0">
                  {{ taskStore.environmentsError }}
                </div>
              </el-option>
            </template>
            <template v-else-if="!taskStore.environmentsLoading && taskStore.environments.length === 0">
              <el-option disabled value="">
                <div style="color: var(--el-text-color-secondary); text-align: center; padding: 8px 0">
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

        <el-form-item label="执行计划" prop="cron_expr">
          <CronEditor v-model="taskForm.cron_expr" />
        </el-form-item>

        <el-form-item label="失败告警">
          <el-switch v-model="taskForm.notify_on_failure" />
          <span style="margin-left: 12px; color: var(--text-secondary); font-size: 13px">
            脚本执行失败时发送钉钉/企业微信通知
          </span>
        </el-form-item>

        <el-form-item label="启用状态">
          <el-switch v-model="taskForm.enabled" />
          <span style="margin-left: 12px; color: var(--text-secondary); font-size: 13px">
            关闭后任务将不会自动执行
          </span>
        </el-form-item>
      </el-form>

      <template #footer>
        <el-button @click="showCreateDialog = false">取消</el-button>
        <el-button type="primary" @click="saveTask" :loading="saving">
          {{ editingTask ? '保存' : '创建' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, VideoPlay, Edit, Delete, Document, Box, Clock, Bell } from '@element-plus/icons-vue'
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

function getScheduleText(cronExpr) {
  // 简单解析显示（可以用cronstrue库更准确）
  return cronExpr || '未设置'
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
    await taskStore.updateTask(task)
    ElMessage.success(task.enabled ? '任务已启用' : '任务已停用')
  } catch (error) {
    ElMessage.error('操作失败: ' + error.message)
    task.enabled = !task.enabled // 回滚
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
    // 显示真实错误信息，同时提示用户可以手动输入
    ElMessage.warning(`选择文件失败: ${error.message || '未知错误'}，请手动输入脚本路径`)
  }
}
</script>

<style lang="scss" scoped>
.task-list-page {
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
  }

  .task-cards {
    .cards-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
      gap: 20px;
    }

    .task-card {
      background: linear-gradient(135deg, #1e293b 0%, #0f172a 100%);
      border: 1px solid var(--border-color);
      border-radius: 12px;
      padding: 20px;
      transition: all 0.3s ease;

      &:hover {
        transform: translateY(-4px);
        box-shadow: 0 8px 16px rgba(99, 102, 241, 0.2);
        border-color: #6366f1;
      }

      &.disabled {
        opacity: 0.6;
      }

      .card-header {
        display: flex;
        justify-content: space-between;
        align-items: flex-start;
        margin-bottom: 16px;
        padding-bottom: 16px;
        border-bottom: 1px solid rgba(255, 255, 255, 0.1);

        .task-info {
          h3 {
            font-size: 18px;
            font-weight: 600;
            margin-bottom: 8px;
          }
        }

        .card-actions {
          display: flex;
          gap: 8px;
        }
      }

      .card-body {
        .info-row {
          display: flex;
          align-items: center;
          gap: 8px;
          margin-bottom: 12px;
          font-size: 13px;

          .el-icon {
            color: #6366f1;
            font-size: 16px;
          }

          .label {
            color: var(--text-secondary);
            min-width: 70px;
          }

          .value {
            color: var(--text-primary);
            flex: 1;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;

            &.schedule {
              font-weight: 500;
              color: #10b981;
            }
          }
        }
      }

      .card-footer {
        margin-top: 16px;
        padding-top: 16px;
        border-top: 1px solid rgba(255, 255, 255, 0.1);
        display: flex;
        justify-content: flex-end;
      }
    }
  }
}

.list-enter-active,
.list-leave-active {
  transition: all 0.3s ease;
}

.list-enter-from {
  opacity: 0;
  transform: translateY(30px);
}

.list-leave-to {
  opacity: 0;
  transform: scale(0.9);
}
</style>
