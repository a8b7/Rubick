<template>
  <div class="container-terminal flex flex-col h-full">
    <!-- 终端配置 -->
    <div v-if="!connected" class="mb-4">
      <div class="card bg-base-100 shadow">
        <div class="card-body">
          <h3 class="card-title text-base mb-4">终端配置</h3>

          <div class="grid grid-cols-2 gap-4">
            <div class="form-control">
              <label class="label">
                <span class="label-text">命令</span>
              </label>
              <select v-model="config.command" class="select select-bordered select-sm">
                <option value="/bin/sh">/bin/sh</option>
                <option value="/bin/bash">/bin/bash</option>
                <option value="/bin/ash">/bin/ash</option>
              </select>
            </div>

            <div class="form-control">
              <label class="label">
                <span class="label-text">用户</span>
              </label>
              <input
                v-model="config.user"
                type="text"
                placeholder="root"
                class="input input-bordered input-sm"
              />
            </div>
          </div>

          <div class="card-actions mt-4">
            <button
              class="btn btn-primary btn-sm"
              @click="connectTerminal"
              :disabled="connecting"
            >
              <span v-if="connecting" class="loading loading-spinner loading-sm"></span>
              <Icon v-else icon="mdi:console" class="text-lg" />
              连接终端
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 终端区域 -->
    <div v-show="connected" class="terminal-wrapper flex-1 flex flex-col">
      <div class="terminal-toolbar flex items-center justify-between px-3 py-2 bg-base-100 rounded-t-lg border border-base-content/10 border-b-0 shrink-0">
        <div class="flex items-center gap-2 text-sm">
          <Icon icon="mdi:console" class="text-lg text-primary" />
          <span>{{ config.command }}</span>
        </div>
        <button class="btn btn-xs btn-ghost" @click="disconnect">
          <Icon icon="mdi:close" class="text-lg" />
          断开
        </button>
      </div>
      <div ref="terminalContainer" class="terminal-container flex-1"></div>
    </div>

    <!-- 容器未运行提示 -->
    <div v-if="!isRunning && !connected" class="flex-1 flex flex-col items-center justify-center text-base-content/50">
      <Icon icon="mdi:cube-off-outline" class="text-4xl mb-2" />
      <p>容器未运行，无法打开终端</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { Icon } from '@iconify/vue'
import { containerApi, type Container } from '@/api'
import { showToast } from '@/utils/toast'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'

const props = defineProps<{
  hostId: string
  containerId: string
}>()

const container = ref<Container | null>(null)
const loading = ref(true)
const connecting = ref(false)
const connected = ref(false)

const config = reactive({
  command: '/bin/sh',
  user: '',
})

const terminalContainer = ref<HTMLElement | null>(null)
let terminal: Terminal | null = null
let fitAddon: FitAddon | null = null
let ws: WebSocket | null = null

const isRunning = computed(() => container.value?.state === 'running')

async function loadContainer() {
  loading.value = true
  try {
    const res = await containerApi.get(props.hostId, props.containerId)
    container.value = res.data.data
  } catch {
    container.value = null
  } finally {
    loading.value = false
  }
}

async function connectTerminal() {
  if (!isRunning.value) {
    showToast('容器未运行', 'warning')
    return
  }

  connecting.value = true

  try {
    // 创建 exec 实例
    const execRes = await containerApi.createExec(props.hostId, props.containerId, {
      cmd: [config.command],
      user: config.user || undefined,
      tty: true,
      stdin: true,
      stdout: true,
      stderr: true,
    })

    const execId = execRes.data.data.id

    // 初始化终端
    await nextTick()

    terminal = new Terminal({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'Monaco, Menlo, "Courier New", monospace',
      theme: {
        background: '#1e1e1e',
      },
    })

    fitAddon = new FitAddon()
    terminal.loadAddon(fitAddon)

    if (terminalContainer.value) {
      terminal.open(terminalContainer.value)
      fitAddon.fit()
    }

    // 连接 WebSocket
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const wsUrl = `${protocol}//${host}/api/v1/ws/containers/${props.containerId}/exec?host_id=${props.hostId}&exec_id=${execId}`

    ws = new WebSocket(wsUrl)

    ws.onopen = () => {
      connected.value = true
      connecting.value = false

      // 发送初始大小
      if (terminal) {
        ws?.send(JSON.stringify({
          type: 'resize',
          cols: terminal.cols,
          rows: terminal.rows,
        }))
      }
    }

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        if (data.type === 'data') {
          terminal?.write(data.content)
        }
      } catch {
        terminal?.write(event.data)
      }
    }

    ws.onerror = () => {
      showToast('终端连接失败', 'error')
      disconnect()
    }

    ws.onclose = () => {
      connected.value = false
    }

    // 监听终端输入
    terminal.onData((data) => {
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({
          type: 'input',
          content: data,
        }))
      }
    })

    // 监听终端大小变化
    terminal.onResize(({ cols, rows }) => {
      if (ws && ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({
          type: 'resize',
          cols,
          rows,
        }))
      }
    })

    // 监听窗口大小变化
    window.addEventListener('resize', handleResize)

  } catch (error) {
    showToast('创建终端失败', 'error')
    connecting.value = false
  }
}

function handleResize() {
  if (fitAddon && terminal) {
    fitAddon.fit()
  }
}

function disconnect() {
  if (ws) {
    ws.close()
    ws = null
  }
  if (terminal) {
    terminal.dispose()
    terminal = null
  }
  connected.value = false
  window.removeEventListener('resize', handleResize)
}

// 当容器 ID 变化时断开连接
watch(() => props.containerId, () => {
  disconnect()
  loadContainer()
})

onMounted(loadContainer)

onUnmounted(() => {
  disconnect()
})
</script>

<style scoped>
.container-terminal {
  height: 100%;
}

.terminal-container {
  background: #1e1e1e;
  border: 1px solid oklch(var(--bc) / 0.1);
  border-radius: 0 0 0.5rem 0.5rem;
  padding: 8px;
}
</style>
