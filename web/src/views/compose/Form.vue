<template>
  <div class="page-container">
    <div class="card bg-base-100 shadow-xl">
      <div class="card-body">
        <div class="flex items-center gap-4 mb-4">
          <button class="btn btn-ghost" @click="$router.back()">
            <Icon icon="mdi:arrow-left" class="text-xl" />
            返回
          </button>
          <h2 class="card-title">{{ isEdit ? '编辑项目' : '新建项目' }}</h2>
        </div>

        <div class="max-w-3xl">
          <div class="form-control mb-4">
            <label class="label">
              <span class="label-text">项目名称 <span class="text-error">*</span></span>
            </label>
            <input
              v-model="form.name"
              type="text"
              placeholder="my-project"
              class="input input-bordered w-full"
              :disabled="isEdit"
            />
          </div>

          <div class="form-control mb-4">
            <label class="label">
              <span class="label-text">主机 <span class="text-error">*</span></span>
            </label>
            <select v-model="form.host_id" class="select select-bordered w-full" :disabled="isEdit">
              <option value="" disabled>选择主机</option>
              <option v-for="h in hostStore.hosts" :key="h.id" :value="h.id">{{ h.name }}</option>
            </select>
          </div>

          <div v-if="!isEdit" class="form-control mb-4">
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
                class="textarea textarea-bordered h-80 w-full font-mono"
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
                  :disabled="!form.host_id"
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
                :host-id="form.host_id"
                :initial-path="form.work_dir"
                @browse="showDirBrowser = true"
                @success="handleUploadSuccess"
              />
            </div>
          </template>

          <div class="flex gap-2 mt-6">
            <button class="btn btn-primary" @click="save" :disabled="saving">
              <span v-if="saving" class="loading loading-spinner loading-sm"></span>
              保存
            </button>
            <button class="btn btn-ghost" @click="$router.back()">取消</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 目录浏览对话框 -->
    <DirBrowser
      v-model="showDirBrowser"
      :host-id="form.host_id"
      @select="handleDirSelect"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { Icon } from '@iconify/vue'
import { useHostStore } from '@/stores'
import { composeApi } from '@/api'
import { showToast } from '@/utils/toast'
import DirBrowser from '@/components/DirBrowser.vue'
import DirectoryUpload from '@/components/DirectoryUpload.vue'

const route = useRoute()
const router = useRouter()
const hostStore = useHostStore()

const isEdit = computed(() => !!route.params.id)
const saving = ref(false)
const showDirBrowser = ref(false)
const uploadRef = ref<InstanceType<typeof DirectoryUpload>>()
const composeFiles = ref<string[]>([])
const envFiles = ref<string[]>([])

const form = reactive({
  name: '',
  host_id: '',
  source_type: 'content' as 'content' | 'directory',
  content: '',
  work_dir: '',
  compose_file: '',
  env_file: '',
})

onMounted(async () => {
  await hostStore.loadHosts()

  if (isEdit.value) {
    try {
      const res = await composeApi.get(route.params.id as string)
      const project = res.data.data
      form.name = project.name
      form.host_id = project.host_id
      form.source_type = project.source_type || 'content'
      form.content = project.content || ''
      form.work_dir = project.work_dir || ''
      form.compose_file = project.compose_file || ''
      form.env_file = project.env_file || ''
    } catch {
      router.back()
    }
  } else if (hostStore.currentHostId) {
    form.host_id = hostStore.currentHostId
  }
})

async function scanComposeFiles() {
  if (!form.host_id || !form.work_dir) return

  try {
    const res = await composeApi.scanComposeFiles(form.host_id, form.work_dir)
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
  // 自动扫描 compose 文件
  scanComposeFiles()
}

function handleUploadSuccess(path: string) {
  form.work_dir = path
  scanComposeFiles()
}

async function save() {
  // 验证必填字段
  if (!form.name) {
    showToast('请输入项目名称', 'warning')
    return
  }
  if (!form.host_id) {
    showToast('请选择主机', 'warning')
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
    const data: Record<string, unknown> = {
      name: form.name,
      host_id: form.host_id,
      source_type: form.source_type,
    }

    if (form.source_type === 'content') {
      data.content = form.content
    } else {
      data.work_dir = form.work_dir
      if (form.compose_file) data.compose_file = form.compose_file
      if (form.env_file) data.env_file = form.env_file
    }

    if (isEdit.value) {
      await composeApi.update(route.params.id as string, data as Parameters<typeof composeApi.update>[1])
      showToast('更新成功', 'success')
    } else {
      await composeApi.create(data as unknown as Parameters<typeof composeApi.create>[0])
      showToast('创建成功', 'success')
    }
    router.push('/compose')
  } catch (error: unknown) {
    const err = error as { response?: { data?: { message?: string } } }
    showToast(err.response?.data?.message || '保存失败', 'error')
  } finally {
    saving.value = false
  }
}
</script>
