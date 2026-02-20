<template>
  <div class="page-container">
    <div class="card bg-base-100 shadow-xl">
      <div class="card-body">
        <div class="flex justify-between items-center mb-4">
          <h2 class="card-title">容器列表</h2>
          <button class="btn btn-primary" @click="refresh">刷新</button>
        </div>

        <div v-if="loading" class="flex justify-center py-8">
          <span class="loading loading-spinner loading-lg"></span>
        </div>

        <template v-else-if="groupedContainers.length > 0">
          <div v-for="group in groupedContainers" :key="group.name" class="mb-4">
            <div
              class="collapse collapse-arrow bg-base-200"
              :class="{ 'collapse-open': activeGroups.includes(group.name) }"
            >
              <input type="checkbox" :checked="activeGroups.includes(group.name)" @change="toggleGroup(group.name)" />
              <div class="collapse-title font-medium flex items-center gap-2">
                <span>{{ group.name }}</span>
                <span class="badge badge-sm">{{ group.containers.length }}</span>
              </div>
              <div class="collapse-content">
                <table class="table table-zebra">
                  <thead>
                    <tr>
                      <th>名称</th>
                      <th>镜像</th>
                      <th>状态</th>
                      <th>端口</th>
                      <th>操作</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="container in group.containers" :key="container.id">
                      <td>
                        <a class="link link-primary" @click="viewDetail(container)">{{ container.name }}</a>
                      </td>
                      <td>{{ container.image }}</td>
                      <td>
                        <span class="badge" :class="getStateBadgeClass(container.state)">
                          {{ container.state }}
                        </span>
                      </td>
                      <td>
                        <div v-for="(p, i) in container.ports" :key="i" class="text-sm">
                          {{ p.ip ? p.ip + ':' : '' }}{{ p.public_port }}->{{ p.private_port }}/{{ p.type }}
                        </div>
                      </td>
                      <td>
                        <div class="join">
                          <button
                            class="btn btn-sm join-item"
                            :disabled="!canStart(container.state)"
                            @click="startContainer(container)"
                          >启动</button>
                          <button
                            class="btn btn-sm join-item btn-warning"
                            :disabled="!canStop(container.state)"
                            @click="stopContainer(container)"
                          >停止</button>
                          <button
                            class="btn btn-sm join-item btn-info"
                            :disabled="!canRestart(container.state)"
                            @click="restartContainer(container)"
                          >重启</button>
                          <button
                            class="btn btn-sm join-item btn-error"
                            @click="removeContainer(container)"
                          >删除</button>
                        </div>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </template>

        <div v-else class="text-center py-8 text-base-content/60">
          <Icon icon="mdi:cube-outline" class="text-6xl mb-2" />
          <p>暂无容器</p>
        </div>
      </div>
    </div>

    <Confirm ref="confirmRef" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { Icon } from '@iconify/vue'
import { useHostStore } from '@/stores'
import { containerApi, type Container } from '@/api'
import { showToast } from '@/utils/toast'
import Confirm from '@/components/Confirm.vue'

interface ContainerGroup {
  name: string
  containers: Container[]
}

const hostStore = useHostStore()
const router = useRouter()
const containers = ref<Container[]>([])
const loading = ref(false)
const activeGroups = ref<string[]>([])
const confirmRef = ref<InstanceType<typeof Confirm> | null>(null)

// 按Compose项目分组容器
const groupedContainers = computed<ContainerGroup[]>(() => {
  const groups = new Map<string, Container[]>()
  const standalone: Container[] = []

  for (const c of containers.value) {
    const project = c.labels?.['com.docker.compose.project']
    if (project) {
      if (!groups.has(project)) {
        groups.set(project, [])
      }
      groups.get(project)!.push(c)
    } else {
      standalone.push(c)
    }
  }

  const result: ContainerGroup[] = Array.from(groups.entries())
    .sort((a, b) => a[0].localeCompare(b[0]))
    .map(([name, containers]) => ({ name, containers }))

  if (standalone.length > 0) {
    result.push({ name: '独立容器', containers: standalone })
  }

  return result
})

// 获取状态 badge 类名
function getStateBadgeClass(state: string): string {
  switch (state) {
    case 'running':
      return 'badge-success'
    case 'exited':
    case 'dead':
      return 'badge-error'
    case 'paused':
      return 'badge-warning'
    default:
      return 'badge-info'
  }
}

// 判断是否可以启动
function canStart(state: string): boolean {
  return state !== 'running' && state !== 'restarting'
}

// 判断是否可以停止
function canStop(state: string): boolean {
  return state === 'running'
}

// 判断是否可以重启
function canRestart(state: string): boolean {
  return state === 'running'
}

function toggleGroup(name: string) {
  const index = activeGroups.value.indexOf(name)
  if (index > -1) {
    activeGroups.value.splice(index, 1)
  } else {
    activeGroups.value.push(name)
  }
}

async function loadContainers() {
  if (!hostStore.currentHostId) return
  loading.value = true
  try {
    const res = await containerApi.list(hostStore.currentHostId)
    containers.value = res.data.data
    // 默认展开所有分组
    activeGroups.value = groupedContainers.value.map(g => g.name)
  } finally {
    loading.value = false
  }
}

function refresh() {
  loadContainers()
}

function viewDetail(container: Container) {
  router.push(`/containers/${container.id}`)
}

async function startContainer(container: Container) {
  try {
    await containerApi.start(hostStore.currentHostId, container.id)
    showToast('容器已启动', 'success')
    loadContainers()
  } catch {}
}

async function stopContainer(container: Container) {
  try {
    await containerApi.stop(hostStore.currentHostId, container.id)
    showToast('容器已停止', 'success')
    loadContainers()
  } catch {}
}

async function restartContainer(container: Container) {
  try {
    await containerApi.restart(hostStore.currentHostId, container.id)
    showToast('容器已重启', 'success')
    loadContainers()
  } catch {}
}

async function removeContainer(container: Container) {
  const result = await confirmRef.value?.confirm({
    title: '删除确认',
    message: `确定删除容器 "${container.name}"?`,
    type: 'danger',
    options: [
      { key: 'removeImage', label: '同时删除关联镜像', default: false }
    ]
  })
  if (result && result.confirmed) {
    const removeImage = result.options?.removeImage || false
    await containerApi.remove(hostStore.currentHostId, container.id, true, false, removeImage)
    showToast(removeImage ? '容器和镜像已删除' : '容器已删除', 'success')
    loadContainers()
  }
}

watch(() => hostStore.currentHostId, loadContainers)
onMounted(loadContainers)
</script>
