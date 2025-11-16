import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '@/api'

export const useTaskStore = defineStore('task', () => {
  const tasks = ref([])
  const environments = ref([])
  const executions = ref([])
  const loading = ref(false)

  // 加载所有任务
  async function loadTasks() {
    loading.value = true
    try {
      tasks.value = await api.getTasks()
    } catch (error) {
      console.error('加载任务失败:', error)
    } finally {
      loading.value = false
    }
  }

  // 加载环境列表
  async function loadEnvironments() {
    try {
      environments.value = await api.getEnvironments()
    } catch (error) {
      console.error('加载环境失败:', error)
    }
  }

  // 创建任务
  async function createTask(task) {
    try {
      await api.createTask(task)
      await loadTasks()
      return true
    } catch (error) {
      console.error('创建任务失败:', error)
      return false
    }
  }

  // 更新任务
  async function updateTask(task) {
    try {
      await api.updateTask(task)
      await loadTasks()
      return true
    } catch (error) {
      console.error('更新任务失败:', error)
      return false
    }
  }

  // 删除任务
  async function deleteTask(taskId) {
    try {
      await api.deleteTask(taskId)
      await loadTasks()
      return true
    } catch (error) {
      console.error('删除任务失败:', error)
      return false
    }
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
    try {
      executions.value = await api.getExecutions(taskId, limit)
    } catch (error) {
      console.error('加载执行历史失败:', error)
    }
  }

  return {
    tasks,
    environments,
    executions,
    loading,
    loadTasks,
    loadEnvironments,
    createTask,
    updateTask,
    deleteTask,
    executeTask,
    loadExecutions
  }
})
