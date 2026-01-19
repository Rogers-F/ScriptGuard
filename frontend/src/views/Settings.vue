<template>
  <div class="settings-page">
    <div class="page-header">
      <h1>系统设置</h1>
      <p>配置ScriptGuard的系统参数和告警通知</p>
    </div>

    <el-tabs v-model="activeTab" class="settings-tabs">
      <!-- 告警通知配置 -->
      <el-tab-pane label="告警通知" name="notification">
        <div class="tab-content">
          <el-form :model="notificationForm" label-width="140px" class="settings-form">
            <div class="form-section">
              <h3>钉钉机器人</h3>
              <el-form-item label="启用钉钉通知">
                <el-switch v-model="notificationForm.dingtalk_enabled" />
              </el-form-item>
              <el-form-item label="Webhook URL">
                <el-input
                  v-model="notificationForm.dingtalk_webhook"
                  placeholder="https://oapi.dingtalk.com/robot/send?access_token=..."
                  :disabled="!notificationForm.dingtalk_enabled"
                  clearable
                >
                  <template #append>
                    <el-button @click="testNotification('dingtalk')">测试</el-button>
                  </template>
                </el-input>
                <span class="form-help">
                  在钉钉群中添加自定义机器人，获取Webhook URL
                </span>
              </el-form-item>
            </div>

            <el-divider />

            <div class="form-section">
              <h3>企业微信机器人</h3>
              <el-form-item label="启用企业微信通知">
                <el-switch v-model="notificationForm.wecom_enabled" />
              </el-form-item>
              <el-form-item label="Webhook URL">
                <el-input
                  v-model="notificationForm.wecom_webhook"
                  placeholder="https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=..."
                  :disabled="!notificationForm.wecom_enabled"
                  clearable
                >
                  <template #append>
                    <el-button @click="testNotification('wecom')">测试</el-button>
                  </template>
                </el-input>
                <span class="form-help">
                  在企业微信群中添加群机器人，获取Webhook URL
                </span>
              </el-form-item>
            </div>
          </el-form>
        </div>
      </el-tab-pane>

      <!-- 系统配置 -->
      <el-tab-pane label="系统配置" name="system">
        <div class="tab-content">
          <el-form :model="systemForm" label-width="140px" class="settings-form">
            <div class="form-section">
              <h3>日志管理</h3>
              <el-form-item label="日志保留天数">
                <el-input-number
                  v-model="systemForm.log_retention_days"
                  :min="1"
                  :max="365"
                />
                <span class="form-help">
                  超过此天数的日志将被自动清理
                </span>
              </el-form-item>
            </div>

            <el-divider />

            <div class="form-section">
              <h3>执行设置</h3>
              <el-form-item label="最大并发任务数">
                <el-input-number
                  v-model="systemForm.max_concurrency"
                  :min="1"
                  :max="20"
                />
                <span class="form-help">
                  同时执行的任务数量上限
                </span>
              </el-form-item>

              <el-form-item label="执行超时（秒）">
                <el-input-number
                  v-model="systemForm.execution_timeout_seconds"
                  :min="0"
                  :max="86400"
                  :step="60"
                />
                <span class="form-help">
                  0 表示不限制；允许范围 0 或 60~86400 秒（1分钟~24小时）
                </span>
              </el-form-item>
            </div>
          </el-form>
        </div>
      </el-tab-pane>

      <!-- 环境管理 -->
      <el-tab-pane label="环境管理" name="environment">
        <div class="tab-content">
          <div class="env-header">
            <p>管理系统中可用的Conda环境</p>
            <el-button type="primary" @click="refreshEnvironments" :loading="refreshing">
              <el-icon><Refresh /></el-icon>
              刷新环境列表
            </el-button>
          </div>

          <el-table :data="environments" stripe style="width: 100%; margin-top: 20px">
            <el-table-column prop="name" label="环境名称" width="200">
              <template #default="{ row }">
                <div class="env-name">
                  <el-icon><Box /></el-icon>
                  {{ row.name }}
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="path" label="路径" show-overflow-tooltip />
            <el-table-column prop="python_path" label="Python路径" show-overflow-tooltip />
            <el-table-column prop="is_valid" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.is_valid ? 'success' : 'danger'" size="small">
                  {{ row.is_valid ? '有效' : '无效' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-tab-pane>

      <!-- 关于 -->
      <el-tab-pane label="关于" name="about">
        <div class="tab-content about-content">
          <div class="app-logo">
            <el-icon :size="64"><Monitor /></el-icon>
          </div>
          <h2>ScriptGuard</h2>
          <p class="version">版本 1.0.0</p>
          <p class="description">
            Python脚本监控与定时执行系统
          </p>

          <el-divider />

          <div class="features">
            <h3>主要特性</h3>
            <ul>
              <li>支持Cron表达式的灵活定时任务</li>
              <li>多Conda环境隔离执行</li>
              <li>实时日志监控与持久化</li>
              <li>执行历史统计分析</li>
              <li>钉钉/企业微信告警通知</li>
              <li>现代化UI界面</li>
            </ul>
          </div>

          <el-divider />

          <div class="tech-stack">
            <h3>技术栈</h3>
            <el-tag>Wails 3</el-tag>
            <el-tag>Go 1.21</el-tag>
            <el-tag>Vue 3</el-tag>
            <el-tag>Element Plus</el-tag>
            <el-tag>SQLite</el-tag>
            <el-tag>ECharts</el-tag>
          </div>

          <el-divider />

          <div class="links">
            <el-button text>
              <el-icon><Link /></el-icon>
              GitHub
            </el-button>
            <el-button text>
              <el-icon><Document /></el-icon>
              使用文档
            </el-button>
          </div>

          <p class="copyright">
            © 2024 ScriptGuard. All rights reserved.
          </p>
        </div>
      </el-tab-pane>
    </el-tabs>

    <!-- 底部操作栏 -->
    <div class="settings-footer" v-if="activeTab !== 'about' && activeTab !== 'environment'">
      <el-button @click="resetForm">重置</el-button>
      <el-button type="primary" @click="saveSettings" :loading="saving">
        保存设置
      </el-button>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh, Box, Monitor, Link, Document } from '@element-plus/icons-vue'
import { useTaskStore } from '@/stores/task'
import api from '@/api'

const taskStore = useTaskStore()

const activeTab = ref('notification')
const saving = ref(false)
const refreshing = ref(false)
const environments = ref([])

const notificationForm = reactive({
  dingtalk_enabled: false,
  dingtalk_webhook: '',
  wecom_enabled: false,
  wecom_webhook: ''
})

const systemForm = reactive({
  log_retention_days: 30,
  max_concurrency: 5,
  execution_timeout_seconds: 3600
})

onMounted(async () => {
  await loadSettings()
  await loadEnvironments()
})

async function loadSettings() {
  try {
    const config = await api.getAllConfig()

    // 加载告警配置
    notificationForm.dingtalk_webhook = config.dingtalk_webhook || ''
    notificationForm.wecom_webhook = config.wecom_webhook || ''
    notificationForm.dingtalk_enabled = !!config.dingtalk_webhook
    notificationForm.wecom_enabled = !!config.wecom_webhook

    // 加载系统配置
    systemForm.log_retention_days = parseInt(config.log_retention_days) || 30
    systemForm.max_concurrency = parseInt(config.max_concurrency) || 5
    systemForm.execution_timeout_seconds = parseInt(config.execution_timeout_seconds) || 3600
  } catch (error) {
    ElMessage.error('加载配置失败: ' + error.message)
  }
}

async function saveSettings() {
  saving.value = true
  try {
    // 保存告警配置
    await api.updateConfig('dingtalk_webhook', notificationForm.dingtalk_enabled ? notificationForm.dingtalk_webhook : '')
    await api.updateConfig('wecom_webhook', notificationForm.wecom_enabled ? notificationForm.wecom_webhook : '')

    // 保存系统配置
    await api.updateConfig('log_retention_days', systemForm.log_retention_days.toString())
    await api.updateConfig('max_concurrency', systemForm.max_concurrency.toString())
    await api.updateConfig('execution_timeout_seconds', systemForm.execution_timeout_seconds.toString())

    ElMessage.success('设置已保存')
  } catch (error) {
    ElMessage.error('保存失败: ' + error.message)
  } finally {
    saving.value = false
  }
}

function resetForm() {
  loadSettings()
  ElMessage.info('已重置为上次保存的配置')
}

async function loadEnvironments() {
  try {
    environments.value = await taskStore.loadEnvironments()
  } catch (error) {
    ElMessage.error('加载环境列表失败: ' + error.message)
  }
}

async function refreshEnvironments() {
  refreshing.value = true
  try {
    await loadEnvironments()
    ElMessage.success('环境列表已刷新')
  } catch (error) {
    ElMessage.error('刷新失败: ' + error.message)
  } finally {
    refreshing.value = false
  }
}

function testNotification(type) {
  const webhook = type === 'dingtalk' ? notificationForm.dingtalk_webhook : notificationForm.wecom_webhook

  if (!webhook) {
    ElMessage.warning('请先配置Webhook URL')
    return
  }

  // SG-013: 调用后端测试通知 API
  api.testNotification(type, webhook)
    .then(() => {
      ElMessage.success('测试通知已发送，请检查群消息')
    })
    .catch((error) => {
      ElMessage.error('测试失败: ' + (error?.message || String(error)))
    })
}
</script>

<style lang="scss" scoped>
.settings-page {
  max-width: 1000px;
  margin: 0 auto;

  .page-header {
    margin-bottom: 24px;
    padding-bottom: 20px;
    border-bottom: 1px solid var(--border-color);

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

  .settings-tabs {
    background: linear-gradient(135deg, #1e293b 0%, #0f172a 100%);
    border: 1px solid var(--border-color);
    border-radius: 12px;
    padding: 24px;

    .tab-content {
      padding: 20px 0;

      .settings-form {
        .form-section {
          margin-bottom: 24px;

          h3 {
            font-size: 16px;
            font-weight: 600;
            margin-bottom: 16px;
            color: var(--text-primary);
          }

          .form-help {
            display: block;
            margin-top: 8px;
            font-size: 12px;
            color: var(--text-secondary);
          }
        }
      }

      .env-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 12px;

        p {
          color: var(--text-secondary);
          font-size: 14px;
        }
      }

      .env-name {
        display: flex;
        align-items: center;
        gap: 8px;
        font-weight: 500;
      }

      .about-content {
        text-align: center;
        padding: 40px 20px;

        .app-logo {
          color: #6366f1;
          margin-bottom: 20px;
        }

        h2 {
          font-size: 32px;
          font-weight: 700;
          margin-bottom: 8px;
          background: linear-gradient(90deg, #6366f1, #8b5cf6);
          -webkit-background-clip: text;
          -webkit-text-fill-color: transparent;
        }

        .version {
          color: var(--text-secondary);
          font-size: 14px;
          margin-bottom: 12px;
        }

        .description {
          color: var(--text-primary);
          font-size: 16px;
          margin-bottom: 32px;
        }

        .features,
        .tech-stack {
          text-align: left;
          max-width: 600px;
          margin: 0 auto 32px;

          h3 {
            font-size: 18px;
            font-weight: 600;
            margin-bottom: 16px;
            color: var(--text-primary);
          }

          ul {
            padding-left: 24px;

            li {
              color: var(--text-secondary);
              margin-bottom: 8px;
              line-height: 1.6;
            }
          }
        }

        .tech-stack {
          text-align: center;

          .el-tag {
            margin: 4px;
          }
        }

        .links {
          margin: 32px 0;

          .el-button {
            margin: 0 8px;
          }
        }

        .copyright {
          color: var(--text-secondary);
          font-size: 12px;
          margin-top: 40px;
        }
      }
    }
  }

  .settings-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    margin-top: 24px;
    padding-top: 20px;
    border-top: 1px solid var(--border-color);
  }
}

:root {
  --text-primary: #e2e8f0;
  --text-secondary: #94a3b8;
  --border-color: #334155;
}
</style>
