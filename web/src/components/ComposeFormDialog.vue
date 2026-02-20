<template>
  <dialog :class="['modal', { 'modal-open': visible }]">
    <div class="modal-box w-full max-w-2xl">
      <h3 class="font-bold text-lg mb-4">新建 Compose 项目</h3>

      <div class="form-control mb-4">
        <label class="label">
          <span class="label-text">项目名称 <span class="text-error">*</span></span>
        </label>
        <input
          v-model="form.name"
          type="text"
          placeholder="my-project"
          class="input input-bordered w-full"
        />
      </div>

      <div class="form-control mb-4">
        <label class="label">
          <span class="label-text">源类型 <span class="text-error">*</span></span>
        </label>
        <div class="flex gap-4">
          <label class="label cursor-pointer gap-2">
            <input type="radio" v-model="form.source_type" value="content" class="radio radio-primary" />
            <span class="label-text">直接输入内容</span>
          </label>
          <label class="label cursor-pointer gap-2">
            <input type="radio" v-model="form.source_type" value="directory" class="radio radio-primary" />
            <span class="label-text">指定目录</span>
          </label>
        </div>
      </div>

      <!-- 内容模式 -->
      <template v-if="form.source_type === 'content'">
        <div class="form-control mb-4">
          <label class="label">
            <span class="label-text">Compose 内容 <span class="text-error">*</span></span>
          </label>
          <textarea
            v-model="form.content"
            class="textarea textarea-bordered h-64 w-full font-mono text-sm"
            placeholder="version: '3'
services:
  web:
    image: nginx
    ports:
      - '80:80'"
          ></textarea>
        </div>
      </template>

      <!-- 目录模式 -->
      <template v-else-if="form.source_type === 'directory'">
        <div class="form-control mb-4">
          <label class="label">
            <span class="label-text">工作目录 <span class="text-error">*</span></span>
          </label>
          <div class="join w-full">
            <input
              v-model="form.work_dir"
              type="text"
              placeholder="/path/to/project"
              class="input input-bordered join-item flex-1"
            />
            <button
              class="btn join-item"
              :disabled="!hostId"
              @click="showDirBrowser = true"
            >浏览</button>
          </div>
        </div>

        <div class="form-control mb-4">
          <label class="label">
            <span class="label-text">Compose 文件</span>
          </label>
          <select
            v-model="form.compose_file"
            class="select select-bordered w-full"
            :disabled="!form.work_dir"
            @focus="scanComposeFiles"
          >
            <option value="">选择 compose 文件</option>
            <option v-for="file in composeFiles" :key="file" :value="file">{{ file }}</option>
          </select>
          <label class="label">
            <span class="label-text-alt text-base-content/60">留空则默认使用 docker-compose.yml</span>
          </label>
        </div>

        <div class="form-control mb-4">
          <label class="label">
            <span class="label-text">环境变量文件</span>
          </label>
          <select
            v-model="form.env_file"
            class="select select-bordered w-full"
            :disabled="!form.work_dir"
            @focus="scanComposeFiles"
          >
            <option value="">选择环境变量文件（可选）</option>
            <option v-for="file in envFiles" :key="file" :value="file">{{ file }}</option>
          </select>
        </div>

        <div class="divider">或者上传目录</div>

        <div class="form-control mb-4">
          <label class="label">
            <span class="label-text">上传目录</span>
          </label>
          <DirectoryUpload
            ref="uploadRef"
            :host-id="hostId"
            :initial-path="form.work_dir"
            @browse="showDirBrowser = true"
            @success="handleUploadSuccess"
          />
        </div>
      </template>

      <div class="modal-action">
        <button class="btn btn-ghost" @click="handleCancel">取消</button>
        <button class="btn btn-primary" @click="handleSave" :disabled="saving">
          <span v-if="saving" class="loading loading-spinner loading-sm"></span>
          创建
        </button>
      </div>
    </div>
    <form method="dialog" class="modal-backdrop">
      <button @click="handleCancel">close</button>
    </form>

    <!-- 目录浏览对话框 -->
    <DirBrowser
      v-model="showDirBrowser"
      :host-id="hostId"
      @select="handleDirSelect"
    />
  </dialog>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { composeApi } from '@/api'
import { showToast } from '@/utils/toast'
import DirBrowser from '@/components/DirBrowser.vue'
import DirectoryUpload from '@/components/DirectoryUpload.vue'

const props = defineProps<{
  visible: boolean
  hostId: string
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'saved'): void
}>()

const saving = ref(false)
const showDirBrowser = ref(false)
const uploadRef = ref<InstanceType<typeof DirectoryUpload>>()
const composeFiles = ref<string[]>([])
const envFiles = ref<string[]>([])

const form = reactive({
  name: '',
  source_type: 'content' as 'content' | 'directory',
  content: '',
  work_dir: '',
  compose_file: '',
  env_file: '',
})

function resetForm() {
  Object.assign(form, {
    name: '',
    source_type: 'content',
    content: '',
    work_dir: '',
    compose_file: '',
    env_file: '',
  })
  composeFiles.value = []
  envFiles.value = []
}

watch(() => props.visible, (val) => {
  if (val) {
    resetForm()
  }
})

async function scanComposeFiles() {
  if (!props.hostId || !form.work_dir) return

  try {
    const res = await composeApi.scanComposeFiles(props.hostId, form.work_dir)
    composeFiles.value = res.data.data.compose_files
    envFiles.value = res.data.data.env_files

    if (composeFiles.value.length === 0) {
      showToast('未在该目录中找到 compose 文件', 'warning')
    }
  } catch (error: unknown) {
    const err = error as { response?: { data?: { message?: string } } }
    showToast(err.response?.data?.message || '扫描目录失败', 'error')
  }
}

function handleDirSelect(path: string) {
  form.work_dir = path
  if (uploadRef.value) {
    uploadRef.value.setTargetPath(path)
  }
  scanComposeFiles()
}

function handleUploadSuccess(path: string) {
  form.work_dir = path
  scanComposeFiles()
}

function handleCancel() {
  emit('update:visible', false)
  resetForm()
}

async function handleSave() {
  if (!form.name) {
    showToast('请输入项目名称', 'warning')
    return
  }

  if (form.source_type === 'content') {
    if (!form.content) {
      showToast('请输入 Compose 内容', 'warning')
      return
    }
  } else if (form.source_type === 'directory') {
    if (!form.work_dir) {
      showToast('请指定工作目录', 'warning')
      return
    }
  }

  saving.value = true
  try {
    const data = {
      name: form.name,
      host_id: props.hostId,
      source_type: form.source_type,
      content: form.source_type === 'content' ? form.content : undefined,
      work_dir: form.source_type === 'directory' ? form.work_dir : undefined,
      compose_file: form.source_type === 'directory' && form.compose_file ? form.compose_file : undefined,
      env_file: form.source_type === 'directory' && form.env_file ? form.env_file : undefined,
    }

    await composeApi.create(data as unknown as Parameters<typeof composeApi.create>[0])
    showToast('创建成功', 'success')
    emit('update:visible', false)
    emit('saved')
  } catch (error: unknown) {
    const err = error as { response?: { data?: { message?: string } } }
    showToast(err.response?.data?.message || '创建失败', 'error')
  } finally {
    saving.value = false
  }
}
</script>
