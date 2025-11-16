<template>
  <div class="schedule-picker">
    <!-- 执行频率选择 -->
    <div class="frequency-selector">
      <label class="label">执行频率</label>
      <el-radio-group v-model="frequency" @change="handleFrequencyChange" size="large">
        <el-radio-button value="specific">指定时间</el-radio-button>
        <el-radio-button value="daily">每天</el-radio-button>
        <el-radio-button value="weekly">每周</el-radio-button>
        <el-radio-button value="monthly">每月</el-radio-button>
        <el-radio-button value="interval">间隔时间</el-radio-button>
      </el-radio-group>
    </div>

    <!-- 指定时间配置（新增）-->
    <div v-if="frequency === 'specific'" class="config-section">
      <label class="label">执行时间（北京时间）</label>
      <div style="display: flex; gap: 12px; align-items: center; flex-wrap: wrap">
        <el-input
          v-model="specificHour"
          placeholder="时"
          type="number"
          :min="0"
          :max="23"
          @change="generateCron"
          style="width: 80px"
        >
          <template #append>时</template>
        </el-input>
        <span>:</span>
        <el-input
          v-model="specificMinute"
          placeholder="分"
          type="number"
          :min="0"
          :max="59"
          @change="generateCron"
          style="width: 80px"
        >
          <template #append>分</template>
        </el-input>
        <span>:</span>
        <el-input
          v-model="specificSecond"
          placeholder="秒"
          type="number"
          :min="0"
          :max="59"
          @change="generateCron"
          style="width: 80px"
        >
          <template #append>秒</template>
        </el-input>
        <el-button @click="setCurrentTime" size="small">
          <el-icon><Clock /></el-icon>
          使用当前时间
        </el-button>
      </div>
      <div style="margin-top: 12px; font-size: 13px; color: var(--text-secondary)">
        💡 示例：11:00:53 表示每天上午11点00分53秒执行
      </div>
    </div>

    <!-- 每天配置 -->
    <div v-if="frequency === 'daily'" class="config-section">
      <label class="label">执行时间（北京时间，支持秒）</label>
      <el-time-picker
        v-model="dailyTime"
        format="HH:mm:ss"
        placeholder="选择时间"
        @change="generateCron"
        style="width: 200px"
      />
      <div style="margin-top: 8px; font-size: 13px; color: var(--text-secondary)">
        🕐 当前北京时间：{{ currentBeijingTime }}
      </div>
    </div>

    <!-- 每周配置 -->
    <div v-else-if="frequency === 'weekly'" class="config-section">
      <label class="label">选择星期</label>
      <el-checkbox-group v-model="weekDays" @change="generateCron">
        <el-checkbox-button v-for="day in weekOptions" :key="day.value" :value="day.value">
          {{ day.label }}
        </el-checkbox-button>
      </el-checkbox-group>
      <label class="label" style="margin-top: 16px">执行时间（北京时间）</label>
      <el-time-picker
        v-model="weeklyTime"
        format="HH:mm:ss"
        placeholder="选择时间"
        @change="generateCron"
        style="width: 200px"
      />
    </div>

    <!-- 每月配置 -->
    <div v-else-if="frequency === 'monthly'" class="config-section">
      <label class="label">选择日期</label>
      <el-select v-model="monthDay" placeholder="选择日期" @change="generateCron" style="width: 200px">
        <el-option v-for="day in 31" :key="day" :label="`每月${day}号`" :value="day" />
      </el-select>
      <label class="label" style="margin-top: 16px">执行时间（北京时间）</label>
      <el-time-picker
        v-model="monthlyTime"
        format="HH:mm:ss"
        placeholder="选择时间"
        @change="generateCron"
        style="width: 200px"
      />
    </div>

    <!-- 间隔时间配置 -->
    <div v-else-if="frequency === 'interval'" class="config-section">
      <label class="label">间隔设置</label>
      <div style="display: flex; gap: 12px; align-items: center">
        <span>每</span>
        <el-input-number
          v-model="intervalValue"
          :min="1"
          :max="999"
          @change="generateCron"
          style="width: 120px"
        />
        <el-select v-model="intervalUnit" @change="generateCron" style="width: 120px">
          <el-option label="分钟" value="minute" />
          <el-option label="���时" value="hour" />
          <el-option label="天" value="day" />
        </el-select>
        <span>执行一次</span>
      </div>
    </div>


    <!-- 执行说明 -->
    <div class="schedule-preview">
      <el-alert :title="scheduleDescription" type="info" :closable="false">
        <template #default>
          <div style="margin-top: 8px; font-size: 12px; opacity: 0.8">
            Cron表达式: <el-tag size="small">{{ cronExpr }}</el-tag>
          </div>
        </template>
      </el-alert>
    </div>
  </div>
</template>

