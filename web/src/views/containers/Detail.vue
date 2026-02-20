<template>
  <div class="page-container">
    <div class="card bg-base-100 shadow-xl">
      <div class="card-body">
        <div class="flex items-center gap-4 mb-4">
          <button class="btn btn-ghost" @click="$router.back()">
            <Icon icon="mdi:arrow-left" class="text-xl" />
            返回
          </button>
          <h2 class="card-title">容器详情: {{ container?.name }}</h2>
        </div>

        <div v-if="container" class="overflow-x-auto">
          <table class="table">
            <tbody>
              <tr>
                <td class="font-medium w-32">ID</td>
                <td>{{ container.id }}</td>
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
                  <span class="badge" :class="container.state === 'running' ? 'badge-success' : 'badge-error'">
                    {{ container.state }}
                  </span>
                </td>
              </tr>
              <tr>
                <td class="font-medium">创建时间</td>
                <td>{{ new Date(container.created * 1000).toLocaleString() }}</td>
              </tr>
              <tr>
                <td class="font-medium">状态信息</td>
                <td>{{ container.status }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- 挂载卷信息 -->
    <div class="card bg-base-100 shadow-xl mt-4">
      <div class="card-body">
        <h3 class="card-title mb-4">
          <Icon icon="mdi:folder-outline" class="mr-2" />
          挂载卷
        </h3>
        <div v-if="container?.mounts && container.mounts.length > 0" class="overflow-x-auto">
          <table class="table table-zebra">
            <thead>
              <tr>
                <th>类型</th>
                <th>源路径</th>
                <th>目标路径</th>
                <th>模式</th>
                <th>读写</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="mount in container.mounts" :key="mount.destination">
                <td>
                  <span class="badge badge-outline">{{ mount.type }}</span>
                </td>
                <td class="font-mono text-sm">{{ mount.source }}</td>
                <td class="font-mono text-sm">{{ mount.destination }}</td>
                <td>{{ mount.mode || '-' }}</td>
                <td>
                  <span class="badge" :class="mount.rw ? 'badge-success' : 'badge-warning'">
                    {{ mount.rw ? 'RW' : 'RO' }}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div v-else class="text-center py-8 text-base-content/60">
          <Icon icon="mdi:folder-off-outline" class="text-4xl mb-2" />
          <p>无挂载卷</p>
        </div>
      </div>
    </div>

    <!-- 网络信息 -->
    <div class="card bg-base-100 shadow-xl mt-4">
      <div class="card-body">
        <h3 class="card-title mb-4">
          <Icon icon="mdi:network-outline" class="mr-2" />
          网络
        </h3>
        <div v-if="container?.networks && container.networks.length > 0" class="overflow-x-auto">
          <table class="table table-zebra">
            <thead>
              <tr>
                <th>网络名称</th>
                <th>IP 地址</th>
                <th>MAC 地址</th>
                <th>网关</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="network in container.networks" :key="network.network_id">
                <td>
                  <span class="badge badge-primary badge-outline">{{ network.name }}</span>
                </td>
                <td class="font-mono text-sm">{{ network.ip_address || '-' }}</td>
                <td class="font-mono text-sm">{{ network.mac_address || '-' }}</td>
                <td class="font-mono text-sm">{{ network.gateway || '-' }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <div v-else class="text-center py-8 text-base-content/60">
          <Icon icon="mdi:network-off-outline" class="text-4xl mb-2" />
          <p>无网络连接</p>
        </div>
      </div>
    </div>

    <div class="card bg-base-100 shadow-xl mt-4">
      <div class="card-body">
        <div class="flex justify-between items-center mb-4">
          <h3 class="card-title">实时日志</h3>
          <span v-if="!isRunning" class="badge badge-info">容器未运行</span>
        </div>
        <LogViewer
          v-if="isRunning && wsUrl"
          :ws-url="wsUrl"
          :auto-scroll="true"
        />
        <div v-else class="log-disabled">
          <div class="text-center py-16">
            <Icon icon="mdi:text-box-outline" class="text-6xl text-base-content/40 mb-4" />
            <p class="text-base-content/60 mb-4">容器未运行，无法查看实时日志</p>
            <button class="btn btn-primary" @click="handleStart">启动容器</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { Icon } from '@iconify/vue'
import { useHostStore } from '@/stores'
import { containerApi, type Container } from '@/api'
import LogViewer from '@/components/LogViewer.vue'

const route = useRoute()
const hostStore = useHostStore()
const container = ref<Container | null>(null)

const isRunning = computed(() => container.value?.state === 'running')

const wsUrl = computed(() => {
  if (!hostStore.currentHostId || !container.value || !isRunning.value) {
    return ''
  }
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host
  const containerId = route.params.id as string
  return `${protocol}//${host}/api/v1/ws/containers/${containerId}/logs?host_id=${hostStore.currentHostId}`
})

async function loadContainer() {
  if (!hostStore.currentHostId) return
  try {
    const res = await containerApi.get(hostStore.currentHostId, route.params.id as string)
    container.value = res.data.data
  } catch {
    // Handle error silently
  }
}

async function handleStart() {
  if (!hostStore.currentHostId || !container.value) return
  try {
    await containerApi.start(hostStore.currentHostId, container.value.id)
    await loadContainer()
  } catch {
    // Handle error silently
  }
}

// Watch for host changes
watch(() => hostStore.currentHostId, (newId) => {
  if (newId) {
    loadContainer()
  }
})

onMounted(() => {
  if (hostStore.currentHostId) {
    loadContainer()
  }
})
</script>

<style scoped>
.log-disabled {
  min-height: 300px;
  background: var(--fallback-b2, oklch(var(--b2) / 1));
  border-radius: 0.5rem;
}
</style>
