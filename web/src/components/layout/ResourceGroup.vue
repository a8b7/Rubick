<template>
  <div class="resource-group">
    <!-- 分组标题 -->
    <div
      :class="[
        'flex items-center gap-2 px-2 py-1 rounded cursor-pointer transition-colors',
        expanded ? 'bg-base-200/50' : 'hover:bg-base-200/30'
      ]"
      @click="$emit('toggle')"
    >
      <Icon
        :icon="expanded ? 'mdi:chevron-down' : 'mdi:chevron-right'"
        class="text-sm text-base-content/50"
      />
      <Icon :icon="getIcon()" class="text-base" :class="getIconColor()" />
      <span class="text-sm text-base-content/80">{{ getLabel() }}</span>
      <span class="text-xs text-base-content/50 ml-auto">{{ items.length }}</span>
      <!-- Compose 类型的添加按钮 -->
      <button
        v-if="type === 'compose'"
        class="btn btn-xs btn-ghost btn-circle"
        @click.stop="$emit('create', hostId)"
        title="创建项目"
      >
        <Icon icon="mdi:plus" class="text-base" />
      </button>
    </div>

    <!-- 资源列表 -->
    <div v-show="expanded" class="ml-4 mt-0.5 space-y-0.5">
      <!-- 容器类型：按 compose 项目分组 -->
      <template v-if="type === 'container'">
        <!-- Compose 项目分组 -->
        <template v-for="group in composeGroups" :key="group.name">
          <div
            :class="[
              'flex items-center gap-2 px-2 py-1 rounded cursor-pointer transition-colors',
              'hover:bg-base-200/50'
            ]"
            @click="toggleComposeGroup(group.name)"
          >
            <Icon
              :icon="expandedComposeGroups.has(group.name) ? 'mdi:chevron-down' : 'mdi:chevron-right'"
              class="text-sm text-base-content/40"
            />
            <Icon icon="mdi:file-document-outline" class="text-sm text-secondary" />
            <span class="text-sm text-base-content/70">{{ group.name }}</span>
            <span class="text-xs text-base-content/40 ml-auto">{{ group.containers.length }}</span>
          </div>
          <!-- 项目下的容器 -->
          <div v-show="expandedComposeGroups.has(group.name)" class="ml-4 space-y-0.5">
            <div
              v-for="item in group.containers"
              :key="(item as Container).id"
              :class="[
                'flex items-center gap-2 px-2 py-1 rounded cursor-pointer transition-colors',
                'hover:bg-primary/10 hover:text-primary'
              ]"
              @click="handleSelect(item)"
            >
              <span
                class="w-1.5 h-1.5 rounded-full shrink-0"
                :class="(item as Container).state === 'running' ? 'bg-success' : 'bg-error'"
              ></span>
              <span class="text-sm truncate flex-1">{{ (item as Container).name }}</span>
            </div>
          </div>
        </template>

        <!-- 独立容器（不属于 compose 项目） -->
        <template v-if="standaloneContainers.length > 0">
          <div
            :class="[
              'flex items-center gap-2 px-2 py-1 rounded cursor-pointer transition-colors',
              'hover:bg-base-200/50'
            ]"
            @click="toggleComposeGroup('__standalone__')"
          >
            <Icon
              :icon="expandedComposeGroups.has('__standalone__') ? 'mdi:chevron-down' : 'mdi:chevron-right'"
              class="text-sm text-base-content/40"
            />
            <Icon icon="mdi:cube-outline" class="text-sm text-base-content/50" />
            <span class="text-sm text-base-content/60">独立容器</span>
            <span class="text-xs text-base-content/40 ml-auto">{{ standaloneContainers.length }}</span>
          </div>
          <div v-show="expandedComposeGroups.has('__standalone__')" class="ml-4 space-y-0.5">
            <div
              v-for="item in standaloneContainers"
              :key="(item as Container).id"
              :class="[
                'flex items-center gap-2 px-2 py-1 rounded cursor-pointer transition-colors',
                'hover:bg-primary/10 hover:text-primary'
              ]"
              @click="handleSelect(item)"
            >
              <span
                class="w-1.5 h-1.5 rounded-full shrink-0"
                :class="(item as Container).state === 'running' ? 'bg-success' : 'bg-error'"
              ></span>
              <span class="text-sm truncate flex-1">{{ (item as Container).name }}</span>
            </div>
          </div>
        </template>

        <div v-if="items.length === 0" class="px-2 py-1 text-xs text-base-content/40">
          暂无容器
        </div>
      </template>

      <!-- 非容器类型：直接显示列表 -->
      <template v-else>
        <div
          v-for="item in items"
          :key="getItemKey(item)"
          :class="[
            'flex items-center gap-2 px-2 py-1 rounded cursor-pointer transition-colors',
            'hover:bg-primary/10 hover:text-primary'
          ]"
          @click="handleSelect(item)"
        >
          <Icon :icon="getItemIcon(item)" class="text-sm text-base-content/50 shrink-0" />
          <span class="text-sm truncate flex-1">{{ getItemName(item) }}</span>
        </div>

        <div v-if="items.length === 0" class="px-2 py-1 text-xs text-base-content/40">
          暂无资源
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue'
import { Icon } from '@iconify/vue'
import type { Container } from '@/api/container'
import type { Image } from '@/api/image'
import type { Volume } from '@/api/volume'
import type { Network } from '@/api/network'
import type { ComposeProject } from '@/api/compose'

