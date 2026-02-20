<template>
  <div class="image-detail">
    <div v-if="loading" class="flex justify-center py-8">
      <span class="loading loading-spinner loading-lg"></span>
    </div>

    <template v-else-if="image">
      <!-- 操作按钮 -->
      <div class="flex items-center gap-2 mb-4">
        <button
          class="btn btn-sm btn-error"
          @click="handleDelete"
          :disabled="operating"
        >
          <Icon icon="mdi:delete" class="text-lg" />
          删除镜像
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
                  <td class="font-medium w-32">ID</td>
                  <td class="font-mono text-sm">{{ image.id }}</td>
                </tr>
                <tr>
                  <td class="font-medium">标签</td>
                  <td>
                    <div class="flex flex-wrap gap-1">
                      <span
                        v-for="tag in image.repo_tags"
                        :key="tag"
                        class="badge badge-primary badge-sm"
                      >
                        {{ tag }}
                      </span>
                      <span v-if="!image.repo_tags?.length" class="text-base-content/50">
                        &lt;none&gt;
                      </span>
                    </div>
                  </td>
                </tr>
                <tr>
                  <td class="font-medium">大小</td>
                  <td>{{ formatSize(image.size) }}</td>
                </tr>
                <tr>
                  <td class="font-medium">创建时间</td>
                  <td>{{ formatDate(image.created) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- 标签 -->
      <div v-if="Object.keys(image.labels || {}).length" class="card bg-base-100 shadow mt-4">
        <div class="card-body">
          <h3 class="card-title text-base mb-4">标签</h3>
          <div class="flex flex-wrap gap-2">
            <span
              v-for="(value, key) in image.labels"
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
      <Icon icon="mdi:image-off-outline" class="text-4xl mb-2" />
      <p>无法加载镜像信息</p>
    </div>

    <Confirm ref="confirmRef" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { Icon } from '@iconify/vue'
import { imageApi, type Image } from '@/api'
import { showToast } from '@/utils/toast'
import { useResourceStore } from '@/stores'
import Confirm from '@/components/Confirm.vue'

const props = defineProps<{
  hostId: string
  imageId: string
}>()

const resourceStore = useResourceStore()

const image = ref<Image | null>(null)
const loading = ref(false)
const operating = ref(false)
const confirmRef = ref<InstanceType<typeof Confirm> | null>(null)

async function loadImage() {
  loading.value = true
  try {
    const res = await imageApi.get(props.hostId, props.imageId)
    image.value = res.data.data
  } catch {
    image.value = null
  } finally {
    loading.value = false
  }
}

async function handleDelete() {
  const result = await confirmRef.value?.confirm({
    title: '删除确认',
    message: `确定删除镜像 "${image.value?.repo_tags?.join(', ') || image.value?.id}"?`,
    type: 'danger'
  })
  if (result && result.confirmed) {
    operating.value = true
    try {
      await imageApi.remove(props.hostId, props.imageId, true)
      showToast('镜像已删除', 'success')
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

function formatDate(timestamp: number): string {
  return new Date(timestamp * 1000).toLocaleString()
}

watch(() => props.imageId, loadImage)
onMounted(loadImage)
</script>
