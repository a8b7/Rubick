<template>
  <div class="page-container">
    <div class="card bg-base-100 shadow-xl">
      <div class="card-body">
        <div class="flex justify-between items-center mb-4">
          <h2 class="card-title">Compose 项目</h2>
          <button class="btn btn-primary" @click="$router.push('/compose/create')">新建项目</button>
        </div>

        <div v-if="loading" class="flex justify-center py-8">
          <span class="loading loading-spinner loading-lg"></span>
        </div>
        <table v-else class="table table-zebra">
          <thead>
            <tr>
              <th>项目名称</th>
              <th>主机</th>
              <th>状态</th>
              <th>创建时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="project in projects" :key="project.id">
              <td>{{ project.name }}</td>
              <td>{{ project.host?.name || '-' }}</td>
              <td>
                <span class="badge" :class="project.status === 'running' ? 'badge-success' : 'badge-info'">
                  {{ project.status }}
                </span>
              </td>
              <td>{{ new Date(project.created_at).toLocaleString() }}</td>
              <td>
                <div class="join">
                  <button
                    class="btn btn-sm join-item btn-success"
                    :disabled="project.status === 'running'"
                    @click="composeUp(project)"
                  >启动</button>
                  <button
                    class="btn btn-sm join-item btn-warning"
                    :disabled="project.status !== 'running'"
                    @click="composeDown(project)"
                  >停止</button>
                  <button
                    class="btn btn-sm join-item"
                    @click="$router.push(`/compose/${project.id}`)"
                  >详情</button>
                  <button
                    class="btn btn-sm join-item"
                    @click="$router.push(`/compose/${project.id}/edit`)"
                  >编辑</button>
                  <button
                    class="btn btn-sm join-item btn-error"
                    @click="deleteProject(project)"
                  >删除</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <Confirm ref="confirmRef" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { composeApi, type ComposeProject } from '@/api'
import { showToast } from '@/utils/toast'
import Confirm from '@/components/Confirm.vue'

const projects = ref<ComposeProject[]>([])
const loading = ref(false)
const confirmRef = ref<InstanceType<typeof Confirm> | null>(null)

async function loadProjects() {
  loading.value = true
  try {
    const res = await composeApi.list()
    projects.value = res.data.data
  } finally {
    loading.value = false
  }
}

async function composeUp(project: ComposeProject) {
  try {
    await composeApi.up(project.id, { detach: true })
    showToast('项目已启动', 'success')
    loadProjects()
  } catch {}
}

async function composeDown(project: ComposeProject) {
  try {
    await composeApi.down(project.id)
    showToast('项目已停止', 'success')
    loadProjects()
  } catch {}
}

async function deleteProject(project: ComposeProject) {
  const result = await confirmRef.value?.confirm({
    title: '删除确认',
    message: `确定删除项目 "${project.name}"?`,
    type: 'danger'
  })
  if (result && result.confirmed) {
    await composeApi.delete(project.id)
    showToast('项目已删除', 'success')
    loadProjects()
  }
}

onMounted(loadProjects)
</script>
