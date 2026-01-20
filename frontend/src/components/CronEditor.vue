<template>
  <div class="schedule-picker">
    <!-- 执行频率选择 -->
    <div class="frequency-selector">
      <label class="label">{{ isChinese ? '执行频率' : 'Frequency' }}</label>
      <el-radio-group v-model="frequency" @change="handleFrequencyChange" size="large">
        <el-radio-button value="specific">{{ isChinese ? '指定时间' : 'Specific Time' }}</el-radio-button>
        <el-radio-button value="daily">{{ isChinese ? '每天' : 'Daily' }}</el-radio-button>
        <el-radio-button value="weekly">{{ isChinese ? '每周' : 'Weekly' }}</el-radio-button>
        <el-radio-button value="monthly">{{ isChinese ? '每月' : 'Monthly' }}</el-radio-button>
        <el-radio-button value="interval">{{ isChinese ? '间隔时间' : 'Interval' }}</el-radio-button>
      </el-radio-group>
    </div>

    <!-- 指定时间配置 -->
    <div v-if="frequency === 'specific'" class="config-section">
      <label class="label">{{ isChinese ? '执行时间（北京时间）' : 'Execution Time (Beijing Time)' }}</label>
      <div style="display: flex; gap: 12px; align-items: center; flex-wrap: wrap">
        <el-time-picker
          v-model="specificTime"
          format="HH:mm:ss"
          value-format="HH:mm:ss"
          :placeholder="isChinese ? '选择时间' : 'Select time'"
          @change="generateCron"
          style="width: 200px"
        />
        <el-button @click="setCurrentTime" size="small">
          <el-icon><Clock /></el-icon>
          {{ isChinese ? '使用当前时间' : 'Use Current Time' }}
        </el-button>
      </div>
      <div style="margin-top: 12px; font-size: 13px; color: var(--text-secondary)">
        {{ isChinese ? '当前北京时间：' : 'Current Beijing Time: ' }}{{ currentBeijingTime }}
      </div>
    </div>

    <!-- 每天配置 -->
    <div v-if="frequency === 'daily'" class="config-section">
      <label class="label">{{ isChinese ? '执行时间（北京时间，支持秒）' : 'Execution Time (Beijing Time)' }}</label>
      <el-time-picker
        v-model="dailyTime"
        format="HH:mm:ss"
        value-format="HH:mm:ss"
        :placeholder="isChinese ? '选择时间' : 'Select time'"
        @change="generateCron"
        style="width: 200px"
      />
      <div style="margin-top: 8px; font-size: 13px; color: var(--text-secondary)">
        {{ isChinese ? '当前北京时间：' : 'Current Beijing Time: ' }}{{ currentBeijingTime }}
      </div>
    </div>

    <!-- 每周配置 -->
    <div v-else-if="frequency === 'weekly'" class="config-section">
      <label class="label">{{ isChinese ? '选择星期' : 'Select Days' }}</label>
      <el-checkbox-group v-model="weekDays" @change="generateCron">
        <el-checkbox-button v-for="day in weekOptions" :key="day.value" :value="day.value">
          {{ day.label }}
        </el-checkbox-button>
      </el-checkbox-group>
      <label class="label" style="margin-top: 16px">{{ isChinese ? '执行时间（北京时间）' : 'Execution Time (Beijing Time)' }}</label>
      <el-time-picker
        v-model="weeklyTime"
        format="HH:mm:ss"
        value-format="HH:mm:ss"
        :placeholder="isChinese ? '选择时间' : 'Select time'"
        @change="generateCron"
        style="width: 200px"
      />
    </div>

    <!-- 每月配置 -->
    <div v-else-if="frequency === 'monthly'" class="config-section">
      <label class="label">{{ isChinese ? '选择日期' : 'Select Day' }}</label>
      <el-select v-model="monthDay" :placeholder="isChinese ? '选择日期' : 'Select day'" @change="generateCron" style="width: 200px">
        <el-option v-for="day in 31" :key="day" :label="isChinese ? `每月${day}号` : `Day ${day}`" :value="day" />
      </el-select>
      <label class="label" style="margin-top: 16px">{{ isChinese ? '执行时间（北京时间）' : 'Execution Time (Beijing Time)' }}</label>
      <el-time-picker
        v-model="monthlyTime"
        format="HH:mm:ss"
        value-format="HH:mm:ss"
        :placeholder="isChinese ? '选择时间' : 'Select time'"
        @change="generateCron"
        style="width: 200px"
      />
    </div>

    <!-- 间隔时间配置 -->
    <div v-else-if="frequency === 'interval'" class="config-section">
      <label class="label">{{ isChinese ? '间隔设置' : 'Interval Settings' }}</label>
      <div style="display: flex; gap: 12px; align-items: center">
        <span>{{ isChinese ? '每' : 'Every' }}</span>
        <el-input-number
          v-model="intervalValue"
          :min="1"
          :max="999"
          @change="generateCron"
          style="width: 120px"
        />
        <el-select v-model="intervalUnit" @change="generateCron" style="width: 120px">
          <el-option :label="isChinese ? '分钟' : 'Minutes'" value="minute" />
          <el-option :label="isChinese ? '小时' : 'Hours'" value="hour" />
          <el-option :label="isChinese ? '天' : 'Days'" value="day" />
        </el-select>
        <span v-if="isChinese">执行一次</span>
      </div>
    </div>

    <!-- 执行说明 -->
    <div class="schedule-preview">
      <el-alert :title="scheduleDescription" type="info" :closable="false">
        <template #default>
          <div style="margin-top: 8px; font-size: 12px; opacity: 0.8">
            {{ isChinese ? 'Cron表达式: ' : 'Cron Expression: ' }}<el-tag size="small">{{ cronExpr }}</el-tag>
          </div>
        </template>
      </el-alert>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, computed, onMounted, onUnmounted } from 'vue'
