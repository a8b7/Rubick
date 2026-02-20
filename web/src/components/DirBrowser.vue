<template>
  <dialog :class="['modal', { 'modal-open': visible }]">
    <div class="modal-box w-full max-w-2xl">
      <h3 class="font-bold text-lg mb-4">浏览目录</h3>

      <div class="browser-container">
        <!-- 面包屑导航 -->
        <div class="breadcrumb mb-4">
          <div class="flex items-center gap-1 flex-wrap">
            <template v-for="(part, index) in pathParts" :key="index">
              <a class="link link-primary text-sm" @click="navigateTo(index)">{{ part || '/' }}</a>
              <span v-if="index < pathParts.length - 1" class="text-base-content/40">/</span>
            </template>
          </div>
        </div>

        <!-- 路径输入 -->
        <div class="join w-full mb-4">
          <input
            v-model="currentPath"
            type="text"
            placeholder="输入路径"
            class="input input-bordered join-item flex-1"
            @keyup.enter="loadDirectory"
          />
          <button class="btn join-item" :disabled="loading" @click="loadDirectory">跳转</button>
        </div>

        <!-- 文件列表 -->
        <div v-if="loading" class="flex justify-center py-8">
          <span class="loading loading-spinner loading-lg"></span>
        </div>
        <table v-else class="table table-zebra w-full">
          <tbody>
            <tr
              v-for="file in files"
              :key="file.path"
              class="hover cursor-pointer"
              @dblclick="handleRowDblClick(file)"
            >
              <td class="w-12">
                <Icon
                  v-if="file.is_dir"
                  icon="mdi:folder"
                  class="text-2xl text-primary"
                />
                <Icon
                  v-else
                  icon="mdi:file-document-outline"
                  class="text-2xl text-base-content/60"
                />
              </td>
              <td>{{ file.name }}</td>
              <td class="w-24 text-right text-sm text-base-content/60">
                {{ file.is_dir ? '-' : formatSize(file.size) }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="modal-action">
        <button class="btn btn-ghost" @click="visible = false">取消</button>
        <button class="btn btn-primary" @click="selectCurrentDir">选择此目录</button>
      </div>
    </div>
    <form method="dialog" class="modal-backdrop">
      <button @click="visible = false">close</button>
    </form>
  </dialog>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Icon } from '@iconify/vue'
import { composeApi, type FileInfo } from '@/api/compose'
import { showToast } from '@/utils/toast'

const props = defineProps<{
  modelValue: boolean
  hostId: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void
  (e: 'select', path: string): void
}>()

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
})

const currentPath = ref('/')
const files = ref<FileInfo[]>([])
const loading = ref(false)

const pathParts = computed(() => {
  return currentPath.value.split('/').filter((p) => p || currentPath.value === '/')
})

watch(visible, (val) => {
  if (val && files.value.length === 0) {
    loadDirectory()
  }
})

async function loadDirectory() {
  if (!props.hostId) {
    showToast('请先选择主机', 'warning')
    return
  }

  loading.value = true
  try {
    const res = await composeApi.browseDir(props.hostId, currentPath.value)
    currentPath.value = res.data.data.path
    files.value = res.data.data.files
  } catch (error: unknown) {
    const err = error as { response?: { data?: { message?: string } } }
    showToast(err.response?.data?.message || '加载目录失败', 'error')
  } finally {
    loading.value = false
  }
}

function navigateTo(index: number) {
  const parts = currentPath.value.split('/').filter(Boolean)
  const newPath = '/' + parts.slice(0, index + 1).join('/')
  currentPath.value = newPath || '/'
  loadDirectory()
}

function handleRowDblClick(row: FileInfo) {
  if (row.is_dir) {
    currentPath.value = row.path
    loadDirectory()
  }
}

function selectCurrentDir() {
  emit('select', currentPath.value)
  visible.value = false
}

function formatSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

// 当 hostId 变化时重新加载
watch(
  () => props.hostId,
  () => {
    if (visible.value) {
      loadDirectory()
    }
  }
)
</script>

<style scoped>
.browser-container {
  min-height: 300px;
}

.breadcrumb {
  padding: 8px 12px;
  background: var(--fallback-b2, oklch(var(--b2) / 1));
  border-radius: 0.5rem;
}
</style>
