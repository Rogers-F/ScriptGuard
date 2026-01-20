import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

// 语言配置
const messages = {
  zh: {
    // 通用
    common: {
      save: '保存',
      cancel: '取消',
      delete: '删除',
      edit: '编辑',
      confirm: '确认',
      reset: '重置',
      refresh: '刷新',
      test: '测试',
      search: '搜索...',
      loading: '加载中...',
      noData: '暂无数据',
      success: '成功',
      failed: '失败',
      running: '运行中',
      unknown: '未知',
      enabled: '已启用',
      disabled: '已停用',
      valid: '有效',
      invalid: '无效',
      viewAll: '查看全部',
      viewLogs: '查看日志',
      requiresAttention: '需要关注',
      noIssues: '无异常'
    },

    // 侧边栏菜单
    menu: {
      analytics: '数据概览',
      dashboard: '仪表盘',
      history: '执行历史',
      scheduling: '任务调度',
      tasks: '任务管理',
      logs: '日志监控',
      settings: '系统设置'
    },

    // 仪表盘
    dashboard: {
      title: '仪表盘',
      subtitle: '系统概览与性能指标',
      totalTasks: '总任务数',
      active: '运行中',
      successToday: '今日成功',
      rate: '成功率',
      failedToday: '今日失败',
      avgDuration: '平均耗时',
      sevenDayAvg: '近7日平均',
      executionTrend: '执行趋势',
      taskDistribution: '任务分布',
      recentActivity: '最近活动',
      viewAllHistory: '查看全部历史',
      task: '任务',
      time: '时间',
      status: '状态',
      duration: '耗时',
      stopped: '已停用',
      weekdays: ['周一', '周二', '周三', '周四', '周五', '周六', '周日']
    },

    // 任务管理
    tasks: {
      title: '任务管理',
      subtitle: '配置与调度自动化脚本',
      createTask: '新建任务',
      editTask: '编辑任务',
      noTasks: '暂无任务',
      taskName: '任务名称',
      taskNamePlaceholder: '如：每日数据处理',
      environment: '执行环境',
      selectEnv: '选择执行环境',
      loadingEnv: '加载中...',
      scriptFile: '脚本文件',
      scriptPlaceholder: 'Python 脚本路径',
      browse: '浏览',
      schedule: '执行计划',
      manualOnly: '仅手动',
      failureAlert: '失败告警通知',
      enableNow: '立即启用',
      runNow: '立即执行',
      toggle: '切换状态',
      deleteConfirm: '确定要删除任务',
      deleteTitle: '删除确认',
      taskStarted: '任务已开始执行',
      saveSuccess: '保存成功',
      createSuccess: '创建成功',
      deleted: '已删除'
    },

    // 日志查看器
    logs: {
      title: '日志',
      allTasks: '全部任务',
      autoscrollOn: '自动滚动 开',
      autoscrollOff: '自动滚动 关',
      clear: '清空',
      export: '导出',
      waitingForLogs: '等待日志...',
      retry: '重试',
      savedTo: '已保存至'
    },

    // 执行历史
    history: {
      title: '执行历史',
      subtitle: '所有任务运行的审计记录',
      filterTask: '筛选任务',
      startDate: '开始日期',
      endDate: '结束日期',
      recentActivity: '最近活动',
      noRecords: '暂无记录',
      exitCode: '退出码'
    },

    // 设置
    settings: {
      title: '系统设置',

      // 标签页
      tabs: {
        general: '通用',
        alerts: '告警通知',
        system: '系统',
        environments: '环境管理',
        about: '关于'
      },

      // 通用设置
      general: {
        title: '通用设置',
        subtitle: '应用程序基本配置',
        language: '界面语言',
        languageDesc: '选择应用程序显示语言',
        chinese: '简体中文',
        english: 'English'
      },

      // 告警通知
      alerts: {
        title: '告警通知',
        subtitle: '配置任务失败时的告警渠道',
        dingtalk: '钉钉机器人',
        wecom: '企业微信机器人',
        webhookUrl: 'Webhook 地址',
        urlRequired: '请输入 Webhook 地址',
        testSent: '测试消息已发送'
      },

      // 系统设置
      system: {
        title: '系统配置',
        retentionPolicy: '日志保留策略',
        keepLogs: '保留天数',
        execution: '执行配置',
        maxConcurrency: '最大并发数',
        timeout: '超时时间（秒）'
      },

      // 环境管理
      environments: {
        title: 'Conda 环境',
        subtitle: '任务执行可用的 Python 环境',
        refreshList: '刷新环境列表',
        refreshed: '环境列表已刷新',
        name: '环境名称',
        path: '路径',
        status: '状态'
      },

      // 关于
      about: {
        version: '版本',
        desc: '安全可靠的脚本自动化执行系统'
      },

      // 操作
      saveChanges: '保存更改',
      saved: '已保存',
      loadFailed: '加载失败',
      refreshFailed: '刷新失败'
    }
  },

  en: {
    common: {
      save: 'Save',
      cancel: 'Cancel',
      delete: 'Delete',
      edit: 'Edit',
      confirm: 'Confirm',
      reset: 'Reset',
      refresh: 'Refresh',
      test: 'Test',
      search: 'Search...',
      loading: 'Loading...',
      noData: 'No data',
      success: 'Success',
      failed: 'Failed',
      running: 'Running',
      unknown: 'Unknown',
      enabled: 'Enabled',
      disabled: 'Disabled',
      valid: 'Valid',
      invalid: 'Invalid',
      viewAll: 'View All',
      viewLogs: 'View Logs',
      requiresAttention: 'Requires attention',
      noIssues: 'No issues'
    },

    menu: {
      analytics: 'Analytics',
      dashboard: 'Dashboard',
      history: 'History',
      scheduling: 'Scheduling',
      tasks: 'Tasks',
      logs: 'Logs',
      settings: 'Settings'
    },

    dashboard: {
      title: 'Dashboard',
      subtitle: 'System overview and performance metrics',
      totalTasks: 'Total Tasks',
      active: 'active',
      successToday: 'Success Today',
      rate: 'Rate',
      failedToday: 'Failed Today',
      avgDuration: 'Avg Duration',
      sevenDayAvg: '7-day average',
      executionTrend: 'Execution Trend',
      taskDistribution: 'Task Distribution',
      recentActivity: 'Recent Activity',
      viewAllHistory: 'View All History',
      task: 'Task',
      time: 'Time',
      status: 'Status',
      duration: 'Duration',
      stopped: 'Stopped',
      weekdays: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun']
    },

    tasks: {
      title: 'Tasks',
      subtitle: 'Configure and schedule automation scripts',
      createTask: 'New Task',
      editTask: 'Edit Task',
      noTasks: 'No tasks',
      taskName: 'Task Name',
      taskNamePlaceholder: 'e.g., Daily Data Processing',
      environment: 'Environment',
      selectEnv: 'Select environment',
      loadingEnv: 'Loading...',
      scriptFile: 'Script File',
      scriptPlaceholder: 'Python script path',
      browse: 'Browse',
      schedule: 'Schedule',
      manualOnly: 'Manual only',
      failureAlert: 'Failure Alert',
      enableNow: 'Enable Now',
      runNow: 'Run Now',
      toggle: 'Toggle',
      deleteConfirm: 'Are you sure you want to delete task',
      deleteTitle: 'Delete Confirmation',
      taskStarted: 'Task started',
      saveSuccess: 'Saved successfully',
      createSuccess: 'Created successfully',
      deleted: 'Deleted'
    },

    logs: {
      title: 'Logs',
      allTasks: 'All Tasks',
      autoscrollOn: 'Autoscroll On',
      autoscrollOff: 'Autoscroll Off',
      clear: 'Clear',
      export: 'Export',
      waitingForLogs: 'Waiting for logs...',
      retry: 'Retry',
      savedTo: 'Saved to'
    },

    history: {
      title: 'Execution History',
      subtitle: 'Audit trail of all task runs',
      filterTask: 'Filter Task',
      startDate: 'Start',
      endDate: 'End',
      recentActivity: 'Recent Activity',
      noRecords: 'No records',
      exitCode: 'Exit'
    },

    settings: {
      title: 'Settings',

      tabs: {
        general: 'General',
        alerts: 'Alerts',
        system: 'System',
        environments: 'Environments',
        about: 'About'
      },

      general: {
        title: 'General Settings',
        subtitle: 'Application preferences',
        language: 'Language',
        languageDesc: 'Select display language',
        chinese: '简体中文',
        english: 'English'
      },

      alerts: {
        title: 'Notifications',
        subtitle: 'Configure alert channels for task failures',
        dingtalk: 'DingTalk Robot',
        wecom: 'WeCom Robot',
        webhookUrl: 'Webhook URL',
        urlRequired: 'URL required',
        testSent: 'Test notification sent'
      },

      system: {
        title: 'System',
        retentionPolicy: 'Retention Policy',
        keepLogs: 'Keep logs for (days)',
        execution: 'Execution',
        maxConcurrency: 'Max Concurrency',
        timeout: 'Timeout (Seconds)'
      },

      environments: {
        title: 'Conda Environments',
        subtitle: 'Available Python environments for task execution',
        refreshList: 'Refresh List',
        refreshed: 'Environments refreshed',
        name: 'Name',
        path: 'Path',
        status: 'Status'
      },

      about: {
        version: 'Version',
        desc: 'Secure task automation for the modern web'
      },

      saveChanges: 'Save Changes',
      saved: 'Saved',
      loadFailed: 'Load failed',
      refreshFailed: 'Refresh failed'
    }
  }
}

export const useLanguageStore = defineStore('language', () => {
  // 从localStorage读取保存的语言设置，默认中文
  const currentLang = ref(localStorage.getItem('app-language') || 'zh')

  // 获取翻译文本
  const t = computed(() => messages[currentLang.value])

  // 切换语言
  function setLanguage(lang) {
    if (messages[lang]) {
      currentLang.value = lang
      localStorage.setItem('app-language', lang)
    }
  }

  // 获取当前语言
  function getLanguage() {
    return currentLang.value
  }

  // 检查是否为中文
  const isChinese = computed(() => currentLang.value === 'zh')

  return {
    currentLang,
    t,
    isChinese,
    setLanguage,
    getLanguage
  }
})
