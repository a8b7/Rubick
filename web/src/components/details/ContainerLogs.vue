<template>
  <div class="container-logs flex flex-col h-full">
    <div v-if="loading" class="flex justify-center py-8">
      <span class="loading loading-spinner loading-lg"></span>
    </div>

    <template v-else>
      <!-- 容器未运行提示 -->
      <div v-if="!isRunning" class="flex-1 flex flex-col items-center justify-center text-base-content/40">
        <Icon icon="mdi:text-box-outline" class="text-6xl mb-4" />
        <p class="text-base-content/60 mb-4">容器未运行，无法查看实时日志</p>
      </div>

      <!-- 日志查看器 -->
      <LogViewer
        v-else-if="wsUrl"
        :ws-url="wsUrl"
        :auto-scroll="true"
        class="flex-1"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { Icon } from '@iconify/vue'
import { containerApi, type Container } from '@/api'
import LogViewer from '@/components/LogViewer.vue'

const props = defineProps<{
  hostId: string
  containerId: string
}>()

const container = ref<Container | null>(null)
const loading = ref(true)

const isRunning = computed(() => container.value?.state === 'running')

const wsUrl = computed(() => {
  if (!props.hostId || !props.containerId || !isRunning.value) {
    return ''
  }
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host
  return `${protocol}//${host}/api/v1/ws/containers/${props.containerId}/logs?host_id=${props.hostId}`
})

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

watch(() => props.containerId, loadContainer)
onMounted(loadContainer)
</script>

<style scoped>
.container-logs {
  height: 100%;
}
</style>
