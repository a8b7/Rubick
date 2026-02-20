<template>
  <div class="volume-detail">
    <div v-if="loading" class="flex justify-center py-8">
      <span class="loading loading-spinner loading-lg"></span>
    </div>

    <template v-else-if="volume">
      <!-- 操作按钮 -->
      <div class="flex items-center gap-2 mb-4">
        <button
          class="btn btn-sm btn-error"
          @click="handleDelete"
          :disabled="operating"
        >
          <Icon icon="mdi:delete" class="text-lg" />
          删除卷
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
                  <td class="font-medium w-32">名称</td>
                  <td>{{ volume.name }}</td>
                </tr>
                <tr>
                  <td class="font-medium">驱动</td>
                  <td>{{ volume.driver }}</td>
                </tr>
                <tr>
                  <td class="font-medium">挂载点</td>
                  <td class="font-mono text-sm">{{ volume.mountpoint }}</td>
                </tr>
                <tr>
                  <td class="font-medium">作用域</td>
                  <td>{{ volume.scope }}</td>
                </tr>
                <tr>
                  <td class="font-medium">创建时间</td>
                  <td>{{ volume.created_at }}</td>
                </tr>
                <tr v-if="volume.usage_data">
                  <td class="font-medium">使用情况</td>
                  <td>
                    <span class="badge badge-info badge-sm mr-2">
                      {{ volume.usage_data.ref_count }} 个容器使用
                    </span>
                    <span class="badge badge-outline badge-sm">
                      {{ formatSize(volume.usage_data.size) }}
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- 标签 -->
      <div v-if="Object.keys(volume.labels || {}).length" class="card bg-base-100 shadow mt-4">
        <div class="card-body">
          <h3 class="card-title text-base mb-4">标签</h3>
          <div class="flex flex-wrap gap-2">
            <span
              v-for="(value, key) in volume.labels"
              :key="key"
              class="badge badge-outline"
            >
              {{ key }}={{ value }}
            </span>
          </div>
        </div>
      </div>
    </template>

    <div v-else class="text-center py-8 text-base-content/50">
      <Icon icon="mdi:folder-off-outline" class="text-4xl mb-2" />
      <p>无法加载卷信息</p>
    </div>

    <Confirm ref="confirmRef" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { Icon } from '@iconify/vue'
import { volumeApi, type Volume } from '@/api'
import { showToast } from '@/utils/toast'
import { useResourceStore } from '@/stores'
import Confirm from '@/components/Confirm.vue'

const props = defineProps<{
  hostId: string
  volumeName: string
}>()

const resourceStore = useResourceStore()

const volume = ref<Volume | null>(null)
const loading = ref(false)
const operating = ref(false)
const confirmRef = ref<InstanceType<typeof Confirm> | null>(null)

async function loadVolume() {
  loading.value = true
  try {
    const res = await volumeApi.get(props.hostId, props.volumeName)
    volume.value = res.data.data
  } catch {
    volume.value = null
  } finally {
    loading.value = false
  }
}

async function handleDelete() {
  const result = await confirmRef.value?.confirm({
    title: '删除确认',
    message: `确定删除卷 "${volume.value?.name}"?`,
    type: 'danger'
  })
  if (result && result.confirmed) {
    operating.value = true
    try {
      await volumeApi.remove(props.hostId, props.volumeName, true)
      showToast('卷已删除', 'success')
      resourceStore.refreshHostResources(props.hostId)
    } finally {
      operating.value = false
    }
  }
}

function formatSize(size: number): string {
  if (size < 1024 * 1024) return (size / 1024).toFixed(2) + ' KB'
  if (size < 1024 * 1024 * 1024) return (size / 1024 / 1024).toFixed(2) + ' MB'
  return (size / 1024 / 1024 / 1024).toFixed(2) + ' GB'
}

watch(() => props.volumeName, loadVolume)
onMounted(loadVolume)
</script>
