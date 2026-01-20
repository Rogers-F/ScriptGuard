<template>
  <div class="page-container settings-page">
    <div class="page-header">
      <h1>{{ t.settings.title }}</h1>
    </div>

    <div class="settings-layout glass-panel">
      <el-tabs tab-position="left" class="custom-tabs" v-model="activeTab">
        <!-- 通用设置 -->
        <el-tab-pane :label="t.settings.tabs.general" name="general">
          <div class="setting-content">
            <h2>{{ t.settings.general.title }}</h2>
            <p class="subtitle">{{ t.settings.general.subtitle }}</p>

            <div class="form-group">
              <h3>{{ t.settings.general.language }}</h3>
              <p class="field-desc">{{ t.settings.general.languageDesc }}</p>
              <div class="language-selector">
                <el-radio-group v-model="selectedLanguage" @change="handleLanguageChange">
                  <el-radio-button value="zh">
                    <span class="lang-option">
                      <span class="lang-flag">🇨🇳</span>
                      {{ t.settings.general.chinese }}
                    </span>
                  </el-radio-button>
                  <el-radio-button value="en">
                    <span class="lang-option">
                      <span class="lang-flag">🇺🇸</span>
                      {{ t.settings.general.english }}
                    </span>
                  </el-radio-button>
                </el-radio-group>
              </div>
            </div>
          </div>
        </el-tab-pane>

        <!-- 告警通知 -->
        <el-tab-pane :label="t.settings.tabs.alerts" name="notification">
          <div class="setting-content">
            <h2>{{ t.settings.alerts.title }}</h2>
            <p class="subtitle">{{ t.settings.alerts.subtitle }}</p>

            <el-form label-position="top" class="clean-form">
              <div class="form-section">
                <div class="section-head">
                    <h3>{{ t.settings.alerts.dingtalk }}</h3>
                    <el-switch v-model="notificationForm.dingtalk_enabled" />
                </div>
                <el-form-item v-if="notificationForm.dingtalk_enabled" :label="t.settings.alerts.webhookUrl">
                  <el-input v-model="notificationForm.dingtalk_webhook" placeholder="https://oapi.dingtalk.com/...">
                    <template #append><el-button @click="testNotification('dingtalk')">{{ t.common.test }}</el-button></template>
                  </el-input>
                </el-form-item>
              </div>

              <el-divider />

              <div class="form-section">
                <div class="section-head">
                    <h3>{{ t.settings.alerts.wecom }}</h3>
                    <el-switch v-model="notificationForm.wecom_enabled" />
                </div>
                <el-form-item v-if="notificationForm.wecom_enabled" :label="t.settings.alerts.webhookUrl">
                  <el-input v-model="notificationForm.wecom_webhook" placeholder="https://qyapi.weixin.qq.com/...">
                    <template #append><el-button @click="testNotification('wecom')">{{ t.common.test }}</el-button></template>
                  </el-input>
                </el-form-item>
              </div>
            </el-form>
          </div>
        </el-tab-pane>

        <!-- 系统设置 -->
        <el-tab-pane :label="t.settings.tabs.system" name="system">
          <div class="setting-content">
            <h2>{{ t.settings.system.title }}</h2>

            <div class="form-group">
                <h3>{{ t.settings.system.retentionPolicy }}</h3>
                <el-form-item :label="t.settings.system.keepLogs">
                  <el-input-number v-model="systemForm.log_retention_days" :min="1" :max="365" />
                </el-form-item>
            </div>

            <div class="form-group">
                <h3>{{ t.settings.system.execution }}</h3>
                <el-form-item :label="t.settings.system.maxConcurrency">
                  <el-input-number v-model="systemForm.max_concurrency" :min="1" :max="20" />
                </el-form-item>
                <el-form-item :label="t.settings.system.timeout">
                   <el-input-number v-model="systemForm.execution_timeout_seconds" :min="0" :step="60" />
                </el-form-item>
            </div>
          </div>
        </el-tab-pane>

        <!-- 环境管理 -->
        <el-tab-pane :label="t.settings.tabs.environments" name="environment">
          <div class="setting-content">
            <h2>{{ t.settings.environments.title }}</h2>
            <p class="subtitle">{{ t.settings.environments.subtitle }}</p>

            <div class="env-header">
              <el-button type="primary" @click="refreshEnvironments" :loading="refreshing">
                {{ t.settings.environments.refreshList }}
              </el-button>
            </div>

            <el-table :data="environments" style="width: 100%; margin-top: 20px">
              <el-table-column prop="name" :label="t.settings.environments.name" width="200">
                <template #default="{ row }">
                  <span class="font-medium">{{ row.name }}</span>
                </template>
              </el-table-column>
              <el-table-column prop="path" :label="t.settings.environments.path" show-overflow-tooltip />
              <el-table-column prop="is_valid" :label="t.settings.environments.status" width="100">
                <template #default="{ row }">
                  <el-tag :type="row.is_valid ? 'success' : 'danger'" size="small" effect="light">
                    {{ row.is_valid ? t.common.valid : t.common.invalid }}
                  </el-tag>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-tab-pane>

        <!-- 关于 -->
        <el-tab-pane :label="t.settings.tabs.about" name="about">
          <div class="setting-content about-view">
             <div class="logo">S</div>
             <h2>ScriptGuard</h2>
             <span class="version">v1.3.0</span>
             <p class="desc">{{ t.settings.about.desc }}</p>
          </div>
        </el-tab-pane>
      </el-tabs>

      <div class="settings-actions" v-if="activeTab === 'notification' || activeTab === 'system'">
          <el-button @click="resetForm">{{ t.common.reset }}</el-button>
          <el-button type="primary" @click="saveSettings" :loading="saving">{{ t.settings.saveChanges }}</el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { useTaskStore } from '@/stores/task'
