<script setup lang="ts">
import { computed } from 'vue'
import { Icon } from '@iconify/vue'
import type { Host } from '@/api'

const props = defineProps<{
  visible: boolean
  host: Host | null
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
}>()

const getTypeLabel = (type: string): string => {
  const labels: Record<string, string> = {
    local: '本地',
    tcp: 'TCP',
    ssh: 'SSH'
  }
  return labels[type] || type
}

const getTypeIcon = (type: string): string => {
  const icons: Record<string, string> = {
    local: 'mdi:desktop-classic',
    tcp: 'mdi:server-network',
    ssh: 'mdi:server'
  }
  return icons[type] || 'mdi:server'
}

const formatDate = (dateStr: string): string => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

const handleClose = () => {
  emit('update:visible', false)
}

const infoItems = computed(() => {
  if (!props.host) return []

  const items = [
    { label: '名称', value: props.host.name },
    { label: '类型', value: getTypeLabel(props.host.type), icon: getTypeIcon(props.host.type) },
    { label: '状态', value: props.host.is_active ? '在线' : '离线', isStatus: true, isActive: props.host.is_active },
    { label: '默认主机', value: props.host.is_default ? '是' : '否' },
  ]

  if (props.host.type !== 'local') {
    items.push({ label: '主机地址', value: props.host.host || '-' })
    items.push({ label: 'Docker 端口', value: props.host.docker_port?.toString() || '-' })
  }

  if (props.host.type === 'ssh') {
    items.push({ label: 'SSH 用户', value: props.host.ssh_user || '-' })
    items.push({ label: 'SSH 端口', value: props.host.ssh_port?.toString() || '-' })
    items.push({ label: '认证方式', value: props.host.ssh_auth_type === 'password' ? '密码' : '私钥' })
  }

  items.push({ label: '描述', value: props.host.description || '-' })
  items.push({ label: '创建时间', value: formatDate(props.host.created_at) })
  items.push({ label: '更新时间', value: formatDate(props.host.updated_at) })

  return items
})
</script>

<template>
  <dialog :class="['modal', { 'modal-open': visible }]">
    <div class="modal-box w-full max-w-md">
      <div class="flex items-center justify-between mb-4">
        <h3 class="font-bold text-lg">主机详情</h3>
        <button class="btn btn-sm btn-ghost btn-circle" @click="handleClose">
          <Icon icon="mdi:close" class="text-lg" />
        </button>
      </div>

      <div v-if="host" class="space-y-3">
        <div
          v-for="(item, index) in infoItems"
          :key="index"
          class="flex items-start gap-3 py-2 border-b border-base-200 last:border-0"
        >
          <span class="text-base-content/60 w-24 shrink-0 text-sm">{{ item.label }}</span>
          <div class="flex-1 flex items-center gap-2">
            <Icon v-if="item.icon" :icon="item.icon" class="text-lg text-base-content/60" />
            <span
              v-if="item.isStatus"
              class="flex items-center gap-1.5"
              :class="item.isActive ? 'text-success' : 'text-error'"
            >
              <span class="w-2 h-2 rounded-full" :class="item.isActive ? 'bg-success' : 'bg-error'"></span>
              {{ item.value }}
            </span>
            <span v-else class="text-sm">{{ item.value }}</span>
          </div>
        </div>
      </div>

      <div class="modal-action">
        <button class="btn btn-primary" @click="handleClose">关闭</button>
      </div>
    </div>
    <form method="dialog" class="modal-backdrop">
      <button @click="handleClose">close</button>
    </form>
  </dialog>
</template>
