<template>
  <div class="page-container settings-page">
    <div class="page-header">
      <h1>系统设置</h1>
    </div>

    <div class="settings-layout">
      <el-tabs tab-position="left" class="custom-tabs" v-model="activeTab">
        <el-tab-pane label="通知配置" name="notification">
          <div class="setting-panel">
            <h2>消息通知</h2>
            <p class="subtitle">配置脚本执行失败时的告警通道</p>

            <el-form label-position="top" class="mt-6">
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
                </el-form-item>
              </div>
            </el-form>
          </div>
        </el-tab-pane>

        <el-tab-pane label="系统参数" name="system">
          <div class="setting-panel">
            <h2>运行参数</h2>
            <p class="subtitle">配置系统运行时的各项参数</p>

            <el-form label-position="top" class="mt-6">
              <div class="form-section">
                <h3>日志管理</h3>
                <el-form-item label="日志保留天数">
                  <el-input-number
                    v-model="systemForm.log_retention_days"
                    :min="1"
                    :max="365"
                  />
                  <span class="form-help">超过此天数的日志将被自动清理</span>
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
                  <span class="form-help">同时执行的任务数量上限</span>
                </el-form-item>

                <el-form-item label="执行超时（秒）">
                  <el-input-number
                    v-model="systemForm.execution_timeout_seconds"
                    :min="0"
                    :max="86400"
                    :step="60"
                  />
                  <span class="form-help">0 表示不限制；允许范围 0 或 60~86400 秒</span>
                </el-form-item>
              </div>
            </el-form>
          </div>
        </el-tab-pane>

        <el-tab-pane label="环境管理" name="environment">
          <div class="setting-panel">
            <h2>Conda 环境</h2>
            <div class="env-header">
              <p class="subtitle">管理系统中可用的Conda环境</p>
              <el-button type="primary" @click="refreshEnvironments" :loading="refreshing">
                刷新列表
              </el-button>
            </div>

            <el-table :data="environments" style="width: 100%; margin-top: 20px">
              <el-table-column prop="name" label="环境名称" width="200">
                <template #default="{ row }">
                  <span class="font-medium">{{ row.name }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="path" label="路径" show-overflow-tooltip />
              <el-table-column prop="python_path" label="Python路径" show-overflow-tooltip />
              <el-table-column prop="is_valid" label="状态" width="100">
                <template #default="{ row }">
                  <el-tag :type="row.is_valid ? 'success' : 'danger'" size="small" effect="dark">
                    {{ row.is_valid ? '有效' : '无效' }}
                  </el-tag>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-tab-pane>

        <el-tab-pane label="关于" name="about">
          <div class="setting-panel about-content">
            <div class="app-logo">
              <el-icon :size="48"><Monitor /></el-icon>
            </div>
            <h2>ScriptGuard</h2>
            <p class="version">版本 1.2.0</p>
            <p class="description">Python脚本监控与定时执行系统</p>

            <el-divider />

            <div class="features">
              <h3>主要特性</h3>
              <ul>
                <li>支持Cron表达式的灵活定时任务</li>
                <li>多Conda环境隔离执行</li>
                <li>实时日志监控与持久化</li>
                <li>执行历史统计分析</li>
                <li>钉钉/企业微信告警通知</li>
              </ul>
            </div>

            <el-divider />

            <div class="tech-stack">
              <h3>技术栈</h3>
              <div class="tags">
                <el-tag effect="plain">Wails 3</el-tag>
                <el-tag effect="plain">Go 1.23</el-tag>
                <el-tag effect="plain">Vue 3</el-tag>
                <el-tag effect="plain">Element Plus</el-tag>
                <el-tag effect="plain">SQLite</el-tag>
              </div>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>

    <!-- Footer -->
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
import { Monitor } from '@element-plus/icons-vue'
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

    notificationForm.dingtalk_webhook = config.dingtalk_webhook || ''
    notificationForm.wecom_webhook = config.wecom_webhook || ''
    notificationForm.dingtalk_enabled = !!config.dingtalk_webhook
    notificationForm.wecom_enabled = !!config.wecom_webhook

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
    await api.updateConfig('dingtalk_webhook', notificationForm.dingtalk_enabled ? notificationForm.dingtalk_webhook : '')
    await api.updateConfig('wecom_webhook', notificationForm.wecom_enabled ? notificationForm.wecom_webhook : '')
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
  max-width: 1200px;

  .settings-layout {
    background: var(--bg-secondary);
    border: 1px solid var(--border-light);
    border-radius: 8px;
    min-height: 600px;

    :deep(.el-tabs__header) {
      background: var(--bg-tertiary);
      border-right: 1px solid var(--border-light);
    }

    :deep(.el-tabs__item) {
      color: var(--text-secondary);
      padding: 0 24px;
      height: 48px;

      &.is-active {
        color: var(--color-primary);
        background: var(--bg-secondary);
      }

      &:hover {
        color: var(--text-primary);
      }
    }

    :deep(.el-tabs__nav-wrap::after) {
      display: none;
    }
  }

  .setting-panel {
    padding: 32px 40px;
    max-width: 700px;

    h2 {
      font-size: 20px;
      font-weight: 600;
      margin-bottom: 8px;
      color: var(--text-primary);
    }

    .subtitle {
      color: var(--text-secondary);
      font-size: 14px;
      margin-bottom: 24px;
    }

    .form-section {
      margin-bottom: 24px;

      h3 {
        font-size: 15px;
        font-weight: 500;
        margin-bottom: 16px;
        color: var(--text-primary);
      }
    }

    .form-help {
      display: block;
      margin-top: 6px;
      font-size: 12px;
      color: var(--text-tertiary);
    }

    .env-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 16px;
    }

    &.about-content {
      text-align: center;
      padding: 48px 40px;

      .app-logo {
        color: var(--color-primary);
        margin-bottom: 16px;
      }

      h2 {
        font-size: 28px;
        margin-bottom: 8px;
      }

      .version {
        color: var(--text-tertiary);
        font-size: 14px;
        margin-bottom: 8px;
      }

      .description {
        color: var(--text-secondary);
        margin-bottom: 32px;
      }

      .features,
      .tech-stack {
        text-align: left;
        max-width: 500px;
        margin: 0 auto 24px;

        h3 {
          font-size: 16px;
          font-weight: 500;
          margin-bottom: 12px;
        }

        ul {
          padding-left: 20px;

          li {
            color: var(--text-secondary);
            margin-bottom: 6px;
            line-height: 1.6;
          }
        }
      }

      .tech-stack {
        text-align: center;

        .tags {
          display: flex;
          justify-content: center;
          flex-wrap: wrap;
          gap: 8px;
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
    border-top: 1px solid var(--border-light);
  }
}

.mt-6 { margin-top: 24px; }
.font-medium { font-weight: 500; }
</style>
