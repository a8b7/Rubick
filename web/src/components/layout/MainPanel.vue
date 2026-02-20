<template>
  <div class="main-panel flex-1 flex flex-col h-full bg-base-200 overflow-hidden">
    <!-- 空状态 -->
    <div
      v-if="!resourceStore.selection"
      class="flex-1 flex flex-col items-center justify-center text-base-content/40"
    >
      <Icon icon="mdi:cursor-default-click" class="text-8xl mb-4" />
      <p class="text-lg">从左侧选择一个资源</p>
      <p class="text-sm mt-2">展开主机查看容器、镜像、卷和网络</p>
    </div>

    <!-- 详情内容 -->
    <template v-else>
      <!-- 面包屑导航 -->
      <div class="h-[50px] flex items-center px-4 border-b border-base-content/10 bg-base-100">
        <div class="flex items-center gap-2 text-sm">
          <Icon :icon="getTypeIcon()" class="text-lg" :class="getTypeIconColor()" />
          <span class="font-medium">{{ getHostName() }}</span>
          <Icon icon="mdi:chevron-right" class="text-base-content/40" />
          <span class="text-base-content/60">{{ getTypeLabel() }}</span>
          <Icon icon="mdi:chevron-right" class="text-base-content/40" />
          <span class="text-primary">{{ resourceStore.selection.name }}</span>
        </div>

        <!-- 刷新按钮 -->
        <button class="btn btn-sm btn-ghost ml-auto" @click="handleRefresh">
          <Icon icon="mdi:refresh" :class="{ 'animate-spin': refreshing }" />
        </button>
      </div>

      <!-- Tab 切换器 -->
      <TabSwitcher
        v-model="activeTab"
        :type="resourceStore.selection.type"
      />

      <!-- 详情面板 -->
      <DetailPanel
        :tab="activeTab"
        :selection="resourceStore.selection"
        :key="panelKey"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Icon } from '@iconify/vue'
import { useResourceStore, useHostStore } from '@/stores'
import TabSwitcher from './TabSwitcher.vue'
import DetailPanel from './DetailPanel.vue'

const resourceStore = useResourceStore()
const hostStore = useHostStore()

const activeTab = ref<'detail' | 'logs' | 'terminal'>('detail')
const refreshing = ref(false)

// 用于强制刷新详情面板
const panelKey = computed(() => {
  const sel = resourceStore.selection
  return sel ? `${sel.hostId}:${sel.type}:${sel.id}:${activeTab.value}` : ''
})

// 当选中项变化时重置 Tab
watch(() => resourceStore.selection, (newSel, oldSel) => {
  if (newSel?.id !== oldSel?.id || newSel?.type !== oldSel?.type) {
    activeTab.value = 'detail'
  }
})

function getHostName(): string {
  const hostId = resourceStore.selection?.hostId
  if (!hostId) return ''
  const host = hostStore.hosts.find(h => h.id === hostId)
  return host?.name || 'Unknown'
}

function getTypeIcon(): string {
  switch (resourceStore.selection?.type) {
    case 'host':
      return 'mdi:server'
    case 'container':
      return 'mdi:cube-outline'
    case 'image':
      return 'mdi:image-outline'
    case 'volume':
      return 'mdi:folder-outline'
    case 'network':
      return 'mdi:lan'
    case 'compose':
      return 'mdi:file-document-outline'
    default:
      return 'mdi:file-outline'
  }
}

function getTypeIconColor(): string {
  switch (resourceStore.selection?.type) {
    case 'host':
      return 'text-primary'
    case 'container':
      return 'text-info'
    case 'image':
      return 'text-warning'
    case 'volume':
      return 'text-success'
    case 'network':
      return 'text-secondary'
    case 'compose':
      return 'text-secondary'
    default:
      return 'text-base-content/60'
  }
}

function getTypeLabel(): string {
  switch (resourceStore.selection?.type) {
    case 'host':
      return '主机'
    case 'container':
      return '容器'
    case 'image':
      return '镜像'
    case 'volume':
      return '卷'
    case 'network':
      return '网络'
    case 'compose':
      return 'Compose'
    default:
      return '资源'
  }
}

async function handleRefresh() {
  const hostId = resourceStore.selection?.hostId
  if (!hostId) return

  refreshing.value = true
  try {
    await resourceStore.refreshHostResources(hostId)
  } finally {
    refreshing.value = false
  }
}
</script>
