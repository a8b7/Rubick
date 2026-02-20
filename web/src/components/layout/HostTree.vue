<template>
  <div class="host-tree">
    <div v-if="hostStore.loading" class="flex justify-center py-8">
      <span class="loading loading-spinner loading-md"></span>
    </div>

    <div v-else-if="hostStore.hosts.length === 0" class="text-center py-8 text-base-content/50">
      <Icon icon="mdi:server-off" class="text-4xl mb-2" />
      <p>暂无主机</p>
      <p class="text-sm">点击上方按钮添加主机</p>
    </div>

    <div v-else class="space-y-1">
      <HostTreeNode
        v-for="host in hostStore.hosts"
        :key="host.id"
        :host="host"
        @select="$emit('select', $event)"
        @create-compose="$emit('create-compose', $event)"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { Icon } from '@iconify/vue'
import { useHostStore } from '@/stores'
import HostTreeNode from './HostTreeNode.vue'

defineEmits<{
  (e: 'select', selection: { hostId: string; type: string; id: string; name: string }): void
  (e: 'create-compose', hostId: string): void
}>()

const hostStore = useHostStore()
</script>
