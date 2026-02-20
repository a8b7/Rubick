<template>
  <div class="page-container">
    <div class="card bg-base-100 shadow-xl">
      <div class="card-body">
        <div class="flex justify-between items-center mb-4">
          <div class="flex items-center gap-2">
            <button class="btn btn-ghost btn-sm" @click="$router.back()">
              <Icon icon="mdi:arrow-left" class="text-xl" />
            </button>
            <button class="btn btn-primary btn-sm" @click="$router.push(`/compose/${project?.id}/edit`)">
              编辑
            </button>
          </div>
          <div class="flex gap-2">
            <button
              class="btn btn-success btn-sm"
              :disabled="upping || project?.status === 'running'"
              @click="composeUp"
            >
              <span v-if="upping" class="loading loading-spinner loading-sm"></span>
              启动
            </button>
            <button
              class="btn btn-warning btn-sm"
              :disabled="downing || project?.status !== 'running'"
              @click="composeDown"
            >
              <span v-if="downing" class="loading loading-spinner loading-sm"></span>
              停止
            </button>
            <button class="btn btn-sm" @click="loadServices">刷新状态</button>
          </div>
        </div>

        <div v-if="project" class="overflow-x-auto">
          <table class="table">
            <tbody>
              <tr>
                <td class="font-medium w-32">项目名称</td>
                <td>{{ project.name }}</td>
              </tr>
              <tr>
                <td class="font-medium">主机</td>
                <td>{{ project.host?.name }}</td>
              </tr>
              <tr>
                <td class="font-medium">源类型</td>
                <td>
                  <span class="badge" :class="project.source_type === 'directory' ? 'badge-warning' : 'badge-primary'">
                    {{ project.source_type === 'directory' ? '目录模式' : '内容模式' }}
                  </span>
                </td>
              </tr>
              <tr>
                <td class="font-medium">状态</td>
                <td>
                  <span class="badge" :class="project.status === 'running' ? 'badge-success' : 'badge-info'">
                    {{ project.status }}
                  </span>
                </td>
              </tr>
              <tr>
                <td class="font-medium">创建时间</td>
                <td>{{ new Date(project.created_at).toLocaleString() }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <div class="card bg-base-100 shadow-xl mt-4">
      <div class="card-body">
        <div class="flex justify-between items-center mb-4">
          <h3 class="card-title">实时日志</h3>
          <span v-if="!isRunning" class="badge badge-info">项目未运行</span>
        </div>
        <LogViewer
          v-if="isRunning && wsUrl"
          :ws-url="wsUrl"
          :auto-scroll="true"
        />
        <div v-else class="log-disabled">
          <div class="text-center py-16">
            <Icon icon="mdi:text-box-outline" class="text-6xl text-base-content/40 mb-4" />
            <p class="text-base-content/60 mb-4">项目未运行，无法查看实时日志</p>
            <button class="btn btn-primary" @click="composeUp" :disabled="upping">
              <span v-if="upping" class="loading loading-spinner loading-sm"></span>
              启动项目
            </button>
          </div>
        </div>
      </div>
    </div>

    <div class="card bg-base-100 shadow-xl mt-4">
      <div class="card-body">
        <h3 class="card-title mb-4">服务状态</h3>

        <div v-if="loadingServices" class="flex justify-center py-8">
          <span class="loading loading-spinner loading-lg"></span>
        </div>
        <table v-else class="table table-zebra">
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
                <span class="badge" :class="service.state === 'running' ? 'badge-success' : 'badge-error'">
                  {{ service.state }}
                </span>
              </td>
              <td>{{ service.status }}</td>
              <td>
                <span v-for="(p, i) in service.publishers" :key="i" class="mr-2">
                  {{ p.published_port }}:{{ p.target_port }}/{{ p.protocol }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="card bg-base-100 shadow-xl mt-4">
      <div class="card-body">
        <h3 class="card-title mb-4">Compose 配置</h3>

        <!-- 目录模式 -->
        <template v-if="project?.source_type === 'directory'">
          <table class="table">
            <tbody>
              <tr>
                <td class="font-medium w-32">工作目录</td>
                <td><code class="bg-base-200 px-2 py-1 rounded">{{ project.work_dir }}</code></td>
              </tr>
              <tr>
                <td class="font-medium">Compose 文件</td>
                <td><code class="bg-base-200 px-2 py-1 rounded">{{ project.compose_file || 'docker-compose.yml' }}</code></td>
              </tr>
              <tr v-if="project.env_file">
                <td class="font-medium">环境变量文件</td>
                <td><code class="bg-base-200 px-2 py-1 rounded">{{ project.env_file }}</code></td>
              </tr>
            </tbody>
          </table>
          <div class="alert alert-warning mt-4">
            <Icon icon="mdi:information" class="text-xl" />
            <span>此项目使用服务器上的目录，compose 文件位于 {{ project.work_dir }}/{{ project.compose_file || 'docker-compose.yml' }}</span>
          </div>
        </template>

        <!-- 内容模式 -->
        <template v-else>
          <div class="yaml-container">
            <pre>{{ project?.content }}</pre>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Icon } from '@iconify/vue'
import { composeApi, type ComposeProject, type ServiceStatus } from '@/api'
import { showToast } from '@/utils/toast'
import LogViewer from '@/components/LogViewer.vue'

const route = useRoute()
const router = useRouter()

const project = ref<ComposeProject | null>(null)
const services = ref<ServiceStatus[]>([])
const loadingServices = ref(false)
const upping = ref(false)
const downing = ref(false)

const isRunning = computed(() => project.value?.status === 'running')

const wsUrl = computed(() => {
  if (!project.value || !isRunning.value) {
    return ''
  }
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host
  const projectId = route.params.id as string
  return `${protocol}//${host}/api/v1/ws/compose/${projectId}/logs`
})

async function loadProject() {
  try {
    const res = await composeApi.get(route.params.id as string)
    project.value = res.data.data
  } catch {
    router.back()
  }
}

async function loadServices() {
  loadingServices.value = true
  try {
    const res = await composeApi.ps(route.params.id as string)
    services.value = res.data.data
  } finally {
    loadingServices.value = false
  }
}

async function composeUp() {
  upping.value = true
  try {
    await composeApi.up(route.params.id as string, { detach: true })
    showToast('项目已启动', 'success')
    await loadProject()
    await loadServices()
  } finally {
    upping.value = false
  }
}

async function composeDown() {
  downing.value = true
  try {
    await composeApi.down(route.params.id as string)
    showToast('项目已停止', 'success')
    await loadProject()
    await loadServices()
  } finally {
    downing.value = false
  }
}

onMounted(() => {
  loadProject()
  loadServices()
})
</script>

<style scoped>
.yaml-container {
  background: #1e1e1e;
  color: #d4d4d4;
  padding: 16px;
  border-radius: 0.5rem;
  max-height: 400px;
  overflow: auto;
}
.yaml-container pre {
  margin: 0;
  white-space: pre-wrap;
}
.log-disabled {
  min-height: 300px;
  background: var(--fallback-b2, oklch(var(--b2) / 1));
  border-radius: 0.5rem;
}
</style>
