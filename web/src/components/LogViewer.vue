<template>
  <div class="log-viewer">
    <div class="log-toolbar">
      <div class="log-status flex items-center gap-2">
        <span
          class="badge badge-sm"
          :class="{
            'badge-success': connectionStatus === 'connected',
            'badge-warning': connectionStatus === 'connecting',
            'badge-info': connectionStatus === 'disconnected'
          }"
        >
          {{ connectionStatusText }}
        </span>
        <span v-if="lineCount > 0" class="text-xs text-base-content/60">
          {{ lineCount }}/{{ maxLines }} 行
        </span>
      </div>
      <div class="log-actions flex items-center gap-2">
        <div class="join">
          <button class="btn btn-sm join-item" @click="togglePause">
            <Icon :icon="paused ? 'mdi:play' : 'mdi:pause'" class="text-lg" />
            {{ paused ? '恢复' : '暂停' }}
          </button>
          <button class="btn btn-sm join-item" @click="scrollToTop">
            <Icon icon="mdi:arrow-up" class="text-lg" />
          </button>
          <button class="btn btn-sm join-item" @click="scrollToBottom">
            <Icon icon="mdi:arrow-down" class="text-lg" />
          </button>
          <button class="btn btn-sm join-item" @click="clearLogs">
            <Icon icon="mdi:delete" class="text-lg" />
          </button>
        </div>
        <select v-model="tailLines" class="select select-sm select-bordered w-28" @change="reconnect">
          <option value="100">最新100行</option>
          <option value="500">最新500行</option>
          <option value="1000">最新1000行</option>
          <option value="all">全部</option>
        </select>
      </div>
    </div>
    <div ref="logContainer" class="log-container">
      <div v-if="logs.length === 0 && connectionStatus !== 'connecting'" class="log-empty">
        暂无日志
      </div>
      <div v-else class="log-content">
        <div
          v-for="(line, index) in logs"
          :key="index"
          :class="['log-line', line.type]"
        >
          {{ line.content }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { Icon } from '@iconify/vue'

interface LogLine {
  content: string
  type: 'stdout' | 'stderr' | 'system'
}

const props = defineProps<{
  wsUrl: string
  autoScroll?: boolean
  maxLines?: number
}>()

const emit = defineEmits<{
  (e: 'status-change', status: string): void
}>()

const logContainer = ref<HTMLElement | null>(null)
const logs = ref<LogLine[]>([])
const paused = ref(false)
const autoScroll = ref(props.autoScroll ?? true)
const connectionStatus = ref<'disconnected' | 'connecting' | 'connected'>('disconnected')
const tailLines = ref('100')
const maxLines = computed(() => {
  // 根据 tailLines 设置最大行数，防止内存溢出
  if (tailLines.value === 'all') {
    return props.maxLines ?? 5000 // 全部模式下默认最大5000行
  }
  return Math.max(parseInt(tailLines.value) * 2, 200) // 至少保留 tail 行数的2倍
})
const lineCount = computed(() => logs.value.length)

let ws: WebSocket | null = null
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
let pendingLogs: LogLine[] = []

const connectionStatusText = computed(() => {
  switch (connectionStatus.value) {
    case 'connected':
      return '已连接'
    case 'connecting':
      return '连接中...'
    default:
      return '已断开'
  }
})

function connect() {
  if (!props.wsUrl) return

  connectionStatus.value = 'connecting'
  emit('status-change', 'connecting')

  // 构建 WebSocket URL，添加 tail 参数
  let wsUrl = props.wsUrl
  const separator = wsUrl.includes('?') ? '&' : '?'
  wsUrl = `${wsUrl}${separator}tail=${tailLines.value}`

  try {
    ws = new WebSocket(wsUrl)

    ws.onopen = () => {
      connectionStatus.value = 'connected'
      emit('status-change', 'connected')
      if (pendingLogs.length > 0 && !paused.value) {
        logs.value.push(...pendingLogs)
        pendingLogs = []
        if (autoScroll.value) {
          scrollToBottom()
        }
      }
    }

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        if (data.type === 'log') {
          const logLine: LogLine = {
            content: data.content,
            type: data.stream === 'stderr' ? 'stderr' : 'stdout'
          }
          if (paused.value) {
            pendingLogs.push(logLine)
          } else {
            logs.value.push(logLine)
            trimLogs()
            if (autoScroll.value) {
              nextTick(() => scrollToBottom())
            }
          }
        } else if (data.type === 'ping') {
          // Heartbeat, ignore
        } else if (data.type === 'end') {
          connectionStatus.value = 'disconnected'
          emit('status-change', 'disconnected')
        } else if (data.type === 'error') {
          logs.value.push({
            content: `[错误] ${data.content}`,
            type: 'system'
          })
        }
      } catch {
        // Plain text message
        if (!paused.value) {
          logs.value.push({
            content: event.data,
            type: 'stdout'
          })
          trimLogs()
          if (autoScroll.value) {
            nextTick(() => scrollToBottom())
          }
        }
      }
    }

    ws.onerror = () => {
      connectionStatus.value = 'disconnected'
      emit('status-change', 'disconnected')
      logs.value.push({
        content: '[系统] WebSocket 连接错误',
        type: 'system'
      })
    }

    ws.onclose = () => {
      connectionStatus.value = 'disconnected'
      emit('status-change', 'disconnected')
    }
  } catch (error) {
    connectionStatus.value = 'disconnected'
    emit('status-change', 'disconnected')
    logs.value.push({
      content: `[系统] 无法连接到日志服务: ${error}`,
      type: 'system'
    })
  }
}