import { useLanguageStore } from '@/stores/language'
import api from '@/api'

const taskStore = useTaskStore()
const langStore = useLanguageStore()
const t = computed(() => langStore.t)

const activeTab = ref('general')
const saving = ref(false)
const refreshing = ref(false)
const environments = ref([])
const selectedLanguage = ref(langStore.currentLang)

const notificationForm = reactive({ dingtalk_enabled: false, dingtalk_webhook: '', wecom_enabled: false, wecom_webhook: '' })
const systemForm = reactive({ log_retention_days: 30, max_concurrency: 5, execution_timeout_seconds: 3600 })

onMounted(async () => {
  await loadSettings()
  await loadEnvironments()
})

function handleLanguageChange(lang) {
  langStore.setLanguage(lang)
  ElMessage.success(lang === 'zh' ? '语言已切换为中文' : 'Language changed to English')
}

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
  } catch (err) { ElMessage.error(t.value.settings.loadFailed) }
}

async function loadEnvironments() {
  try {
    environments.value = await taskStore.loadEnvironments()
  } catch (error) {
    ElMessage.error(t.value.settings.loadFailed)
  }
}

async function refreshEnvironments() {
  refreshing.value = true
  try {
    await loadEnvironments()
    ElMessage.success(t.value.settings.environments.refreshed)
  } catch (error) {
    ElMessage.error(t.value.settings.refreshFailed)
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
    ElMessage.success(t.value.settings.saved)
  } catch (err) { ElMessage.error(err.message) } finally { saving.value = false }
}

function resetForm() { loadSettings() }
function testNotification(type) {
  const webhook = type === 'dingtalk' ? notificationForm.dingtalk_webhook : notificationForm.wecom_webhook
  if (!webhook) return ElMessage.warning(t.value.settings.alerts.urlRequired)
  api.testNotification(type, webhook).then(() => ElMessage.success(t.value.settings.alerts.testSent)).catch(e => ElMessage.error(e))
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
        .field-desc { color: var(--text-tertiary); font-size: 13px; margin-bottom: 16px; }

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

        .language-selector {
          margin-top: 8px;

          .lang-option {
            display: flex;
            align-items: center;
            gap: 8px;

            .lang-flag {
              font-size: 18px;
            }
          }

          :deep(.el-radio-button__inner) {
            padding: 12px 24px;
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
