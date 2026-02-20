<template>
  <div class="host-detail">
    <div v-if="loading" class="flex justify-center py-8">
      <span class="loading loading-spinner loading-lg"></span>
    </div>

    <template v-else-if="host">
      <!-- 操作按钮 -->
      <div class="flex items-center gap-2 mb-4">
        <button class="btn btn-sm btn-primary" @click="handleTest" :disabled="testing">
          <span v-if="testing" class="loading loading-spinner loading-sm"></span>
          <Icon v-else icon="mdi:connection" class="text-lg" />
          测试连接
        </button>
        <button class="btn btn-sm" @click="$emit('edit', host)">
          <Icon icon="mdi:pencil" class="text-lg" />
          编辑
        </button>
        <button
          class="btn btn-sm btn-error"
          @click="handleDelete"
          :disabled="deleting"
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
                  <td class="font-mono text-sm">{{ host.id }}</td>
                </tr>
                <tr>
                  <td class="font-medium">名称</td>
                  <td>{{ host.name }}</td>
                </tr>
                <tr>
                  <td class="font-medium">类型</td>
                  <td>
                    <span class="badge" :class="getHostTypeBadge()">
                      {{ getHostTypeLabel() }}
                    </span>
                  </td>
                </tr>
                <tr>
                  <td class="font-medium">地址</td>
                  <td>{{ host.host || '本地' }}</td>
                </tr>
                <tr>
                  <td class="font-medium">状态</td>
                  <td>
                    <span :class="['badge', host.is_active ? 'badge-success' : 'badge-error']">
                      {{ host.is_active ? '活跃' : '离线' }}
                    </span>
                  </td>
                </tr>
                <tr>
                  <td class="font-medium">默认主机</td>
                  <td>
                    <span :class="['badge badge-sm', host.is_default ? 'badge-primary' : 'badge-ghost']">
                      {{ host.is_default ? '是' : '否' }}
                    </span>
                  </td>
                </tr>
                <tr v-if="host.docker_port">
                  <td class="font-medium">Docker 端口</td>
                  <td>{{ host.docker_port }}</td>
                </tr>
                <tr v-if="host.ssh_user">
                  <td class="font-medium">SSH 用户</td>
                  <td>{{ host.ssh_user }}</td>
                </tr>
                <tr v-if="host.ssh_port">
                  <td class="font-medium">SSH 端口</td>
                  <td>{{ host.ssh_port }}</td>
                </tr>
                <tr>
                  <td class="font-medium">创建时间</td>
                  <td>{{ formatDate(host.created_at) }}</td>
                </tr>
                <tr v-if="host.description">
                  <td class="font-medium">描述</td>
                  <td>{{ host.description }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Docker 信息 -->
      <div v-if="dockerInfo" class="card bg-base-100 shadow mt-4">
        <div class="card-body">
          <h3 class="card-title text-base mb-4">Docker 信息</h3>
          <div class="overflow-x-auto">
            <table class="table table-sm">
              <tbody>
                <tr>
                  <td class="font-medium w-32">版本</td>
                  <td>{{ dockerInfo.Version }}</td>
                </tr>
                <tr>
                  <td class="font-medium">API 版本</td>
                  <td>{{ dockerInfo.ApiVersion }}</td>
                </tr>
                <tr>
                  <td class="font-medium">操作系统</td>
                  <td>{{ dockerInfo.Os }} {{ dockerInfo.Arch }}</td>
                </tr>
                <tr>
                  <td class="font-medium">内核版本</td>
                  <td>{{ dockerInfo.KernelVersion }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </template>

    <div v-else class="text-center py-8 text-base-content/50">
      <Icon icon="mdi:server-off" class="text-4xl mb-2" />
      <p>无法加载主机信息</p>
    </div>

    <Confirm ref="confirmRef" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { Icon } from '@iconify/vue'
import { hostApi, type Host } from '@/api'
import { showToast } from '@/utils/toast'
import { useHostStore } from '@/stores'
import Confirm from '@/components/Confirm.vue'

const props = defineProps<{
  hostId: string
}>()

const emit = defineEmits<{
  (e: 'edit', host: Host): void
}>()

const hostStore = useHostStore()

const host = ref<Host | null>(null)
const loading = ref(false)
const testing = ref(false)
const deleting = ref(false)
const confirmRef = ref<InstanceType<typeof Confirm> | null>(null)

// Docker 信息（如果后端有提供）
const dockerInfo = ref<{
  Version?: string
  ApiVersion?: string
  Os?: string
  Arch?: string
  KernelVersion?: string
} | null>(null)

async function loadHost() {
  loading.value = true
  try {
    const res = await hostApi.get(props.hostId)
    host.value = res.data.data
  } catch {
    host.value = null
  } finally {
    loading.value = false
  }
}

async function handleTest() {
  if (!host.value) return

  testing.value = true
  try {
    const res = await hostApi.test(host.value.id)
    if (res.data.data.success) {
      showToast('连接成功', 'success')
    } else {
      showToast(res.data.data.message || '连接失败', 'error')
    }
  } finally {
    testing.value = false
  }
}

async function handleDelete() {
  if (!host.value) return

  const result = await confirmRef.value?.confirm({
    title: '删除确认',
    message: `确定删除主机 "${host.value.name}"?`,
    type: 'danger'
  })
  if (result && result.confirmed) {
    deleting.value = true
    try {
      await hostApi.delete(host.value.id)
      showToast('主机已删除', 'success')
      hostStore.loadHosts()
    } finally {
      deleting.value = false
    }
  }
}

function getHostTypeLabel(): string {
  switch (host.value?.type) {
    case 'local':
      return '本地'
    case 'tcp':
      return 'TCP'
    case 'ssh':
      return 'SSH'
    default:
      return host.value?.type || 'Unknown'
  }
}

function getHostTypeBadge(): string {
  switch (host.value?.type) {
    case 'local':
      return 'badge-success'
    case 'tcp':
      return 'badge-info'
    case 'ssh':
      return 'badge-warning'
    default:
      return 'badge-ghost'
  }
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString()
}

watch(() => props.hostId, loadHost)
onMounted(loadHost)
</script>