import { Clock } from '@element-plus/icons-vue'
import { useLanguageStore } from '@/stores/language'

const props = defineProps({
  modelValue: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['update:modelValue'])

const langStore = useLanguageStore()
const isChinese = computed(() => langStore.isChinese)

const frequency = ref('specific')
const initializedFromModelValue = ref(false)

const specificHour = ref(11)
const specificMinute = ref(0)
const specificSecond = ref(0)
const currentBeijingTime = ref('')

const specificTime = computed({
  get: () => {
    const h = String(specificHour.value || 0).padStart(2, '0')
    const m = String(specificMinute.value || 0).padStart(2, '0')
    const s = String(specificSecond.value || 0).padStart(2, '0')
    return `${h}:${m}:${s}`
  },
  set: (val) => {
    const parts = String(val || '00:00:00').split(':')
    specificHour.value = parseInt(parts[0] || '0', 10) || 0
    specificMinute.value = parseInt(parts[1] || '0', 10) || 0
    specificSecond.value = parseInt(parts[2] || '0', 10) || 0
  }
})

const dailyTime = ref('09:00:00')
const weekDays = ref([1])
const weeklyTime = ref('09:00:00')
const monthDay = ref(1)
const monthlyTime = ref('09:00:00')
const intervalValue = ref(30)
const intervalUnit = ref('minute')
const cronExpr = ref('')

let timeInterval = null

const weekOptions = computed(() => isChinese.value ? [
  { label: '周一', value: 1 },
  { label: '周二', value: 2 },
  { label: '周三', value: 3 },
  { label: '周四', value: 4 },
  { label: '周五', value: 5 },
  { label: '周六', value: 6 },
  { label: '周日', value: 0 }
] : [
  { label: 'Mon', value: 1 },
  { label: 'Tue', value: 2 },
  { label: 'Wed', value: 3 },
  { label: 'Thu', value: 4 },
  { label: 'Fri', value: 5 },
  { label: 'Sat', value: 6 },
  { label: 'Sun', value: 0 }
])

const scheduleDescription = computed(() => {
  const unitTextZh = { minute: '分钟', hour: '小时', day: '天' }
  const unitTextEn = { minute: 'minute(s)', hour: 'hour(s)', day: 'day(s)' }

  switch (frequency.value) {
    case 'specific':
      const h = String(specificHour.value || 0).padStart(2, '0')
      const m = String(specificMinute.value || 0).padStart(2, '0')
      const s = String(specificSecond.value || 0).padStart(2, '0')
      return isChinese.value
        ? `每天 ${h}:${m}:${s}（北京时间）执行`
        : `Daily at ${h}:${m}:${s} (Beijing Time)`
    case 'daily':
      return isChinese.value
        ? `每天 ${formatTimeWithSeconds(dailyTime.value)}（北京时间）执行`
        : `Daily at ${formatTimeWithSeconds(dailyTime.value)} (Beijing Time)`
    case 'weekly':
      const days = weekDays.value.map(d => weekOptions.value.find(w => w.value === d)?.label).join(isChinese.value ? '、' : ', ')
      return isChinese.value
        ? `每周 ${days} 的 ${formatTimeWithSeconds(weeklyTime.value)}（北京时间）执行`
        : `Weekly on ${days} at ${formatTimeWithSeconds(weeklyTime.value)} (Beijing Time)`
    case 'monthly':
      return isChinese.value
        ? `每月 ${monthDay.value} 号 ${formatTimeWithSeconds(monthlyTime.value)}（北京时间）执行`
        : `Monthly on day ${monthDay.value} at ${formatTimeWithSeconds(monthlyTime.value)} (Beijing Time)`
    case 'interval':
      const unitText = isChinese.value ? unitTextZh[intervalUnit.value] : unitTextEn[intervalUnit.value]
      return isChinese.value
        ? `每 ${intervalValue.value} ${unitText} 执行一次`
        : `Every ${intervalValue.value} ${unitText}`
    default:
      return ''
  }
})

function formatTimeWithSeconds(value) {
  if (!value) return '00:00:00'
  if (typeof value === 'string') return value
  const h = String(value.getHours()).padStart(2, '0')
  const m = String(value.getMinutes()).padStart(2, '0')
  const s = String(value.getSeconds()).padStart(2, '0')
  return `${h}:${m}:${s}`
}

function getBeijingTime() {
  const now = new Date()
  const formatter = new Intl.DateTimeFormat('zh-CN', {
    timeZone: 'Asia/Shanghai',
    hour: 'numeric',
    minute: 'numeric',
    second: 'numeric',
    hour12: false
  })
  const parts = formatter.formatToParts(now)
  const h = parseInt(parts.find(p => p.type === 'hour')?.value || '0', 10)
  const m = parseInt(parts.find(p => p.type === 'minute')?.value || '0', 10)
  const s = parseInt(parts.find(p => p.type === 'second')?.value || '0', 10)

  return {
    getHours: () => h,
    getMinutes: () => m,
    getSeconds: () => s
  }
}

function updateCurrentTime() {
  const beijing = getBeijingTime()
  currentBeijingTime.value = formatTimeWithSeconds(beijing)
}

function setCurrentTime() {
  const beijing = getBeijingTime()
  specificHour.value = beijing.getHours()
  specificMinute.value = beijing.getMinutes()
  specificSecond.value = beijing.getSeconds()
  generateCron()
}

function handleFrequencyChange() {
  generateCron()
}

function formatTimeString(h, m, s) {
  const hh = String(h || 0).padStart(2, '0')
  const mm = String(m || 0).padStart(2, '0')
  const ss = String(s || 0).padStart(2, '0')
  return `${hh}:${mm}:${ss}`
}

function parseTimeString(val) {
  const parts = String(val || '00:00:00').split(':')
  const h = parseInt(parts[0] || '0', 10) || 0
  const m = parseInt(parts[1] || '0', 10) || 0
  const s = parseInt(parts[2] || '0', 10) || 0
  return { h, m, s }
}

function applyCronToState(expr) {
  const val = (expr || '').trim()
  if (!val) return false

  let match = val.match(/^0 \*\/(\d+) \* \* \* \*$/)
  if (match) {
    frequency.value = 'interval'
    intervalUnit.value = 'minute'
    intervalValue.value = parseInt(match[1], 10)
    return true
  }

  match = val.match(/^0 0 \*\/(\d+) \* \* \*$/)
  if (match) {
    frequency.value = 'interval'
    intervalUnit.value = 'hour'
    intervalValue.value = parseInt(match[1], 10)
    return true
  }

  match = val.match(/^0 0 0 \*\/(\d+) \* \*$/)
  if (match) {
    frequency.value = 'interval'
    intervalUnit.value = 'day'
    intervalValue.value = parseInt(match[1], 10)
    return true
  }

  match = val.match(/^(\d{1,2}) (\d{1,2}) (\d{1,2}) (\d{1,2}) \* \*$/)
  if (match) {
    frequency.value = 'monthly'
    const s = parseInt(match[1], 10)
    const m = parseInt(match[2], 10)
    const h = parseInt(match[3], 10)
    monthDay.value = parseInt(match[4], 10)
    monthlyTime.value = formatTimeString(h, m, s)
    return true
  }

  match = val.match(/^(\d{1,2}) (\d{1,2}) (\d{1,2}) \* \* ([0-6](?:,[0-6])*)$/)
  if (match) {
    frequency.value = 'weekly'
    const s = parseInt(match[1], 10)
    const m = parseInt(match[2], 10)
    const h = parseInt(match[3], 10)
    weekDays.value = match[4].split(',').map(x => parseInt(x, 10))
    weeklyTime.value = formatTimeString(h, m, s)
    return true
  }

  match = val.match(/^(\d{1,2}) (\d{1,2}) (\d{1,2}) \* \* \*$/)
  if (match) {
    frequency.value = 'specific'
    specificSecond.value = parseInt(match[1], 10)
    specificMinute.value = parseInt(match[2], 10)
    specificHour.value = parseInt(match[3], 10)
    dailyTime.value = formatTimeString(
      parseInt(match[3], 10),
      parseInt(match[2], 10),
      parseInt(match[1], 10)
    )
    return true
  }

  return false
}

function generateCron() {
  let cron = ''

  switch (frequency.value) {
    case 'specific': {
      const s = parseInt(specificSecond.value) || 0
      const m = parseInt(specificMinute.value) || 0
      const h = parseInt(specificHour.value) || 0
      cron = `${s} ${m} ${h} * * *`
      break
    }
    case 'daily': {
      const { h, m, s } = parseTimeString(dailyTime.value)
      cron = `${s} ${m} ${h} * * *`
      break
    }
    case 'weekly': {
      const { h, m, s } = parseTimeString(weeklyTime.value)
      const days = weekDays.value.sort((a, b) => a - b).join(',')
      cron = `${s} ${m} ${h} * * ${days}`
      break
    }
    case 'monthly': {
      const { h, m, s } = parseTimeString(monthlyTime.value)
      cron = `${s} ${m} ${h} ${monthDay.value} * *`
      break
    }
    case 'interval': {
      if (intervalUnit.value === 'minute') {
        cron = `0 */${intervalValue.value} * * * *`
      } else if (intervalUnit.value === 'hour') {
        cron = `0 0 */${intervalValue.value} * * *`
      } else if (intervalUnit.value === 'day') {
        cron = `0 0 0 */${intervalValue.value} * *`
      }
      break
    }
  }

  cronExpr.value = cron
  emit('update:modelValue', cron)
}

watch(() => props.modelValue, (val) => {
  cronExpr.value = val || ''
  if (val) {
    initializedFromModelValue.value = true
    applyCronToState(val)
  }
}, { immediate: true })

onMounted(() => {
  updateCurrentTime()
  timeInterval = setInterval(updateCurrentTime, 1000)

  if (!initializedFromModelValue.value) {
    const beijing = getBeijingTime()
    specificHour.value = beijing.getHours()
    specificMinute.value = beijing.getMinutes()
    specificSecond.value = beijing.getSeconds()
    generateCron()
  }
})

onUnmounted(() => {
  if (timeInterval) {
    clearInterval(timeInterval)
  }
})
</script>

<style lang="scss" scoped>
.schedule-picker {
  .frequency-selector {
    margin-bottom: 20px;
  }

  .config-section {
    margin-bottom: 20px;
    padding: 16px;
    background: rgba(99, 102, 241, 0.05);
    border-radius: 8px;
    border: 1px solid rgba(99, 102, 241, 0.2);
  }

  .label {
    display: block;
    margin-bottom: 8px;
    font-size: 14px;
    font-weight: 500;
    color: var(--text-primary);
  }

  .schedule-preview {
    margin-top: 20px;
  }

  :deep(.el-checkbox-button) {
    margin-right: 8px;
    margin-bottom: 8px;
  }
}
</style>
