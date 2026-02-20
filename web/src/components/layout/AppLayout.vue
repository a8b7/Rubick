<template>
  <div class="flex h-screen">
    <!-- 左侧栏 -->
    <SidebarLeft
      @add-host="showHostDialog = true"
      @select="handleSelect"
      @create-compose="handleCreateCompose"
    />

    <!-- 右侧主面板 -->
    <MainPanel />

    <!-- 主机表单对话框 -->
    <HostFormDialog
      v-model:visible="showHostDialog"
      :host="editingHost"
      @saved="handleHostSaved"
    />

    <!-- Compose 项目表单对话框 -->
    <ComposeFormDialog
      v-model:visible="showComposeDialog"
      :host-id="creatingComposeHostId"
      @saved="handleComposeSaved"
    />

    <!-- Toast 组件 -->
    <Toast />

    <!-- 确认对话框 -->
    <Confirm />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useHostStore, useResourceStore, type ResourceType } from '@/stores'
import Toast from '@/components/Toast.vue'
import Confirm from '@/components/Confirm.vue'
import HostFormDialog from '@/components/HostFormDialog.vue'
import ComposeFormDialog from '@/components/ComposeFormDialog.vue'
import SidebarLeft from '@/components/layout/SidebarLeft.vue'
import MainPanel from '@/components/layout/MainPanel.vue'
import type { Host } from '@/api'

const hostStore = useHostStore()
const resourceStore = useResourceStore()

const showHostDialog = ref(false)
const editingHost = ref<Host | null>(null)
const showComposeDialog = ref(false)
const creatingComposeHostId = ref<string>('')

function handleSelect(selection: { hostId: string; type: string; id: string; name: string }) {
  resourceStore.setSelection({
    ...selection,
    type: selection.type as ResourceType,
  })
}

function handleHostSaved() {
  editingHost.value = null
}

function handleCreateCompose(hostId: string) {
  creatingComposeHostId.value = hostId
  showComposeDialog.value = true
}

function handleComposeSaved() {
  creatingComposeHostId.value = ''
  // 刷新资源
  const hostId = resourceStore.selectedHostId
  if (hostId) {
    resourceStore.refreshHostResources(hostId)
  }
}

onMounted(() => {
  hostStore.init()
})
</script>

<style>
/* 全局过渡动画 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
