<template>
  <div class="schedule-picker">
    <!-- 执行频率选择 -->
    <div class="frequency-selector">
      <label class="label">{{ isChinese ? '执行频率' : 'Frequency' }}</label>
      <el-radio-group v-model="frequency" @change="handleFrequencyChange" size="large">
        <el-radio-button value="daily">{{ isChinese ? '每天' : 'Daily' }}</el-radio-button>
        <el-radio-button value="weekly">{{ isChinese ? '每周' : 'Weekly' }}</el-radio-button>
        <el-radio-button value="monthly">{{ isChinese ? '每月' : 'Monthly' }}</el-radio-button>
        <el-radio-button value="interval">{{ isChinese ? '间隔时间' : 'Interval' }}</el-radio-button>
      </el-radio-group>
    </div>

    <!-- 每天/每周/每月 共用的时间点列表 -->
    <div v-if="frequency !== 'interval'" class="config-section">
      <!-- 每周：选择星期 -->
      <template v-if="frequency === 'weekly'">
        <label class="label">{{ isChinese ? '选择星期' : 'Select Days' }}</label>
        <el-checkbox-group v-model="weekDays" @change="handleWeekDaysChange">
          <el-checkbox-button
            v-for="day in weekOptions"
            :key="day.value"
            :value="day.value"
            :disabled="weekDays.length === 1 && weekDays.includes(day.value)"
          >
            {{ day.label }}
          </el-checkbox-button>
        </el-checkbox-group>
      </template>

      <!-- 每月：选择日期 -->
      <template v-if="frequency === 'monthly'">
        <label class="label">{{ isChinese ? '选择日期' : 'Select Day' }}</label>
        <el-select v-model="monthDay" :placeholder="isChinese ? '选择日期' : 'Select day'" @change="generateCronExprs" style="width: 200px">
          <el-option v-for="day in 31" :key="day" :label="isChinese ? `每月${day}号` : `Day ${day}`" :value="day" />
        </el-select>
      </template>

      <!-- 时间点列表管理 -->
      <label class="label" :style="frequency !== 'daily' ? 'margin-top: 16px' : ''">
        {{ isChinese ? `执行时间点（北京时间，最多${maxTimePoints}个）` : `Time Points (Beijing, up to ${maxTimePoints})` }}
      </label>

      <div class="time-point-input">
        <el-time-picker
          v-model="timeToAdd"
          format="HH:mm:ss"
          value-format="HH:mm:ss"
          :placeholder="isChinese ? '选择时间' : 'Select time'"
          style="width: 200px"
        />
        <el-button
          type="primary"
          @click="addTimePoint(timeToAdd)"
          :disabled="timePoints.length >= maxTimePoints"
        >
          {{ isChinese ? '添加' : 'Add' }}
        </el-button>
        <el-button @click="addCurrentBeijingTime" :disabled="timePoints.length >= maxTimePoints">
          <el-icon><Clock /></el-icon>
          {{ isChinese ? '添加当前时间' : 'Add Current' }}
        </el-button>
        <span class="time-count">
          {{ timePoints.length }}/{{ maxTimePoints }}
        </span>
      </div>

      <!-- 已添加的时间点 -->
      <div class="time-points-list">
        <el-tag
          v-for="t in timePoints"
          :key="t"
          :closable="timePoints.length > 1"
          @close="removeTimePoint(t)"
          type="info"
          effect="plain"
        >
          {{ t }}
        </el-tag>
      </div>

      <div class="current-time-hint">
        {{ isChinese ? '当前北京时间：' : 'Current Beijing Time: ' }}{{ currentBeijingTime }}
      </div>
    </div>

    <!-- 间隔时间配置（不支持多时间点） -->
    <div v-else class="config-section">
      <label class="label">{{ isChinese ? '间隔设置' : 'Interval Settings' }}</label>
      <div style="display: flex; gap: 12px; align-items: center">
        <span>{{ isChinese ? '每' : 'Every' }}</span>
        <el-input-number
          v-model="intervalValue"
          :min="1"
          :max="999"
          @change="generateCronExprs"
          style="width: 120px"
        />
        <el-select v-model="intervalUnit" @change="generateCronExprs" style="width: 120px">
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
            {{ isChinese ? 'Cron表达式: ' : 'Cron Expression: ' }}
            <el-tag v-for="(expr, idx) in cronExprs.slice(0, 3)" :key="idx" size="small" style="margin-right: 4px">
              {{ expr }}
            </el-tag>
            <span v-if="cronExprs.length > 3">... {{ isChinese ? `共${cronExprs.length}条` : `${cronExprs.length} total` }}</span>
          </div>
        </template>
      </el-alert>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, computed, onMounted, onUnmounted } from 'vue'
import { Clock } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import { useLanguageStore } from '@/stores/language'

const props = defineProps({
  modelValue: {
    type: [Array, String],
    default: () => []
  }
})

const emit = defineEmits(['update:modelValue'])

const langStore = useLanguageStore()
const isChinese = computed(() => langStore.isChinese)

