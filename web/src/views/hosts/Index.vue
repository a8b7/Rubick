<template>
  <div class="page-container">
    <div class="card bg-base-100 shadow-xl">
      <div class="card-body">
        <div class="flex justify-between items-center mb-4">
          <h2 class="card-title">主机列表</h2>
          <button class="btn btn-primary" @click="showDialog = true">添加主机</button>
        </div>

        <div v-if="hostStore.loading" class="flex justify-center py-8">
          <span class="loading loading-spinner loading-lg"></span>
        </div>
        <table v-else class="table table-zebra">
          <thead>
            <tr>
              <th>名称</th>
              <th>类型</th>
              <th>地址</th>
              <th>状态</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="host in hostStore.hosts" :key="host.id">
              <td>{{ host.name }}</td>
              <td>
                <span class="badge" :class="host.type === 'local' ? 'badge-success' : 'badge-primary'">
                  {{ host.type }}
                </span>
              </td>
              <td>{{ host.host || '-' }}</td>
              <td>
                <span class="badge" :class="host.is_active ? 'badge-success' : 'badge-error'">
                  {{ host.is_active ? '活跃' : '离线' }}
                </span>
              </td>
              <td>
                <div class="join">
                  <button class="btn btn-sm join-item" @click="testConnection(host)">测试</button>
                  <button class="btn btn-sm join-item btn-primary" @click="editHost(host)">编辑</button>
                  <button class="btn btn-sm join-item btn-error" @click="deleteHost(host)">删除</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 添加/编辑主机对话框 -->
    <dialog :class="['modal', { 'modal-open': showDialog }]">
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
          <select v-model="form.type" class="select select-bordered w-full">
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
            <textarea v-model="form.ssh_private_key" class="textarea textarea-bordered h-24 w-full"></textarea>
          </div>
        </template>

        <div class="form-control mb-4">
          <label class="label">
            <span class="label-text">描述</span>
          </label>
          <textarea v-model="form.description" class="textarea textarea-bordered h-16 w-full"></textarea>
        </div>

        <div class="modal-action">
          <button class="btn btn-ghost" @click="showDialog = false">取消</button>
          <button class="btn btn-primary" @click="saveHost" :disabled="saving">
            <span v-if="saving" class="loading loading-spinner loading-sm"></span>
            保存
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button @click="showDialog = false">close</button>
      </form>
    </dialog>

    <Confirm ref="confirmRef" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { useHostStore } from '@/stores'
import { hostApi, type Host, type CreateHostRequest } from '@/api'
import { showToast } from '@/utils/toast'
import Confirm from '@/components/Confirm.vue'

const hostStore = useHostStore()
const showDialog = ref(false)
const saving = ref(false)
const editingHost = ref<Host | null>(null)
const confirmRef = ref<InstanceType<typeof Confirm> | null>(null)

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

watch(showDialog, (val) => {
  if (!val) {
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
})

function editHost(host: Host) {
  editingHost.value = host
  Object.assign(form, {
    id: host.id,
    name: host.name,
    type: host.type,
    host: host.host || '',
    ssh_user: host.ssh_user || '',
    ssh_auth_type: host.ssh_auth_type || 'password',
    ssh_password: '',
    ssh_private_key: '',
    ssh_port: host.ssh_port || 22,
    docker_port: host.docker_port || 2375,
    description: host.description || '',
  })
  showDialog.value = true
}

async function saveHost() {
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
    showDialog.value = false
    hostStore.loadHosts()
  } finally {
    saving.value = false
  }
}

async function testConnection(host: Host) {
  try {
    const res = await hostApi.test(host.id)
    if (res.data.data.success) {
      showToast('连接成功', 'success')
    } else {
      showToast(res.data.data.message || '连接失败', 'error')
    }
  } catch {
    // 错误已在拦截器处理
  }
}

async function deleteHost(host: Host) {
  const result = await confirmRef.value?.confirm({
    title: '删除确认',
    message: `确定删除主机 "${host.name}"?`,
    type: 'danger'
  })
  if (result && result.confirmed) {
    await hostApi.delete(host.id)
    showToast('删除成功', 'success')
    hostStore.loadHosts()
  }
}
</script>