interface ComposeGroup {
  name: string
  containers: Container[]
}

const props = defineProps<{
  hostId: string
  type: 'container' | 'image' | 'volume' | 'network' | 'compose'
  items: Container[] | Image[] | Volume[] | Network[] | ComposeProject[]
  expanded: boolean
}>()

const emit = defineEmits<{
  (e: 'toggle'): void
  (e: 'select', selection: { hostId: string; type: string; id: string; name: string }): void
  (e: 'create', hostId: string): void
}>()

// Compose 分组展开状态
const expandedComposeGroups = ref<Set<string>>(new Set())

// 按 compose 项目分组
const composeGroups = computed<ComposeGroup[]>(() => {
  if (props.type !== 'container') return []

  const containers = props.items as Container[]
  const groupMap = new Map<string, Container[]>()

  containers.forEach(container => {
    const projectName = container.labels?.['com.docker.compose.project']
    if (projectName) {
      if (!groupMap.has(projectName)) {
        groupMap.set(projectName, [])
      }
      groupMap.get(projectName)!.push(container)
    }
  })

  return Array.from(groupMap.entries())
    .map(([name, containers]) => ({ name, containers }))
    .sort((a, b) => a.name.localeCompare(b.name))
})

// 独立容器（不属于任何 compose 项目）
const standaloneContainers = computed<Container[]>(() => {
  if (props.type !== 'container') return []

  const containers = props.items as Container[]
  return containers.filter(c => !c.labels?.['com.docker.compose.project'])
})

function toggleComposeGroup(name: string) {
  if (expandedComposeGroups.value.has(name)) {
    expandedComposeGroups.value.delete(name)
  } else {
    expandedComposeGroups.value.add(name)
  }
}

function getIcon(): string {
  switch (props.type) {
    case 'container':
      return 'mdi:cube-outline'
    case 'image':
      return 'mdi:image-outline'
    case 'volume':
      return 'mdi:folder-outline'
    case 'network':
      return 'mdi:lan'
    case 'compose':
      return 'mdi:file-document-outline'
    default:
      return 'mdi:file-outline'
  }
}

function getIconColor(): string {
  switch (props.type) {
    case 'container':
      return 'text-info'
    case 'image':
      return 'text-warning'
    case 'volume':
      return 'text-success'
    case 'network':
      return 'text-primary'
    case 'compose':
      return 'text-secondary'
    default:
      return 'text-base-content/60'
  }
}

function getLabel(): string {
  switch (props.type) {
    case 'container':
      return '容器'
    case 'image':
      return '镜像'
    case 'volume':
      return '卷'
    case 'network':
      return '网络'
    case 'compose':
      return 'Compose 项目'
    default:
      return '资源'
  }
}

function getItemName(item: Container | Image | Volume | Network | ComposeProject): string {
  if (props.type === 'container') {
    return (item as Container).name
  }
  if (props.type === 'image') {
    const tags = (item as Image).repo_tags
    if (tags && tags.length > 0 && tags[0]) {
      return tags[0]
    }
    return (item as Image).id.substring(0, 12)
  }
  if (props.type === 'volume') {
    return (item as Volume).name
  }
  if (props.type === 'network') {
    return (item as Network).name
  }
  if (props.type === 'compose') {
    return (item as ComposeProject).name
  }
  return 'Unknown'
}

function getItemKey(item: Container | Image | Volume | Network | ComposeProject): string {
  if (props.type === 'volume') {
    return (item as Volume).name
  }
  if (props.type === 'container') {
    return (item as Container).id
  }
  if (props.type === 'image') {
    return (item as Image).id
  }
  if (props.type === 'network') {
    return (item as Network).id
  }
  if (props.type === 'compose') {
    return (item as ComposeProject).id
  }
  return 'unknown'
}

function getItemIcon(item: Container | Image | Volume | Network | ComposeProject): string {
  if (props.type === 'container') {
    return (item as Container).state === 'running' ? 'mdi:cube' : 'mdi:cube-outline'
  }
  if (props.type === 'compose') {
    return (item as ComposeProject).status === 'running' ? 'mdi:file-document' : 'mdi:file-document-outline'
  }
  return getIcon()
}

function handleSelect(item: Container | Image | Volume | Network | ComposeProject) {
  let id: string
  if (props.type === 'volume') {
    id = (item as Volume).name
  } else if (props.type === 'container') {
    id = (item as Container).id
  } else if (props.type === 'image') {
    id = (item as Image).id
  } else if (props.type === 'network') {
    id = (item as Network).id
  } else {
    id = (item as ComposeProject).id
  }

  emit('select', {
    hostId: props.hostId,
    type: props.type,
    id,
    name: getItemName(item),
  })
}
</script>
