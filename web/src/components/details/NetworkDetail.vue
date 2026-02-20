<template>
  <div class="network-detail">
    <div v-if="loading" class="flex justify-center py-8">
      <span class="loading loading-spinner loading-lg"></span>
    </div>

    <template v-else-if="network">
      <!-- 操作按钮 -->
      <div class="flex items-center gap-2 mb-4">
        <button
          class="btn btn-sm btn-error"
          @click="handleDelete"
          :disabled="operating || isSystemNetwork"
        >
          <Icon icon="mdi:delete" class="text-lg" />
          删除网络
        </button>
        <span v-if="isSystemNetwork" class="text-sm text-base-content/50">
          系统网络无法删除
        </span>
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
                  <td class="font-mono text-sm">{{ network.id }}</td>
                </tr>
                <tr>
                  <td class="font-medium">名称</td>
                  <td>{{ network.name }}</td>
                </tr>
                <tr>
                  <td class="font-medium">驱动</td>
                  <td>{{ network.driver }}</td>
                </tr>
                <tr>
                  <td class="font-medium">作用域</td>
                  <td>{{ network.scope }}</td>
                </tr>
                <tr>
                  <td class="font-medium">创建时间</td>
                  <td>{{ network.created }}</td>
                </tr>
                <tr>
                  <td class="font-medium">内部网络</td>
                  <td>
                    <span :class="['badge badge-sm', network.internal ? 'badge-warning' : 'badge-ghost']">
                      {{ network.internal ? '是' : '否' }}
                    </span>
                  </td>
                </tr>
                <tr>
                  <td class="font-medium">可附加</td>
                  <td>
                    <span :class="['badge badge-sm', network.attachable ? 'badge-success' : 'badge-ghost']">
                      {{ network.attachable ? '是' : '否' }}
                    </span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- IPAM 配置 -->
      <div v-if="network.ipam?.config?.length" class="card bg-base-100 shadow mt-4">
        <div class="card-body">
          <h3 class="card-title text-base mb-4">IPAM 配置</h3>
          <div class="overflow-x-auto">
            <table class="table table-sm">
              <thead>
                <tr>
                  <th>子网</th>
                  <th>网关</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(config, index) in network.ipam.config" :key="index">
                  <td class="font-mono">{{ config.subnet }}</td>
                  <td class="font-mono">{{ config.gateway || '-' }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <!-- 标签 -->
      <div v-if="Object.keys(network.labels || {}).length" class="card bg-base-100 shadow mt-4">
        <div class="card-body">
          <h3 class="card-title text-base mb-4">标签</h3>
          <div class="flex flex-wrap gap-2">
            <span
              v-for="(value, key) in network.labels"
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
      <Icon icon="mdi:lan-disconnect" class="text-4xl mb-2" />
      <p>无法加载网络信息</p>
    </div>

    <Confirm ref="confirmRef" />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { Icon } from '@iconify/vue'
import { networkApi, type Network } from '@/api'
import { showToast } from '@/utils/toast'
import { useResourceStore } from '@/stores'
import Confirm from '@/components/Confirm.vue'

const props = defineProps<{
  hostId: string
  networkId: string
}>()

const resourceStore = useResourceStore()

const network = ref<Network | null>(null)
const loading = ref(false)
const operating = ref(false)
const confirmRef = ref<InstanceType<typeof Confirm> | null>(null)

// 系统网络（bridge, host, none）无法删除
const isSystemNetwork = computed(() => {
  const systemNetworks = ['bridge', 'host', 'none']
  return systemNetworks.includes(network.value?.name || '')
})

async function loadNetwork() {
  loading.value = true
  try {
    const res = await networkApi.get(props.hostId, props.networkId)
    network.value = res.data.data
  } catch {
    network.value = null
  } finally {
    loading.value = false
  }
}

async function handleDelete() {
  if (isSystemNetwork.value) return

  const result = await confirmRef.value?.confirm({
    title: '删除确认',
    message: `确定删除网络 "${network.value?.name}"?`,
    type: 'danger'
  })
  if (result && result.confirmed) {
    operating.value = true
    try {
      await networkApi.remove(props.hostId, props.networkId)
      showToast('网络已删除', 'success')
      resourceStore.refreshHostResources(props.hostId)
    } finally {
      operating.value = false
    }
  }
}

watch(() => props.networkId, loadNetwork)
onMounted(loadNetwork)
</script>
