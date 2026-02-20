<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { Icon } from '@iconify/vue'

export interface MenuItem {
  label?: string
  icon?: string
  disabled?: boolean
  danger?: boolean
  divider?: boolean
  action?: () => void
}

const props = defineProps<{
  visible: boolean
  x: number
  y: number
  items: MenuItem[]
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const menuRef = ref<HTMLElement | null>(null)
const position = ref({ x: 0, y: 0 })

// 计算菜单位置，防止超出视口
const updatePosition = () => {
  if (!menuRef.value) return

  const menu = menuRef.value
  const menuWidth = menu.offsetWidth || 150
  const menuHeight = menu.offsetHeight || 200

  let x = props.x
  let y = props.y

  // 右边界检测
  if (x + menuWidth > window.innerWidth) {
    x = window.innerWidth - menuWidth - 8
  }

  // 下边界检测
  if (y + menuHeight > window.innerHeight) {
    y = window.innerHeight - menuHeight - 8
  }

  position.value = { x: Math.max(8, x), y: Math.max(8, y) }
}

// 点击外部关闭
const handleClickOutside = (e: MouseEvent) => {
  if (menuRef.value && !menuRef.value.contains(e.target as Node)) {
    emit('close')
  }
}

// ESC 键关闭
const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Escape') {
    emit('close')
  }
}

watch(() => props.visible, (val) => {
  if (val) {
    // 延迟更新位置，等待 DOM 渲染
    setTimeout(updatePosition, 0)
  }
})

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
  document.addEventListener('keydown', handleKeydown)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
  document.removeEventListener('keydown', handleKeydown)
})

const handleItemClick = (item: MenuItem) => {
  if (item.disabled || item.divider) return
  item.action?.()
  emit('close')
}
</script>

<template>
  <Teleport to="body">
    <Transition name="context-menu">
      <div
        v-if="visible"
        ref="menuRef"
        class="fixed z-50 min-w-36 py-1 bg-base-100 rounded-lg shadow-lg border border-base-300"
        :style="{ left: `${position.x}px`, top: `${position.y}px` }"
      >
        <template v-for="(item, index) in items" :key="index">
          <div v-if="item.divider" class="divider my-1"></div>
          <button
            v-else
            class="w-full flex items-center gap-2 px-3 py-2 text-left text-sm hover:bg-base-200 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            :class="item.danger ? 'text-error hover:bg-error/10' : 'text-base-content'"
            :disabled="item.disabled"
            @click.stop="handleItemClick(item)"
          >
            <Icon v-if="item.icon" :icon="item.icon" class="text-lg" />
            <span>{{ item.label }}</span>
          </button>
        </template>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.context-menu-enter-active,
.context-menu-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}

.context-menu-enter-from,
.context-menu-leave-to {
  opacity: 0;
  transform: scale(0.95);
}
</style>