const maxTimePoints = 60
const frequency = ref('daily')
const initializedFromModelValue = ref(false)

// 时间点列表
const timeToAdd = ref('09:00:00')
const timePoints = ref(['09:00:00'])
const currentBeijingTime = ref('')

// 规则参数
const weekDays = ref([1])
const monthDay = ref(1)
const intervalValue = ref(30)
const intervalUnit = ref('minute')

// 输出的 cron 表达式数组
const cronExprs = ref([])

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

  if (frequency.value === 'interval') {
    const unitText = isChinese.value ? unitTextZh[intervalUnit.value] : unitTextEn[intervalUnit.value]
    return isChinese.value
      ? `每 ${intervalValue.value} ${unitText} 执行一次`
      : `Every ${intervalValue.value} ${unitText}`
  }

  const timesStr = timePoints.value.slice(0, 5).join(isChinese.value ? '、' : ', ')
  const suffix = timePoints.value.length > 5 ? (isChinese.value ? ` 等${timePoints.value.length}个时间点` : ` and ${timePoints.value.length - 5} more`) : ''

  switch (frequency.value) {
    case 'daily':
      return isChinese.value
        ? `每天 ${timesStr}${suffix}（北京时间）执行`
        : `Daily at ${timesStr}${suffix} (Beijing Time)`
    case 'weekly':
      const days = weekDays.value.map(d => weekOptions.value.find(w => w.value === d)?.label).join(isChinese.value ? '、' : ', ')
      return isChinese.value
        ? `每周 ${days} 的 ${timesStr}${suffix}（北京时间）执行`
        : `Weekly on ${days} at ${timesStr}${suffix} (Beijing Time)`
    case 'monthly':
      return isChinese.value
        ? `每月 ${monthDay.value} 号 ${timesStr}${suffix}（北京时间）执行`
        : `Monthly on day ${monthDay.value} at ${timesStr}${suffix} (Beijing Time)`
    default:
      return ''
  }
})

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

function toSeconds(t) {
  const { h, m, s } = parseTimeString(t)
  return h * 3600 + m * 60 + s
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

  return { h, m, s }
}

function updateCurrentTime() {
  const beijing = getBeijingTime()
  currentBeijingTime.value = formatTimeString(beijing.h, beijing.m, beijing.s)
}

function addTimePoint(value) {
  if (timePoints.value.length >= maxTimePoints) {
    ElMessage.warning(isChinese.value ? `最多支持 ${maxTimePoints} 个时间点` : `Up to ${maxTimePoints} time points`)
    return
  }
  const { h, m, s } = parseTimeString(value)
  const normalized = formatTimeString(h, m, s)

  if (timePoints.value.includes(normalized)) {
    ElMessage.warning(isChinese.value ? '该时间点已存在' : 'This time point already exists')
    return
  }

  timePoints.value = [...timePoints.value, normalized].sort((a, b) => toSeconds(a) - toSeconds(b))
  generateCronExprs()
}

function removeTimePoint(value) {
  if (timePoints.value.length <= 1) {
    ElMessage.warning(isChinese.value ? '至少保留 1 个时间点' : 'At least 1 time point required')
    return
  }
  timePoints.value = timePoints.value.filter(x => x !== value)
  generateCronExprs()
}

function addCurrentBeijingTime() {
  const beijing = getBeijingTime()
  addTimePoint(formatTimeString(beijing.h, beijing.m, beijing.s))
}

function handleFrequencyChange() {
  generateCronExprs()
}

// 处理周天选择变化，确保至少选择一天
function handleWeekDaysChange(val) {
  if (val.length === 0) {
    // 恢复到周一
    weekDays.value = [1]
    ElMessage.warning(isChinese.value ? '至少选择一天' : 'At least one day required')
    return
  }
  generateCronExprs()
}

function generateCronExprs() {
  let exprs = []

  // 兜底检查：周模式必须至少选择一天（静默恢复，handleWeekDaysChange 已处理用户提示）
  if (frequency.value === 'weekly' && weekDays.value.length === 0) {
    weekDays.value = [1]
  }

  if (frequency.value === 'interval') {
    // interval：不支持多时间点，固定输出 1 条
    let cron = ''
    if (intervalUnit.value === 'minute') cron = `0 */${intervalValue.value} * * * *`
    else if (intervalUnit.value === 'hour') cron = `0 0 */${intervalValue.value} * * *`
    else cron = `0 0 0 */${intervalValue.value} * *`
    exprs = [cron]
  } else {
    const days = weekDays.value.slice().sort((a, b) => a - b).join(',')

    for (const t of timePoints.value) {
      const { h, m, s } = parseTimeString(t)

      if (frequency.value === 'weekly') {
        exprs.push(`${s} ${m} ${h} * * ${days}`)
      } else if (frequency.value === 'monthly') {
        exprs.push(`${s} ${m} ${h} ${monthDay.value} * *`)
      } else {
        // daily
        exprs.push(`${s} ${m} ${h} * * *`)
      }
    }
  }

  cronExprs.value = exprs
  emit('update:modelValue', exprs)
}

