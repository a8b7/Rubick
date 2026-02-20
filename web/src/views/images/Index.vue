<template>
  <div class="page-container">
    <div class="card bg-base-100 shadow-xl">
      <div class="card-body">
        <div class="flex justify-between items-center mb-4">
          <h2 class="card-title">镜像列表</h2>
          <div class="flex gap-2">
            <button class="btn" @click="showPullDialog = true">拉取镜像</button>
            <button class="btn btn-primary" @click="refresh">刷新</button>
          </div>
        </div>

        <div v-if="loading" class="flex justify-center py-8">
          <span class="loading loading-spinner loading-lg"></span>
        </div>
        <table v-else class="table table-zebra">
          <thead>
            <tr>
              <th>仓库标签</th>
              <th>大小</th>
              <th>创建时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="image in images" :key="image.id">
              <td>{{ image.repo_tags?.join(', ') || '<none>' }}</td>
              <td>{{ formatSize(image.size) }}</td>
              <td>{{ new Date(image.created * 1000).toLocaleString() }}</td>
              <td>
                <button class="btn btn-sm btn-error" @click="removeImage(image)">删除</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 拉取镜像对话框 -->
    <dialog :class="['modal', { 'modal-open': showPullDialog }]">
      <div class="modal-box">
        <h3 class="font-bold text-lg mb-4">拉取镜像</h3>

        <div class="form-control mb-4">
          <label class="label">
            <span class="label-text">镜像名称</span>
          </label>
          <input
            v-model="pullImageName"
            type="text"
            placeholder="nginx:latest"
            class="input input-bordered w-full"
          />
        </div>

        <div class="modal-action">
          <button class="btn btn-ghost" @click="showPullDialog = false">取消</button>
          <button class="btn btn-primary" @click="pullImage" :disabled="pulling">
            <span v-if="pulling" class="loading loading-spinner loading-sm"></span>
            拉取
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button @click="showPullDialog = false">close</button>
      </form>
    </dialog>

    <Confirm ref="confirmRef" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useHostStore } from '@/stores'
import { imageApi, type Image } from '@/api'
import { showToast } from '@/utils/toast'
import Confirm from '@/components/Confirm.vue'

const hostStore = useHostStore()
const images = ref<Image[]>([])
const loading = ref(false)
const showPullDialog = ref(false)
const pullImageName = ref('')
const pulling = ref(false)
const confirmRef = ref<InstanceType<typeof Confirm> | null>(null)

function formatSize(size: number): string {
  if (size < 1024 * 1024) return (size / 1024).toFixed(2) + ' KB'
  if (size < 1024 * 1024 * 1024) return (size / 1024 / 1024).toFixed(2) + ' MB'
  return (size / 1024 / 1024 / 1024).toFixed(2) + ' GB'
}

async function loadImages() {
  if (!hostStore.currentHostId) return
  loading.value = true
  try {
    const res = await imageApi.list(hostStore.currentHostId)
    images.value = res.data.data
  } finally {
    loading.value = false
  }
}

function refresh() {
  loadImages()
}

async function pullImage() {
  if (!pullImageName.value) {
    showToast('请输入镜像名称', 'warning')
    return
  }
  pulling.value = true
  try {
    await imageApi.pull(hostStore.currentHostId, pullImageName.value)
    showToast('镜像拉取成功', 'success')
    showPullDialog.value = false
    pullImageName.value = ''
    loadImages()
  } finally {
    pulling.value = false
  }
}

async function removeImage(image: Image) {
  const result = await confirmRef.value?.confirm({
    title: '删除确认',
    message: `确定删除镜像 "${image.repo_tags?.join(', ') || image.id}"?`,
    type: 'danger'
  })
  if (result && result.confirmed) {
    await imageApi.remove(hostStore.currentHostId, image.id, true)
    showToast('镜像已删除', 'success')
    loadImages()
  }
}

watch(() => hostStore.currentHostId, loadImages)
onMounted(loadImages)
</script>
