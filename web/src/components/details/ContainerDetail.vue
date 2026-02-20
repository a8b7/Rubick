<template>
  <div class="container-detail">
    <div v-if="loading" class="flex justify-center py-8">
      <span class="loading loading-spinner loading-lg"></span>
    </div>

    <template v-else-if="container">
      <!-- 操作按钮 -->
      <div class="flex items-center gap-2 mb-4">
        <button
          v-if="container.state !== 'running'"
          class="btn btn-sm btn-success"
          @click="handleStart"
          :disabled="operating"
        >
          <Icon icon="mdi:play" class="text-lg" />
          启动
        </button>
        <button
          v-if="container.state === 'running'"
          class="btn btn-sm btn-warning"
          @click="handleStop"
          :disabled="operating"
        >
          <Icon icon="mdi:stop" class="text-lg" />
          停止
        </button>
        <button
          v-if="container.state === 'running'"
          class="btn btn-sm btn-info"
          @click="handleRestart"
          :disabled="operating"
        >
          <Icon icon="mdi:restart" class="text-lg" />
          重启
        </button>
        <button
          class="btn btn-sm btn-error"
          @click="handleDelete"
          :disabled="operating"
        >
          <Icon icon="mdi:delete" class="text-lg" />
          删除
        </button>
      </div>

      <!-- 基本信息 -->
      <div class="card bg-base-100 shadow">
        <div class="card-body">
          <h3 class="card-title text-base mb-4">基本信息</h3>
          <div class="overflow-x-auto">
            <table class="table table-sm">
              <tbody>
                <tr>
                  <td class="font-medium w-32">ID</td>
                  <td class="font-mono text-sm">{{ container.id }}</td>
                </tr>
                <tr>
                  <td class="font-medium">名称</td>
                  <td>{{ container.name }}</td>
                </tr>
                <tr>
                  <td class="font-medium">镜像</td>
                  <td>{{ container.image }}</td>
                </tr>
                <tr>
                  <td class="font-medium">状态</td>
                  <td>
                    <span
                      class="badge"
                      :class="container.state === 'running' ? 'badge-success' : 'badge-error'"
                    >
                      {{ container.state }}
                    </span>
                    <span v-if="container.status" class="text-sm text-base-content/60 ml-2">
                      {{ container.status }}
                    </span>
                  </td>
                </tr>
                <tr>
                  <td class="font-medium">创建时间</td>
                  <td>{{ formatDate(container.created) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- 端口映射 -->
      <div v-if="container.ports?.length" class="card bg-base-100 shadow mt-4">
        <div class="card-body">
          <h3 class="card-title text-base mb-4">端口映射</h3>
          <div class="overflow-x-auto">
            <table class="table table-sm">
              <thead>
                <tr>
                  <th>容器端口</th>
                  <th>主机端口</th>
                  <th>协议</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(port, index) in container.ports" :key="index">
                  <td>{{ port.private_port }}</td>
                  <td>{{ port.public_port || '-' }} ({{ port.ip || '0.0.0.0' }})</td>
                  <td>{{ port.type }}</td>
                  <td>
                    <button
                      v-if="port.public_port"
                      class="btn btn-xs btn-ghost"
                      @click="copyPort(port.ip || '0.0.0.0', port.public_port)"
                      title="复制端口地址"
                    >
                      <Icon icon="mdi:content-copy" class="text-base" />
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- 标签 -->
      <div v-if="Object.keys(container.labels || {}).length" class="card bg-base-100 shadow mt-4">
        <div class="card-body">
          <h3 class="card-title text-base mb-4">标签</h3>
          <div class="flex flex-wrap gap-2">
            <span
              v-for="(value, key) in container.labels"
              :key="key"
              class="badge badge-outline"
            >
              {{ key }}={{ value }}
            </span>
          </div>
        </div>
      </div>
    </template>

    <div v-else class="text-center py-8 text-base-content/50">
      <Icon icon="mdi:cube-off-outline" class="text-4xl mb-2" />
      <p>无法加载容器信息</p>
    </div>

    <Confirm ref="confirmRef" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { Icon } from '@iconify/vue'
import { containerApi, type Container } from '@/api'
import { showToast } from '@/utils/toast'
import { useResourceStore } from '@/stores'
import Confirm from '@/components/Confirm.vue'

const props = defineProps<{
  hostId: string
  containerId: string
}>()

const emit = defineEmits<{
  (e: 'refresh'): void
}>()

const resourceStore = useResourceStore()

const container = ref<Container | null>(null)
const loading = ref(false)
const operating = ref(false)
const confirmRef = ref<InstanceType<typeof Confirm> | null>(null)

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

async function handleStart() {
  operating.value = true
  try {
    await containerApi.start(props.hostId, props.containerId)
    showToast('容器已启动', 'success')
    await loadContainer()
    resourceStore.refreshHostResources(props.hostId)
  } finally {
    operating.value = false
  }
}

async function handleStop() {
  operating.value = true
  try {
    await containerApi.stop(props.hostId, props.containerId)
    showToast('容器已停止', 'success')
    await loadContainer()
    resourceStore.refreshHostResources(props.hostId)
  } finally {
    operating.value = false
  }
}

async function handleRestart() {
  operating.value = true
  try {
    await containerApi.restart(props.hostId, props.containerId)
    showToast('容器已重启', 'success')
    await loadContainer()
    resourceStore.refreshHostResources(props.hostId)
  } finally {
    operating.value = false
  }
}

async function handleDelete() {
  const result = await confirmRef.value?.confirm({
    title: '删除确认',
    message: `确定删除容器 "${container.value?.name}"?`,
    type: 'danger',
    options: [
      { key: 'removeImage', label: '同时删除关联镜像', default: false }
    ]
  })
  if (result && result.confirmed) {
    operating.value = true
    try {
      const removeImage = result.options?.removeImage || false
      await containerApi.remove(props.hostId, props.containerId, true, false, removeImage)
      showToast(removeImage ? '容器和镜像已删除' : '容器已删除', 'success')
      resourceStore.refreshHostResources(props.hostId)
      emit('refresh')
    } finally {
      operating.value = false
    }
  }
}

function formatDate(timestamp: number): string {
  return new Date(timestamp * 1000).toLocaleString()
}

function copyPort(ip: string, port: number) {
  const address = `${ip}:${port}`
  navigator.clipboard.writeText(address).then(() => {
    showToast(`已复制: ${address}`, 'success')
  }).catch(() => {
    showToast('复制失败', 'error')
  })
}

watch(() => props.containerId, loadContainer)
onMounted(loadContainer)
</script>
