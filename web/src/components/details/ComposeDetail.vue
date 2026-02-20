<template>
  <div class="compose-detail">
    <div v-if="loading" class="flex justify-center py-8">
      <span class="loading loading-spinner loading-lg"></span>
    </div>

    <template v-else-if="project">
      <!-- 操作按钮 -->
      <div class="flex items-center gap-2 mb-4">
        <button
          class="btn btn-sm btn-success"
          :disabled="operating || project.status === 'running'"
          @click="handleUp"
        >
          <Icon icon="mdi:play" class="text-lg" />
          启动
        </button>
        <button
          class="btn btn-sm btn-warning"
          :disabled="operating || project.status !== 'running'"
          @click="handleDown"
        >
          <Icon icon="mdi:stop" class="text-lg" />
          停止
        </button>
        <button
          class="btn btn-sm btn-info"
          :disabled="operating || project.status !== 'running'"
          @click="handleRestart"
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
                  <td class="font-medium w-32">项目名称</td>
                  <td>{{ project.name }}</td>
                </tr>
                <tr>
                  <td class="font-medium">源类型</td>
                  <td>
                    <span
                      class="badge"
                      :class="project.source_type === 'directory' ? 'badge-warning' : 'badge-primary'"
                    >
                      {{ project.source_type === 'directory' ? '目录模式' : '内容模式' }}
                    </span>
                  </td>
                </tr>
                <tr>
                  <td class="font-medium">状态</td>
                  <td>
                    <span
                      class="badge"
                      :class="project.status === 'running' ? 'badge-success' : 'badge-info'"
                    >
                      {{ project.status }}
                    </span>
                  </td>
                </tr>
                <tr>
                  <td class="font-medium">创建时间</td>
                  <td>{{ formatDate(project.created_at) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- 服务状态 -->
      <div class="card bg-base-100 shadow mt-4">
        <div class="card-body">
          <div class="flex items-center justify-between mb-4">
            <h3 class="card-title text-base">服务状态</h3>
            <button class="btn btn-xs btn-ghost" @click="loadServices" :disabled="loadingServices">
              <Icon icon="mdi:refresh" :class="{ 'animate-spin': loadingServices }" />
            </button>
          </div>

          <div v-if="loadingServices" class="flex justify-center py-4">
            <span class="loading loading-spinner loading-md"></span>
          </div>
          <div v-else-if="services.length === 0" class="text-center py-4 text-base-content/50">
            <Icon icon="mdi:cube-off-outline" class="text-3xl mb-2" />
            <p class="text-sm">暂无服务</p>
          </div>
          <div v-else class="overflow-x-auto">
            <table class="table table-sm">
              <thead>
                <tr>
                  <th>服务名称</th>
                  <th>状态</th>
                  <th>详情</th>
                  <th>端口</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="service in services" :key="service.name">
                  <td>{{ service.name }}</td>
                  <td>
                    <span
                      class="badge badge-sm"
                      :class="service.state === 'running' ? 'badge-success' : 'badge-error'"
                    >
                      {{ service.state }}
                    </span>
                  </td>
                  <td class="text-sm text-base-content/70">{{ service.status }}</td>
                  <td class="text-sm">
                    <span v-for="(p, i) in service.publishers" :key="i" class="mr-2">
                      {{ p.published_port }}:{{ p.target_port }}/{{ p.protocol }}
                    </span>
                    <span v-if="!service.publishers?.length">-</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- Compose 配置 -->
      <div class="card bg-base-100 shadow mt-4">
        <div class="card-body">
          <h3 class="card-title text-base mb-4">Compose 配置</h3>

          <!-- 目录模式 -->
          <template v-if="project.source_type === 'directory'">
            <div class="overflow-x-auto">
              <table class="table table-sm">
                <tbody>
                  <tr>
                    <td class="font-medium w-32">工作目录</td>
                    <td><code class="bg-base-200 px-2 py-1 rounded text-sm">{{ project.work_dir }}</code></td>
                  </tr>
                  <tr>
                    <td class="font-medium">Compose 文件</td>
                    <td><code class="bg-base-200 px-2 py-1 rounded text-sm">{{ project.compose_file || 'docker-compose.yml' }}</code></td>
                  </tr>
                  <tr v-if="project.env_file">
                    <td class="font-medium">环境变量文件</td>
                    <td><code class="bg-base-200 px-2 py-1 rounded text-sm">{{ project.env_file }}</code></td>
                  </tr>
                </tbody>
              </table>
            </div>
            <div class="alert alert-warning mt-4 text-sm">
              <Icon icon="mdi:information" class="text-lg" />
              <span>此项目使用服务器上的目录，compose 文件位于 {{ project.work_dir }}/{{ project.compose_file || 'docker-compose.yml' }}</span>
            </div>
          </template>

          <!-- 内容模式 -->
          <template v-else>
            <div class="yaml-container">
              <pre>{{ project.content }}</pre>
            </div>
          </template>
        </div>
      </div>
    </template>

    <div v-else class="text-center py-8 text-base-content/50">
      <Icon icon="mdi:file-document-remove-outline" class="text-4xl mb-2" />
      <p>无法加载项目信息</p>
    </div>

    <Confirm ref="confirmRef" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { Icon } from '@iconify/vue'
import { composeApi, type ComposeProject, type ServiceStatus } from '@/api'
import { showToast } from '@/utils/toast'
import { useResourceStore } from '@/stores'
import Confirm from '@/components/Confirm.vue'

const props = defineProps<{
  hostId: string
  projectId: string
}>()

const emit = defineEmits<{
  (e: 'refresh'): void
}>()

const resourceStore = useResourceStore()

const project = ref<ComposeProject | null>(null)
const services = ref<ServiceStatus[]>([])
const loading = ref(false)
const loadingServices = ref(false)
const operating = ref(false)
const confirmRef = ref<InstanceType<typeof Confirm> | null>(null)

async function loadProject() {
  loading.value = true
  try {
    const res = await composeApi.get(props.projectId)
    project.value = res.data.data
  } catch {
    project.value = null
  } finally {
    loading.value = false
  }
}

async function loadServices() {
  loadingServices.value = true
  try {
    const res = await composeApi.ps(props.projectId)
    services.value = res.data.data
  } catch {
    services.value = []
  } finally {
    loadingServices.value = false
  }
}

async function handleUp() {
  operating.value = true
  try {
    await composeApi.up(props.projectId, { detach: true })
    showToast('项目已启动', 'success')
    await loadProject()
    await loadServices()
    resourceStore.refreshHostResources(props.hostId)
  } finally {
    operating.value = false
  }
}

async function handleDown() {
  operating.value = true
  try {
    await composeApi.down(props.projectId)
    showToast('项目已停止', 'success')
    await loadProject()
    await loadServices()
    resourceStore.refreshHostResources(props.hostId)
  } finally {
    operating.value = false
  }
}

async function handleRestart() {
  operating.value = true
  try {
    await composeApi.restart(props.projectId)
    showToast('项目已重启', 'success')
    await loadProject()
    await loadServices()
    resourceStore.refreshHostResources(props.hostId)
  } finally {
    operating.value = false
  }
}

async function handleDelete() {
  const result = await confirmRef.value?.confirm({
    title: '删除确认',
    message: `确定删除项目 "${project.value?.name}"?`,
    type: 'danger',
    options: [
      { key: 'downFirst', label: '先停止项目', default: true }
    ]
  })
  if (result && result.confirmed) {
    operating.value = true
    try {
      if (result.options?.downFirst && project.value?.status === 'running') {
        await composeApi.down(props.projectId)
      }
      await composeApi.delete(props.projectId)
      showToast('项目已删除', 'success')
      resourceStore.refreshHostResources(props.hostId)
      emit('refresh')
    } finally {
      operating.value = false
    }
  }
}

function formatDate(timestamp: string): string {
  return new Date(timestamp).toLocaleString()
}

watch(() => props.projectId, () => {
  loadProject()
  loadServices()
})

onMounted(() => {
  loadProject()
  loadServices()
})
</script>

<style scoped>
.yaml-container {
  background: #1e1e1e;
  color: #d4d4d4;
  padding: 12px;
  border-radius: 0.5rem;
  max-height: 300px;
  overflow: auto;
  font-size: 12px;
}
.yaml-container pre {
  margin: 0;
  white-space: pre-wrap;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', 'Consolas', monospace;
}
</style>