<script setup>
import { ref, watch, computed, onMounted, onUnmounted } from 'vue'
import { Clock } from '@element-plus/icons-vue'

const props = defineProps({
  modelValue: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['update:modelValue'])

const frequency = ref('specific')
// 新增：指定时间模式
const specificHour = ref(11)
const specificMinute = ref(0)
const specificSecond = ref(0)
const currentBeijingTime = ref('')

const dailyTime = ref(new Date('2024-01-01 09:00:00'))
const weekDays = ref([1]) // 周一
const weeklyTime = ref(new Date('2024-01-01 09:00:00'))
const monthDay = ref(1)
const monthlyTime = ref(new Date('2024-01-01 09:00:00'))
const intervalValue = ref(30)
const intervalUnit = ref('minute')
const cronExpr = ref('')

let timeInterval = null

const weekOptions = [
  { label: '周一', value: 1 },
  { label: '周二', value: 2 },
  { label: '周三', value: 3 },
  { label: '周四', value: 4 },
  { label: '周五', value: 5 },
  { label: '周六', value: 6 },
  { label: '周日', value: 0 }
]

const scheduleDescription = computed(() => {
  switch (frequency.value) {
    case 'specific':
      const h = String(specificHour.value || 0).padStart(2, '0')
      const m = String(specificMinute.value || 0).padStart(2, '0')
      const s = String(specificSecond.value || 0).padStart(2, '0')
      return `每天 ${h}:${m}:${s}（北京时间）执行`
    case 'daily':
      return `每天 ${formatTimeWithSeconds(dailyTime.value)}（北京时间）执行`
    case 'weekly':
      const days = weekDays.value.map(d => weekOptions.find(w => w.value === d)?.label).join('、')
      return `每周 ${days} 的 ${formatTimeWithSeconds(weeklyTime.value)}（北京时间）执行`
    case 'monthly':
      return `每月 ${monthDay.value} 号 ${formatTimeWithSeconds(monthlyTime.value)}（北京时间）执行`
    case 'interval':
      const unitText = { minute: '分钟', hour: '小时', day: '天' }[intervalUnit.value]
      return `每 ${intervalValue.value} ${unitText} 执行一次`
    default:
      return ''
  }
})

function formatTimeWithSeconds(date) {
  if (!date) return '00:00:00'
  const h = String(date.getHours()).padStart(2, '0')
  const m = String(date.getMinutes()).padStart(2, '0')
  const s = String(date.getSeconds()).padStart(2, '0')
  return `${h}:${m}:${s}`
}

// 获取北京时间（UTC+8）
function getBeijingTime() {
  const now = new Date()
  // 获取UTC时间
  const utc = now.getTime() + (now.getTimezoneOffset() * 60000)
  // 转换为北京时间（UTC+8）
  const beijingTime = new Date(utc + (3600000 * 8))
  return beijingTime
}

// 更新当前北京时间显示
function updateCurrentTime() {
  const beijing = getBeijingTime()
  currentBeijingTime.value = formatTimeWithSeconds(beijing)
}

// 使用当前北京时间
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

function generateCron() {
  let cron = ''

  switch (frequency.value) {
    case 'specific': {
      // 新增：指定时间模式（支持秒）
      const s = parseInt(specificSecond.value) || 0
      const m = parseInt(specificMinute.value) || 0
      const h = parseInt(specificHour.value) || 0
      // 格式：秒 分 时 日 月 周
      cron = `${s} ${m} ${h} * * *`
      break
    }
    case 'daily': {
      const s = dailyTime.value.getSeconds()
      const m = dailyTime.value.getMinutes()
      const h = dailyTime.value.getHours()
      cron = `${s} ${m} ${h} * * *`
      break
    }
    case 'weekly': {
      const s = weeklyTime.value.getSeconds()
      const m = weeklyTime.value.getMinutes()
      const h = weeklyTime.value.getHours()
      const days = weekDays.value.sort((a, b) => a - b).join(',')
      cron = `${s} ${m} ${h} * * ${days}`
      break
    }
    case 'monthly': {
      const s = monthlyTime.value.getSeconds()
      const m = monthlyTime.value.getMinutes()
      const h = monthlyTime.value.getHours()
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

// 初始化时生成cron
watch(() => props.modelValue, (val) => {
  if (val) {
    cronExpr.value = val
    // TODO: 可以添加反向解析逻辑
  }
}, { immediate: true })

// 生命周期
onMounted(() => {
  // 初始化时使用当前北京时间
  const beijing = getBeijingTime()
  specificHour.value = beijing.getHours()
  specificMinute.value = beijing.getMinutes()
  specificSecond.value = beijing.getSeconds()

  // 更新当前时间显示
  updateCurrentTime()
  timeInterval = setInterval(updateCurrentTime, 1000)

  // 生成初始Cron
  generateCron()
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
