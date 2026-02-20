<template>
  <aside class="w-[280px] shrink-0 bg-base-300 flex flex-col h-full border-r border-base-content/10">
    <!-- Logo 区域 -->
    <div class="h-[60px] flex items-center justify-between px-4 border-b border-base-content/10">
      <h1 class="text-xl font-bold text-base-content">Rubick</h1>
      <button class="btn btn-sm btn-primary" @click="$emit('add-host')">
        <Icon icon="mdi:plus" class="text-lg" />
        添加主机
      </button>
    </div>

    <!-- 主机树 -->
    <div class="flex-1 overflow-y-auto p-2">
      <HostTree
        @select="$emit('select', $event)"
        @create-compose="$emit('create-compose', $event)"
      />
    </div>

    <!-- 底部版本信息 -->
    <div class="h-[40px] flex items-center justify-center border-t border-base-content/10 text-xs text-base-content/50">
      v1.0.0
    </div>
  </aside>
</template>

<script setup lang="ts">
import { Icon } from '@iconify/vue'
import HostTree from './HostTree.vue'

defineEmits<{
  (e: 'add-host'): void
  (e: 'select', selection: { hostId: string; type: string; id: string; name: string }): void
  (e: 'create-compose', hostId: string): void
}>()
</script>
