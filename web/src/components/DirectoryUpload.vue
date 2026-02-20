<template>
  <div class="upload-container border border-base-300 rounded-lg p-4">
    <div
      class="upload-dropzone border-2 border-dashed border-base-300 rounded-lg p-10 text-center cursor-pointer transition-all"
      :class="{ 'border-primary bg-primary/5': isDragover }"
      @dragover.prevent="isDragover = true"
      @dragleave.prevent="isDragover = false"
      @drop.prevent="handleDrop"
      @click="triggerFileSelect"
    >
      <input
        ref="fileInput"
        type="file"
        webkitdirectory
        directory
        multiple
        class="hidden"
        @change="handleFileSelect"
      />
      <Icon icon="mdi:cloud-upload" class="text-5xl text-base-content/40 mb-4" />
      <div class="upload-text">
        <p>拖拽目录到此处，或点击选择目录</p>
        <p class="text-sm text-base-content/60 mt-2">支持 webkitdirectory 的浏览器可选择整个目录</p>
      </div>
    </div>

    <!-- 已选择的文件列表 -->
    <div v-if="selectedFiles.length > 0" class="mt-4">
      <div class="flex justify-between items-center py-2 border-b border-base-300">
        <span>已选择 {{ selectedFiles.length }} 个文件</span>
        <button class="btn btn-sm btn-error" @click="clearFiles">清空</button>
      </div>
      <div class="max-h-52 overflow-auto">
        <div class="file-item" v-for="(file, index) in displayFiles" :key="index">
          <Icon icon="mdi:file-document-outline" class="text-xl text-base-content/60" />
          <span class="file-name" :title="file.webkitRelativePath || file.name">
            {{ file.webkitRelativePath || file.name }}
          </span>
          <span class="text-sm text-base-content/60">{{ formatSize(file.size) }}</span>
        </div>
        <div v-if="selectedFiles.length > 10" class="text-center py-2 text-base-content/60 text-sm">
          还有 {{ selectedFiles.length - 10 }} 个文件...
        </div>
      </div>
    </div>

    <!-- 目标路径 -->
    <div class="mt-4">
      <div class="join w-full">
        <span class="btn join-item pointer-events-none">目标路径</span>
        <input
          v-model="targetPath"
          type="text"
          placeholder="输入服务器目标路径，如 /opt/myapp"
          class="input input-bordered join-item flex-1"
        />
        <button class="btn join-item" :disabled="!hostId" @click="emit('browse')">浏览</button>
      </div>
    </div>

    <!-- 上传按钮 -->
    <div class="mt-4 text-right">
      <button
        class="btn btn-primary"
        :disabled="selectedFiles.length === 0 || !targetPath"
        @click="upload"
      >
        <span v-if="uploading" class="loading loading-spinner loading-sm"></span>
        上传到服务器
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { Icon } from '@iconify/vue'
import { composeApi } from '@/api/compose'
import { showToast } from '@/utils/toast'

const props = defineProps<{
  hostId: string
  initialPath?: string
}>()

const emit = defineEmits<{
  (e: 'browse'): void
  (e: 'success', path: string): void
}>()

const fileInput = ref<HTMLInputElement>()
const selectedFiles = ref<File[]>([])
const targetPath = ref(props.initialPath || '')
const isDragover = ref(false)
const uploading = ref(false)

const displayFiles = computed(() => selectedFiles.value.slice(0, 10))

function triggerFileSelect() {
  fileInput.value?.click()
}

function handleFileSelect(event: Event) {
  const input = event.target as HTMLInputElement
  if (input.files && input.files.length > 0) {
    selectedFiles.value = Array.from(input.files)
    // 自动设置目标路径为项目名称
    const firstFile = selectedFiles.value[0]
    if (firstFile && firstFile.webkitRelativePath) {
      const firstPath = firstFile.webkitRelativePath
      const dirName = firstPath.split('/')[0]
      if (!targetPath.value) {
        targetPath.value = `/opt/${dirName}`
      }
    }
  }
}

function handleDrop(event: DragEvent) {
  isDragover.value = false
  const items = event.dataTransfer?.items
  if (items) {
    const files: File[] = []
    for (let i = 0; i < items.length; i++) {
      const item = items[i]
      if (item && item.kind === 'file') {
        const entry = item.webkitGetAsEntry?.()
        if (entry) {
          traverseFileTree(entry, files).then(() => {
            selectedFiles.value = files
          })
        }
      }
    }
  }
}

async function traverseFileTree(entry: FileSystemEntry, files: File[]): Promise<void> {
  if (entry.isFile) {
    const fileEntry = entry as FileSystemFileEntry
    const file = await new Promise<File>((resolve) => {
      fileEntry.file(resolve)
    })
    // 创建带有相对路径的文件对象
    const fileWithPath = new File([file], file.name, { type: file.type })
    Object.defineProperty(fileWithPath, 'webkitRelativePath', {
      value: fileEntry.fullPath.startsWith('/') ? fileEntry.fullPath.slice(1) : fileEntry.fullPath,
    })
    files.push(fileWithPath)
  } else if (entry.isDirectory) {
    const dirEntry = entry as FileSystemDirectoryEntry
    const reader = dirEntry.createReader()
    const entries = await new Promise<FileSystemEntry[]>((resolve) => {
      reader.readEntries(resolve)
    })
    for (const childEntry of entries) {
      await traverseFileTree(childEntry, files)
    }
  }
}

function clearFiles() {
  selectedFiles.value = []
  if (fileInput.value) {
    fileInput.value.value = ''
  }
}

async function upload() {
  if (!props.hostId) {
    showToast('请先选择主机', 'warning')
    return
  }
  if (selectedFiles.value.length === 0) {
    showToast('请选择要上传的文件', 'warning')
    return
  }
  if (!targetPath.value) {
    showToast('请输入目标路径', 'warning')
    return
  }

  uploading.value = true
  try {
    // 创建 FileList 类型的对象
    const dataTransfer = new DataTransfer()
    for (const file of selectedFiles.value) {
      dataTransfer.items.add(file)
    }

    await composeApi.uploadDirectory(props.hostId, targetPath.value, dataTransfer.files)
    showToast('上传成功', 'success')
    emit('success', targetPath.value)
    clearFiles()
  } catch (error: unknown) {
    const err = error as { response?: { data?: { message?: string } } }
    showToast(err.response?.data?.message || '上传失败', 'error')
  } finally {
    uploading.value = false
  }
}

function formatSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

function setTargetPath(path: string) {
  targetPath.value = path
}

defineExpose({ setTargetPath })
</script>

<style scoped>
.file-item {
  display: flex;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid var(--fallback-bc, oklch(var(--bc) / 0.1));
}

.file-name {
  flex: 1;
  margin-left: 8px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