// 正则匹配
const reDaily = /^(\d{1,2}) (\d{1,2}) (\d{1,2}) \* \* \*$/
const reWeekly = /^(\d{1,2}) (\d{1,2}) (\d{1,2}) \* \* ([0-6](?:,[0-6])*)$/
const reMonthly = /^(\d{1,2}) (\d{1,2}) (\d{1,2}) (\d{1,2}) \* \*$/

function normalizeModelValue(val) {
  if (Array.isArray(val)) return val.map(x => String(x || '').trim()).filter(Boolean)
  const s = String(val || '').trim()
  return s ? [s] : []
}

function normalizeDays(daysStr) {
  return daysStr
    .split(',')
    .map(x => parseInt(x, 10))
    .filter(x => !Number.isNaN(x))
    .sort((a, b) => a - b)
}

function applyCronExprsToState(input) {
  const exprs = normalizeModelValue(input)
  if (exprs.length === 0) return false

  // interval（仅单条）
  if (exprs.length === 1) {
    let m = exprs[0].match(/^0 \*\/(\d+) \* \* \* \*$/)
    if (m) { frequency.value = 'interval'; intervalUnit.value = 'minute'; intervalValue.value = parseInt(m[1], 10); return true }

    m = exprs[0].match(/^0 0 \*\/(\d+) \* \* \*$/)
    if (m) { frequency.value = 'interval'; intervalUnit.value = 'hour'; intervalValue.value = parseInt(m[1], 10); return true }

    m = exprs[0].match(/^0 0 0 \*\/(\d+) \* \*$/)
    if (m) { frequency.value = 'interval'; intervalUnit.value = 'day'; intervalValue.value = parseInt(m[1], 10); return true }
  }

  // monthly：所有表达式必须同 day-of-month
  const m0 = exprs[0].match(reMonthly)
  if (m0 && exprs.every(e => (e.match(reMonthly) || [])[4] === m0[4])) {
    frequency.value = 'monthly'
    monthDay.value = parseInt(m0[4], 10)
    timePoints.value = exprs
      .map(e => {
        const m = e.match(reMonthly)
        return formatTimeString(parseInt(m[3], 10), parseInt(m[2], 10), parseInt(m[1], 10))
      })
      .sort((a, b) => toSeconds(a) - toSeconds(b))
    return true
  }

  // weekly：所有表达式必须同 weekdays
  const w0 = exprs[0].match(reWeekly)
  if (w0) {
    const baseDays = normalizeDays(w0[4]).join(',')
    const ok = exprs.every(e => {
      const m = e.match(reWeekly)
      return m && normalizeDays(m[4]).join(',') === baseDays
    })
    if (ok) {
      frequency.value = 'weekly'
      weekDays.value = normalizeDays(w0[4])
      timePoints.value = exprs
        .map(e => {
          const m = e.match(reWeekly)
          return formatTimeString(parseInt(m[3], 10), parseInt(m[2], 10), parseInt(m[1], 10))
        })
        .sort((a, b) => toSeconds(a) - toSeconds(b))
      return true
    }
  }

  // daily：全部是 daily 形态
  if (exprs.every(e => reDaily.test(e))) {
    frequency.value = 'daily'
    timePoints.value = exprs
      .map(e => {
        const m = e.match(reDaily)
        return formatTimeString(parseInt(m[3], 10), parseInt(m[2], 10), parseInt(m[1], 10))
      })
      .sort((a, b) => toSeconds(a) - toSeconds(b))
    return true
  }

  return false
}

watch(() => props.modelValue, (val) => {
  if (val && (Array.isArray(val) ? val.length > 0 : val.trim())) {
    initializedFromModelValue.value = true
    applyCronExprsToState(val)
  }
}, { immediate: true })

onMounted(() => {
  updateCurrentTime()
  timeInterval = setInterval(updateCurrentTime, 1000)

  if (!initializedFromModelValue.value) {
    const beijing = getBeijingTime()
    timePoints.value = [formatTimeString(beijing.h, beijing.m, beijing.s)]
    generateCronExprs()
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
    background: rgba(217, 119, 6, 0.05);
    border-radius: 8px;
    border: 1px solid rgba(217, 119, 6, 0.2);
  }

  .label {
    display: block;
    margin-bottom: 8px;
    font-size: 14px;
    font-weight: 500;
    color: var(--text-primary);
  }

  .time-point-input {
    display: flex;
    gap: 12px;
    align-items: center;
    flex-wrap: wrap;
  }

  .time-count {
    font-size: 12px;
    color: var(--text-tertiary);
  }

  .time-points-list {
    margin-top: 12px;
    display: flex;
    flex-wrap: wrap;
    gap: 8px;

    .el-tag {
      font-family: var(--font-mono);
    }
  }

  .current-time-hint {
    margin-top: 12px;
    font-size: 13px;
    color: var(--text-secondary);
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
