<script setup lang="ts">
import { computed, ref } from 'vue'
import { Icon } from '@iconify/vue'
import { useResourceStore, useHostStore } from '@/stores'
import { hostApi, type Host } from '@/api'
import { showToast } from '@/utils/toast'
import ResourceGroup from './ResourceGroup.vue'
import ContextMenu, { type MenuItem } from '@/components/ContextMenu.vue'
import HostDetailDialog from '@/components/HostDetailDialog.vue'
import HostFormDialog from '@/components/HostFormDialog.vue'
import Confirm from '@/components/Confirm.vue'

const props = defineProps<{
  host: Host
}>()

defineEmits<{
  (e: 'select', selection: { hostId: string; type: string; id: string; name: string }): void
  (e: 'create-compose', hostId: string): void
}>()

const resourceStore = useResourceStore()
const hostStore = useHostStore()

// 右键菜单状态
const contextMenuVisible = ref(false)
const contextMenuPosition = ref({ x: 0, y: 0 })

// 对话框状态
const detailDialogVisible = ref(false)
const editDialogVisible = ref(false)
const confirmRef = ref<InstanceType<typeof Confirm> | null>(null)

const isExpanded = computed(() => resourceStore.isHostExpanded(props.host.id))
const cache = computed(() => resourceStore.getHostCache(props.host.id))

function getHostIcon(): string {
  switch (props.host.type) {
    case 'local':
      return 'mdi:desktop-classic'
    case 'tcp':
      return 'mdi:server-network'
    case 'ssh':
      return 'mdi:server'
    default:
      return 'mdi:server'
  }
}

function toggleExpand() {
  resourceStore.toggleHostExpand(props.host.id)
}

function isGroupExpanded(type: string): boolean {
  return resourceStore.isGroupExpanded(`${props.host.id}:${type}`)
}

function toggleGroup(type: string) {
  resourceStore.toggleGroupExpand(`${props.host.id}:${type}`)
}

// 右键菜单处理
function handleContextMenu(e: MouseEvent) {
  e.preventDefault()
  contextMenuPosition.value = { x: e.clientX, y: e.clientY }
  contextMenuVisible.value = true
}

const menuItems = computed<MenuItem[]>(() => [
  {
    label: '查看详情',
    icon: 'mdi:eye',
    action: () => {
      detailDialogVisible.value = true
    }
  },
  {
    label: '编辑',
    icon: 'mdi:pencil',
    action: () => {
      editDialogVisible.value = true
    }
  },
  {
    label: '测试连接',
    icon: 'mdi:lan-connect',
    action: handleTestConnection
  },
  {
    label: '设为默认',
    icon: 'mdi:star',
    disabled: props.host.is_default,
    action: handleSetDefault
  },
  { divider: true },
  {
    label: '删除',
    icon: 'mdi:delete',
    danger: true,
    action: handleDelete
  }
])

function closeContextMenu() {
  contextMenuVisible.value = false
}

// 测试连接
async function handleTestConnection() {
  try {
    const res = await hostApi.test(props.host.id)
    if (res.data.data.success) {
      showToast('连接成功', 'success')
    } else {
      showToast(res.data.data.message || '连接失败', 'error')
    }
  } catch {
    showToast('连接测试失败', 'error')
  }
}

// 设为默认
async function handleSetDefault() {
  try {
    await hostApi.update(props.host.id, { is_default: true } as never)
    showToast('已设为默认主机', 'success')
    hostStore.loadHosts()
  } catch {
    showToast('操作失败', 'error')
  }
}

// 删除主机
async function handleDelete() {
  if (!confirmRef.value) return

  const result = await confirmRef.value.confirm({
    title: '删除主机',
    message: `确定要删除主机「${props.host.name}」吗？此操作不可恢复。`,
    confirmText: '删除',
    cancelText: '取消',
    type: 'danger'
  })

  if (result.confirmed) {
    try {
      await hostApi.delete(props.host.id)
      showToast('删除成功', 'success')
      hostStore.loadHosts()
    } catch {
      showToast('删除失败', 'error')
    }
  }
}
</script>

<template>
  <div class="host-tree-node">
    <!-- 主机节点 -->
    <div
      :class="[
        'flex items-center gap-2 px-2 py-1.5 rounded-lg cursor-pointer transition-colors',
        isExpanded ? 'bg-base-200' : 'hover:bg-base-200/50'
      ]"
      @click="toggleExpand"
      @contextmenu="handleContextMenu"
    >
      <Icon
        :icon="isExpanded ? 'mdi:chevron-down' : 'mdi:chevron-right'"
        class="text-lg text-base-content/60"
      />
      <Icon
        :icon="getHostIcon()"
        :class="['text-lg', host.is_active ? 'text-success' : 'text-error']"
      />
      <span class="flex-1 truncate text-sm">{{ host.name }}</span>
      <span v-if="cache?.loading" class="loading loading-spinner loading-xs"></span>
      <span
        v-else-if="host.is_active"
        class="w-2 h-2 rounded-full bg-success"
        title="在线"
      ></span>
      <span
        v-else
        class="w-2 h-2 rounded-full bg-error"
        title="离线"
      ></span>
    </div>

    <!-- 资源分组 -->
    <div v-if="isExpanded && cache" class="ml-4 mt-1 space-y-0.5">
      <!-- 容器 -->
      <ResourceGroup
        :host-id="host.id"
        type="container"
        :items="cache.containers"
        :expanded="isGroupExpanded('container')"
        @toggle="toggleGroup('container')"
        @select="$emit('select', $event)"
      />

      <!-- 镜像 -->
      <ResourceGroup
        :host-id="host.id"
        type="image"
        :items="cache.images"
        :expanded="isGroupExpanded('image')"
        @toggle="toggleGroup('image')"
        @select="$emit('select', $event)"
      />

      <!-- 卷 -->
      <ResourceGroup
        :host-id="host.id"
        type="volume"
        :items="cache.volumes"
        :expanded="isGroupExpanded('volume')"
        @toggle="toggleGroup('volume')"
        @select="$emit('select', $event)"
      />

      <!-- 网络 -->
      <ResourceGroup
        :host-id="host.id"
        type="network"
        :items="cache.networks"
        :expanded="isGroupExpanded('network')"
        @toggle="toggleGroup('network')"
        @select="$emit('select', $event)"
      />

      <!-- Compose 项目 -->
      <ResourceGroup
        :host-id="host.id"
        type="compose"
        :items="cache.composeProjects"
        :expanded="isGroupExpanded('compose')"
        @toggle="toggleGroup('compose')"
        @select="$emit('select', $event)"
        @create="$emit('create-compose', $event)"
      />
    </div>

    <!-- 右键菜单 -->
    <ContextMenu
      :visible="contextMenuVisible"
      :x="contextMenuPosition.x"
      :y="contextMenuPosition.y"
      :items="menuItems"
      @close="closeContextMenu"
    />

    <!-- 查看详情对话框 -->
    <HostDetailDialog
      v-model:visible="detailDialogVisible"
      :host="host"
    />

    <!-- 编辑对话框 -->
    <HostFormDialog
      v-model:visible="editDialogVisible"
      :host="host"
      @saved="hostStore.loadHosts()"
    />

    <!-- 确认对话框 -->
    <Confirm ref="confirmRef" />
  </div>
</template>
