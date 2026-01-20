import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '@/api'

export const useTaskStore = defineStore('task', () => {
  const tasks = ref([])
  const environments = ref([])
  const executions = ref([])
  const loading = ref(false)
  const tasksError = ref(null)
  const environmentsLoading = ref(false)
  const environmentsError = ref(null)
  const executionsError = ref(null)

  // 加载所有任务
  async function loadTasks() {
    loading.value = true
    tasksError.value = null
    try {
      tasks.value = await api.getTasks()
    } catch (error) {
      console.error('加载任务失败:', error)
      tasksError.value = error.message || '加载任务失败'
    } finally {
      loading.value = false
    }
  }

  // 加载环境列表
  async function loadEnvironments() {
    environmentsLoading.value = true
    environmentsError.value = null
    try {
      environments.value = await api.getEnvironments()
      return environments.value
    } catch (error) {
      console.error('加载环境失败:', error)
      environmentsError.value = error.message || '加载失败'
      return []
    } finally {
      environmentsLoading.value = false
    }
  }

  // 创建任务
  async function createTask(task) {
    await api.createTask(task)
    await loadTasks()
  }

  // 更新任务
  async function updateTask(task) {
    await api.updateTask(task)
    await loadTasks()
  }

  // 删除任务
  async function deleteTask(taskId) {
    await api.deleteTask(taskId)
    await loadTasks()
  }

  // 立即执行任务
  async function executeTask(taskId) {
    try {
      const execution = await api.executeTaskNow(taskId)
      return execution
    } catch (error) {
      console.error('执行任务失败:', error)
      throw error
    }
  }

  // 加载执行历史
  async function loadExecutions(taskId = '', limit = 100) {
    executionsError.value = null
    try {
      executions.value = await api.getExecutions(taskId, limit)
    } catch (error) {
      console.error('加载执行历史失败:', error)
      executionsError.value = error.message || '加载执行历史失败'
    }
  }

  return {
    tasks,
    environments,
    executions,
    loading,
    tasksError,
    environmentsLoading,
    environmentsError,
    executionsError,
    loadTasks,
    loadEnvironments,
    createTask,
    updateTask,
    deleteTask,
    executeTask,
    loadExecutions
  }
})
