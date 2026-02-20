<template>
  <div class="tab-switcher flex items-center gap-1 px-4 py-2 bg-base-100 border-b border-base-content/10">
    <template v-if="type === 'container'">
      <button
        :class="['tab-btn', { active: modelValue === 'detail' }]"
        @click="$emit('update:modelValue', 'detail')"
      >
        <Icon icon="mdi:information-outline" class="text-lg" />
        <span>详情</span>
      </button>
      <button
        :class="['tab-btn', { active: modelValue === 'logs' }]"
        @click="$emit('update:modelValue', 'logs')"
      >
        <Icon icon="mdi:text-box-outline" class="text-lg" />
        <span>日志</span>
      </button>
      <button
        :class="['tab-btn', { active: modelValue === 'terminal' }]"
        @click="$emit('update:modelValue', 'terminal')"
      >
        <Icon icon="mdi:console" class="text-lg" />
        <span>终端</span>
      </button>
    </template>

    <template v-else-if="type === 'compose'">
      <button
        :class="['tab-btn', { active: modelValue === 'detail' }]"
        @click="$emit('update:modelValue', 'detail')"
      >
        <Icon icon="mdi:information-outline" class="text-lg" />
        <span>详情</span>
      </button>
      <button
        :class="['tab-btn', { active: modelValue === 'logs' }]"
        @click="$emit('update:modelValue', 'logs')"
      >
        <Icon icon="mdi:text-box-outline" class="text-lg" />
        <span>日志</span>
      </button>
    </template>

    <template v-else-if="type === 'host'">
      <button
        :class="['tab-btn', { active: modelValue === 'detail' }]"
        @click="$emit('update:modelValue', 'detail')"
      >
        <Icon icon="mdi:information-outline" class="text-lg" />
        <span>详情</span>
      </button>
    </template>

    <template v-else>
      <button
        :class="['tab-btn', { active: modelValue === 'detail' }]"
        @click="$emit('update:modelValue', 'detail')"
      >
        <Icon icon="mdi:information-outline" class="text-lg" />
        <span>详情</span>
      </button>
    </template>
  </div>
</template>

<script setup lang="ts">
import { Icon } from '@iconify/vue'

defineProps<{
  modelValue: 'detail' | 'logs' | 'terminal'
  type: string
}>()

defineEmits<{
  (e: 'update:modelValue', value: 'detail' | 'logs' | 'terminal'): void
}>()
</script>

<style scoped>
.tab-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 14px;
  color: oklch(var(--bc) / 0.6);
  transition: all 0.2s;
}

.tab-btn:hover {
  background: oklch(var(--p) / 0.1);
  color: oklch(var(--p));
}

.tab-btn.active {
  background: oklch(var(--p) / 0.15);
  color: oklch(var(--p));
  font-weight: 500;
}
</style>
