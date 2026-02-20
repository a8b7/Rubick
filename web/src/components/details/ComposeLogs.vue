<template>
  <div class="compose-logs">
    <div v-if="!project" class="flex justify-center py-8">
      <span class="loading loading-spinner loading-lg"></span>
    </div>

    <template v-else>
      <!-- 项目未运行时显示提示 -->
      <div v-if="project.status !== 'running'" class="log-disabled">
        <div class="text-center py-16">
          <Icon icon="mdi:text-box-outline" class="text-6xl text-base-content/40 mb-4" />
          <p class="text-base-content/60 mb-4">项目未运行，无法查看实时日志</p>
          <button
            class="btn btn-primary btn-sm"
            @click="handleUp"
            :disabled="operating"
          >
            <span v-if="operating" class="loading loading-spinner loading-sm"></span>
            <Icon v-else icon="mdi:play" class="text-lg" />
            启动项目
          </button>
        </div>
      </div>

      <!-- 实时日志 -->
      <LogViewer
        v-else
        :ws-url="wsUrl"
        :auto-scroll="true"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { Icon } from '@iconify/vue'
import { composeApi, type ComposeProject } from '@/api'
import { showToast } from '@/utils/toast'
import LogViewer from '@/components/LogViewer.vue'

const props = defineProps<{
  hostId: string
  projectId: string
}>()

const project = ref<ComposeProject | null>(null)
const operating = ref(false)

const wsUrl = computed(() => {
  if (!project.value || project.value.status !== 'running') {
    return ''
  }
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host
  return `${protocol}//${host}/api/v1/ws/compose/${props.projectId}/logs`
})

async function loadProject() {
  try {
    const res = await composeApi.get(props.projectId)
    project.value = res.data.data
  } catch {
    project.value = null
  }
}

async function handleUp() {
  operating.value = true
  try {
    await composeApi.up(props.projectId, { detach: true })
    showToast('项目已启动', 'success')
    await loadProject()
  } finally {
    operating.value = false
  }
}

watch(() => props.projectId, loadProject)

onMounted(loadProject)
</script>

<style scoped>
.compose-logs {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 300px;
}

.log-disabled {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--fallback-b2, oklch(var(--b2) / 1));
  border-radius: 0.5rem;
}
</style>