function disconnect() {
  if (ws) {
    ws.close()
    ws = null
  }
  if (reconnectTimer) {
    clearTimeout(reconnectTimer)
    reconnectTimer = null
  }
}

function reconnect() {
  disconnect()
  logs.value = []
  pendingLogs = []
  connect()
}

function togglePause() {
  paused.value = !paused.value
  if (!paused.value && pendingLogs.length > 0) {
    logs.value.push(...pendingLogs)
    trimLogs()
    pendingLogs = []
    if (autoScroll.value) {
      nextTick(() => scrollToBottom())
    }
  }
}

function scrollToTop() {
  if (logContainer.value) {
    logContainer.value.scrollTop = 0
    autoScroll.value = false
  }
}

function scrollToBottom() {
  if (logContainer.value) {
    logContainer.value.scrollTop = logContainer.value.scrollHeight
  }
}

function clearLogs() {
  logs.value = []
  pendingLogs = []
}

// 修剪日志，保持在最大行数限制内
function trimLogs() {
  if (logs.value.length > maxLines.value) {
    // 删除旧的日志行，保留最新的
    logs.value = logs.value.slice(-maxLines.value)
  }
}

// Watch for URL changes
watch(() => props.wsUrl, (newUrl, oldUrl) => {
  if (newUrl && newUrl !== oldUrl) {
    reconnect()
  }
})

onMounted(() => {
  if (props.wsUrl) {
    connect()
  }
})

onUnmounted(() => {
  disconnect()
})

// Expose methods for parent component
defineExpose({
  reconnect,
  clearLogs,
  scrollToTop,
  scrollToBottom
})
</script>

<style scoped>
.log-viewer {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 300px;
}

.log-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: var(--fallback-b2, oklch(var(--b2) / 1));
  border-bottom: 1px solid var(--fallback-bc, oklch(var(--bc) / 0.1));
  border-radius: 0.5rem 0.5rem 0 0;
}

.log-container {
  flex: 1;
  background: #1e1e1e;
  border-radius: 0 0 0.5rem 0.5rem;
  overflow: auto;
}

.log-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  min-height: 200px;
  color: #909399;
  font-size: 14px;
}

.log-content {
  padding: 12px 16px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace;
  font-size: 12px;
  line-height: 1.6;
}

.log-line {
  white-space: pre-wrap;
  word-break: break-all;
  color: #d4d4d4;
}

.log-line.stderr {
  color: #f48771;
}

.log-line.system {
  color: #569cd6;
  font-style: italic;
}
</style>
