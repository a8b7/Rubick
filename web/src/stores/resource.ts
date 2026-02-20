import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { containerApi, type Container } from '@/api/container'
import { imageApi, type Image } from '@/api/image'
import { volumeApi, type Volume } from '@/api/volume'
import { networkApi, type Network } from '@/api/network'
import { composeApi, type ComposeProject } from '@/api/compose'

// 资源类型
export type ResourceType = 'host' | 'container' | 'image' | 'volume' | 'network' | 'compose'

// 选中状态
export interface Selection {
  hostId: string
  type: ResourceType
  id: string
  name: string
}

// 资源缓存
export interface ResourcesCache {
  containers: Container[]
  images: Image[]
  volumes: Volume[]
  networks: Network[]
  composeProjects: ComposeProject[]
  loading: boolean
  loaded: boolean
}

export const useResourceStore = defineStore('resource', () => {
  // 当前选中项
  const selection = ref<Selection | null>(null)

  // 展开的主机节点
  const expandedHosts = ref<Set<string>>(new Set())

  // 展开的资源分组（格式：hostId:type）
  const expandedGroups = ref<Set<string>>(new Set())

  // 资源缓存（按主机ID索引）
  const resourcesCache = ref<Map<string, ResourcesCache>>(new Map())

  // 当前选中的主机ID
  const selectedHostId = computed(() => selection.value?.hostId || null)

  // 设置选中项
  function setSelection(sel: Selection | null) {
    selection.value = sel
  }

  // 清除选中项
  function clearSelection() {
    selection.value = null
  }

  // 切换主机展开状态
  function toggleHostExpand(hostId: string) {
    if (expandedHosts.value.has(hostId)) {
      expandedHosts.value.delete(hostId)
      // 清除该主机的资源缓存
      resourcesCache.value.delete(hostId)
    } else {
      expandedHosts.value.add(hostId)
      // 加载该主机的资源
      loadHostResources(hostId)
    }
  }

  // 切换资源分组展开状态
  function toggleGroupExpand(key: string) {
    if (expandedGroups.value.has(key)) {
      expandedGroups.value.delete(key)
    } else {
      expandedGroups.value.add(key)
    }
  }

  // 检查主机是否展开
  function isHostExpanded(hostId: string): boolean {
    return expandedHosts.value.has(hostId)
  }

  // 检查分组是否展开
  function isGroupExpanded(key: string): boolean {
    return expandedGroups.value.has(key)
  }

  // 获取主机资源缓存
  function getHostCache(hostId: string): ResourcesCache | undefined {
    return resourcesCache.value.get(hostId)
  }

  // 加载主机资源
  async function loadHostResources(hostId: string) {
    let cache = resourcesCache.value.get(hostId)
    if (!cache) {
      cache = {
        containers: [],
        images: [],
        volumes: [],
        networks: [],
        composeProjects: [],
        loading: false,
        loaded: false,
      }
      resourcesCache.value.set(hostId, cache)
    }

    if (cache.loading || cache.loaded) return

    cache.loading = true
    try {
      // 并行加载所有资源
      const [containersRes, imagesRes, volumesRes, networksRes, composeRes] = await Promise.all([
        containerApi.list(hostId, true),
        imageApi.list(hostId),
        volumeApi.list(hostId),
        networkApi.list(hostId),
        composeApi.list(hostId),
      ])

      cache.containers = containersRes.data.data || []
      cache.images = imagesRes.data.data || []
      cache.volumes = volumesRes.data.data || []
      cache.networks = networksRes.data.data || []
      cache.composeProjects = composeRes.data.data || []
      cache.loaded = true
    } catch (error) {
      console.error('Failed to load host resources:', error)
    } finally {
      cache.loading = false
    }
  }

  // 刷新主机资源
  async function refreshHostResources(hostId: string) {
    const cache = resourcesCache.value.get(hostId)
    if (cache) {
      cache.loaded = false
      cache.loading = false
    }
    await loadHostResources(hostId)
  }

  // 清除所有缓存
  function clearAllCache() {
    resourcesCache.value.clear()
    expandedHosts.value.clear()
    expandedGroups.value.clear()
    selection.value = null
  }

  return {
    selection,
    expandedHosts,
    expandedGroups,
    resourcesCache,
    selectedHostId,
    setSelection,
    clearSelection,
    toggleHostExpand,
    toggleGroupExpand,
    isHostExpanded,
    isGroupExpanded,
    getHostCache,
    loadHostResources,
    refreshHostResources,
    clearAllCache,
  }
})
