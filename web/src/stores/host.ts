import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { hostApi, type Host } from '@/api'

export const useHostStore = defineStore('host', () => {
  const hosts = ref<Host[]>([])
  const currentHostId = ref<string>('')
  const loading = ref(false)

  // 当前主机
  const currentHost = computed(() =>
    hosts.value.find(h => h.id === currentHostId.value)
  )

  // 默认主机
  const defaultHost = computed(() =>
    hosts.value.find(h => h.is_default) || hosts.value[0]
  )

  // 加载主机列表
  async function loadHosts() {
    loading.value = true
    try {
      const res = await hostApi.list()
      hosts.value = res.data.data
      // 自动选择默认主机
      if (!currentHostId.value && hosts.value.length > 0) {
        currentHostId.value = defaultHost.value?.id || hosts.value[0]!.id
      }
    } finally {
      loading.value = false
    }
  }

  // 设置当前主机
  function setCurrentHost(id: string) {
    currentHostId.value = id
    localStorage.setItem('currentHostId', id)
  }

  // 初始化时从 localStorage 恢复
  function init() {
    const savedId = localStorage.getItem('currentHostId')
    if (savedId) {
      currentHostId.value = savedId
    }
    loadHosts()
  }

  return {
    hosts,
    currentHostId,
    currentHost,
    defaultHost,
    loading,
    loadHosts,
    setCurrentHost,
    init,
  }
})
