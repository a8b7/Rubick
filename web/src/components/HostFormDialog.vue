<template>
  <dialog :class="['modal', { 'modal-open': visible }]">
    <div class="modal-box w-full max-w-lg">
      <h3 class="font-bold text-lg mb-4">{{ editingHost ? '编辑主机' : '添加主机' }}</h3>

      <div class="form-control mb-4">
        <label class="label">
          <span class="label-text">名称 <span class="text-error">*</span></span>
        </label>
        <input v-model="form.name" type="text" placeholder="主机名称" class="input input-bordered w-full" />
      </div>

      <div class="form-control mb-4">
        <label class="label">
          <span class="label-text">类型 <span class="text-error">*</span></span>
        </label>
        <select v-model="form.type" class="select select-bordered w-full" :disabled="!!editingHost">
          <option value="local">本地</option>
          <option value="tcp">TCP</option>
          <option value="ssh">SSH</option>
        </select>
      </div>

      <template v-if="form.type !== 'local'">
        <div class="form-control mb-4">
          <label class="label">
            <span class="label-text">主机地址</span>
          </label>
          <input v-model="form.host" type="text" placeholder="IP 或域名" class="input input-bordered w-full" />
        </div>

        <div class="form-control mb-4">
          <label class="label">
            <span class="label-text">Docker 端口</span>
          </label>
          <input v-model.number="form.docker_port" type="number" min="1" max="65535" class="input input-bordered w-full" />
        </div>
      </template>

      <template v-if="form.type === 'ssh'">
        <div class="form-control mb-4">
          <label class="label">
            <span class="label-text">SSH 用户</span>
          </label>
          <input v-model="form.ssh_user" type="text" placeholder="root" class="input input-bordered w-full" />
        </div>

        <div class="form-control mb-4">
          <label class="label">
            <span class="label-text">SSH 端口</span>
          </label>
          <input v-model.number="form.ssh_port" type="number" min="1" max="65535" class="input input-bordered w-full" />
        </div>

        <div class="form-control mb-4">
          <label class="label">
            <span class="label-text">认证方式</span>
          </label>
          <select v-model="form.ssh_auth_type" class="select select-bordered w-full">
            <option value="password">密码</option>
            <option value="key">私钥</option>
          </select>
        </div>

        <div v-if="form.ssh_auth_type === 'password'" class="form-control mb-4">
          <label class="label">
            <span class="label-text">密码</span>
          </label>
          <input
            v-model="form.ssh_password"
            type="password"
            :placeholder="editingHost ? '留空保持原密码不变' : '请输入密码'"
            class="input input-bordered w-full"
          />
        </div>

        <div v-else class="form-control mb-4">
          <label class="label">
            <span class="label-text">私钥</span>
          </label>
          <textarea v-model="form.ssh_private_key" class="textarea textarea-bordered h-24 w-full" :placeholder="editingHost ? '留空保持原私钥不变' : ''"></textarea>
        </div>
      </template>

      <div class="form-control mb-4">
        <label class="label">
          <span class="label-text">描述</span>
        </label>
        <textarea v-model="form.description" class="textarea textarea-bordered h-16 w-full"></textarea>
      </div>

      <div class="modal-action">
        <button class="btn btn-ghost" @click="handleCancel">取消</button>
        <button class="btn btn-primary" @click="handleSave" :disabled="saving">
          <span v-if="saving" class="loading loading-spinner loading-sm"></span>
          保存
        </button>
      </div>
    </div>
    <form method="dialog" class="modal-backdrop">
      <button @click="handleCancel">close</button>
    </form>
  </dialog>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { hostApi, type Host, type CreateHostRequest } from '@/api'
import { showToast } from '@/utils/toast'
import { useHostStore } from '@/stores'

const props = defineProps<{
  visible: boolean
  host?: Host | null
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
  (e: 'saved'): void
}>()

const hostStore = useHostStore()
const saving = ref(false)

const editingHost = ref<Host | null>(null)

const form = reactive<CreateHostRequest & { id?: string }>({
  name: '',
  type: 'local',
  host: '',
  ssh_user: '',
  ssh_auth_type: 'password',
  ssh_password: '',
  ssh_private_key: '',
  ssh_port: 22,
  docker_port: 2375,
  description: '',
})

function resetForm() {
  editingHost.value = null
  Object.assign(form, {
    name: '',
    type: 'local',
    host: '',
    ssh_user: '',
    ssh_auth_type: 'password',
    ssh_password: '',
    ssh_private_key: '',
    ssh_port: 22,
    docker_port: 2375,
    description: '',
  })
}

// 监听 visible 变化，初始化表单
watch(() => props.visible, (val) => {
  if (val) {
    if (props.host) {
      // 编辑模式
      editingHost.value = props.host
      Object.assign(form, {
        id: props.host.id,
        name: props.host.name,
        type: props.host.type,
        host: props.host.host || '',
        ssh_user: props.host.ssh_user || '',
        ssh_auth_type: props.host.ssh_auth_type || 'password',
        ssh_password: '',
        ssh_private_key: '',
        ssh_port: props.host.ssh_port || 22,
        docker_port: props.host.docker_port || 2375,
        description: props.host.description || '',
      })
    } else {
      // 新建模式
      resetForm()
    }
  }
})

function handleCancel() {
  emit('update:visible', false)
  resetForm()
}

async function handleSave() {
  if (!form.name) {
    showToast('请输入主机名称', 'warning')
    return
  }

  // 新建 SSH 主机时，密码认证方式必须提供密码
  if (!editingHost.value && form.type === 'ssh' && form.ssh_auth_type === 'password' && !form.ssh_password) {
    showToast('请输入 SSH 密码', 'warning')
    return
  }

  saving.value = true
  try {
    if (editingHost.value) {
      // 编辑模式：构建更新数据，密码为空时不更新密码字段
      const updateData = { ...form }
      if (form.ssh_auth_type === 'password' && !form.ssh_password) {
        delete updateData.ssh_password
      }
      if (form.ssh_auth_type === 'key' && !form.ssh_private_key) {
        delete updateData.ssh_private_key
      }
      await hostApi.update(editingHost.value.id, updateData)
      showToast('更新成功', 'success')
    } else {
      await hostApi.create(form)
      showToast('添加成功', 'success')
    }
    emit('update:visible', false)
    emit('saved')
    hostStore.loadHosts()
  } finally {
    saving.value = false
  }
}
</script>
