// Wails 运行时导入
import { App } from '../../bindings/scriptguard/backend'

const {
  GetEnvironments,
  GetTasks,
  CreateTask,
  UpdateTask,
  DeleteTask,
  ExecuteTaskNow,
  GetExecutions,
  GetLogs,
  GetConfig,
  GetAllConfig,
  UpdateConfig,
  SelectScriptFile,
  TestNotification
} = App

// API 封装
export const api = {
  // 环境相关
  async getEnvironments() {
    return await GetEnvironments()
  },

  // 任务相关
  async getTasks() {
    return await GetTasks()
  },

  async createTask(task) {
    return await CreateTask(task)
  },

  async updateTask(task) {
    return await UpdateTask(task)
  },

  async deleteTask(taskId) {
    return await DeleteTask(taskId)
  },

  async executeTaskNow(taskId) {
    return await ExecuteTaskNow(taskId)
  },

  // 执行历史相关
  async getExecutions(taskId = '', limit = 100) {
    return await GetExecutions(taskId, limit)
  },

  // 日志相关
  async getLogs(executionId = '', taskId = '', limit = 1000) {
    return await GetLogs(executionId, taskId, limit)
  },

  // 配置相关
  async getConfig(key) {
    return await GetConfig(key)
  },

  async getAllConfig() {
    return await GetAllConfig()
  },

  async updateConfig(key, value) {
    return await UpdateConfig(key, value)
  },

  // 文件选择
  async selectScriptFile() {
    return await SelectScriptFile()
  },

  // SG-013: 测试通知
  async testNotification(target, webhook) {
    return await TestNotification(target, webhook)
  }
}

export default api
