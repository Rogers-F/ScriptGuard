<template>
  <div class="page-container settings-page">
    <div class="page-header">
      <h1>Settings</h1>
    </div>

    <div class="settings-layout glass-panel">
      <el-tabs tab-position="left" class="custom-tabs" v-model="activeTab">
        <el-tab-pane label="Alerts" name="notification">
          <div class="setting-content">
            <h2>Notifications</h2>
            <p class="subtitle">Configure alert channels for task failures.</p>

            <el-form label-position="top" class="clean-form">
              <div class="form-section">
                <div class="section-head">
                    <h3>DingTalk Robot</h3>
                    <el-switch v-model="notificationForm.dingtalk_enabled" />
                </div>
                <el-form-item v-if="notificationForm.dingtalk_enabled" label="Webhook URL">
                  <el-input v-model="notificationForm.dingtalk_webhook" placeholder="https://oapi.dingtalk.com/...">
                    <template #append><el-button @click="testNotification('dingtalk')">Test</el-button></template>
                  </el-input>
                </el-form-item>
              </div>

              <el-divider />

              <div class="form-section">
                <div class="section-head">
                    <h3>WeCom Robot</h3>
                    <el-switch v-model="notificationForm.wecom_enabled" />
                </div>
                <el-form-item v-if="notificationForm.wecom_enabled" label="Webhook URL">
                  <el-input v-model="notificationForm.wecom_webhook" placeholder="https://qyapi.weixin.qq.com/...">
                    <template #append><el-button @click="testNotification('wecom')">Test</el-button></template>
                  </el-input>
                </el-form-item>
              </div>
            </el-form>
          </div>
        </el-tab-pane>

        <el-tab-pane label="System" name="system">
          <div class="setting-content">
            <h2>System</h2>

            <div class="form-group">
                <h3>Retention Policy</h3>
                <el-form-item label="Keep logs for (days)">
                  <el-input-number v-model="systemForm.log_retention_days" :min="1" :max="365" />
                </el-form-item>
            </div>

            <div class="form-group">
                <h3>Execution</h3>
                <el-form-item label="Max Concurrency">
                  <el-input-number v-model="systemForm.max_concurrency" :min="1" :max="20" />
                </el-form-item>
                <el-form-item label="Timeout (Seconds)">
                   <el-input-number v-model="systemForm.execution_timeout_seconds" :min="0" :step="60" />
                </el-form-item>
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane label="Environments" name="environment">
          <div class="setting-content">
            <h2>Conda Environments</h2>
            <p class="subtitle">Available Python environments for task execution.</p>

            <div class="env-header">
              <el-button type="primary" @click="refreshEnvironments" :loading="refreshing">
                Refresh List
              </el-button>
            </div>

            <el-table :data="environments" style="width: 100%; margin-top: 20px">
              <el-table-column prop="name" label="Name" width="200">
                <template #default="{ row }">
                  <span class="font-medium">{{ row.name }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="path" label="Path" show-overflow-tooltip />
              <el-table-column prop="is_valid" label="Status" width="100">
                <template #default="{ row }">
                  <el-tag :type="row.is_valid ? 'success' : 'danger'" size="small" effect="light">
                    {{ row.is_valid ? 'Valid' : 'Invalid' }}
                  </el-tag>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-tab-pane>

        <el-tab-pane label="About" name="about">
          <div class="setting-content about-view">
             <div class="logo">S</div>
             <h2>ScriptGuard</h2>
             <span class="version">v1.2.0</span>
             <p class="desc">Secure task automation for the modern web.</p>
          </div>
        </el-tab-pane>
      </el-tabs>

      <div class="settings-actions" v-if="activeTab !== 'about' && activeTab !== 'environment'">
          <el-button @click="resetForm">Reset</el-button>
          <el-button type="primary" @click="saveSettings" :loading="saving">Save Changes</el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { useTaskStore } from '@/stores/task'
import api from '@/api'

const taskStore = useTaskStore()

const activeTab = ref('notification')
const saving = ref(false)
const refreshing = ref(false)
const environments = ref([])

const notificationForm = reactive({ dingtalk_enabled: false, dingtalk_webhook: '', wecom_enabled: false, wecom_webhook: '' })
const systemForm = reactive({ log_retention_days: 30, max_concurrency: 5, execution_timeout_seconds: 3600 })

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
  } catch (err) { ElMessage.error('Load failed') }
}

async function loadEnvironments() {
  try {
    environments.value = await taskStore.loadEnvironments()
  } catch (error) {
    ElMessage.error('Failed to load environments')
  }
}

async function refreshEnvironments() {
  refreshing.value = true
  try {
    await loadEnvironments()
    ElMessage.success('Environments refreshed')
  } catch (error) {
    ElMessage.error('Refresh failed')
  } finally {
    refreshing.value = false
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
    ElMessage.success('Saved')
  } catch (err) { ElMessage.error(err.message) } finally { saving.value = false }
}

function resetForm() { loadSettings() }
function testNotification(type) {
  const webhook = type === 'dingtalk' ? notificationForm.dingtalk_webhook : notificationForm.wecom_webhook
  if (!webhook) return ElMessage.warning('URL required')
  api.testNotification(type, webhook).then(() => ElMessage.success('Sent')).catch(e => ElMessage.error(e))
}
</script>

<style lang="scss" scoped>
.settings-page {
  max-width: 900px;

  .settings-layout {
    min-height: 500px;
    display: flex; flex-direction: column;

    .custom-tabs {
       flex: 1;
       :deep(.el-tabs__header) { margin-right: 0; background: rgba(0,0,0,0.02); width: 200px; padding-top: 24px; }
       :deep(.el-tabs__item) {
           height: 48px; color: var(--text-secondary); font-family: var(--font-sans);
           &.is-active { color: var(--color-primary); background: rgba(255,255,255,0.6); }
       }
    }

    .setting-content {
        padding: 32px 48px;
        h2 { margin-bottom: 8px; font-size: 24px; }
        .subtitle { color: var(--text-tertiary); margin-bottom: 32px; }

        .section-head {
            display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;
            h3 { font-size: 16px; font-weight: 600; font-family: var(--font-sans); }
        }

        .form-group {
            margin-bottom: 32px;
            h3 {
                font-size: 14px;
                font-family: var(--font-sans);
                color: var(--text-primary);
                font-weight: 600;
                margin-bottom: 16px;
                border-bottom: 1px solid var(--border-light);
                padding-bottom: 8px;
            }
        }

        .env-header {
          display: flex;
          justify-content: flex-end;
          margin-bottom: 16px;
        }

        &.about-view {
            text-align: center; padding-top: 60px;
            .logo {
                width: 64px; height: 64px; background: var(--color-primary); color: white;
                font-family: var(--font-serif); font-size: 32px; display: flex; align-items: center; justify-content: center;
                border-radius: 12px; margin: 0 auto 24px;
            }
            .version {
                display: inline-block; background: var(--border-light); padding: 4px 12px;
                border-radius: 20px; font-size: 12px; margin: 12px 0;
            }
            .desc { color: var(--text-secondary); }
        }
    }

    .settings-actions {
        padding: 16px 48px; border-top: 1px solid var(--border-light);
        display: flex; justify-content: flex-end; gap: 12px;
    }
  }
}
</style>
