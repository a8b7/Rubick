<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { Icon } from '@iconify/vue'

export interface ToastMessage {
  id: number
  message: string
  type: 'success' | 'error' | 'warning' | 'info'
}

const messages = ref<ToastMessage[]>([])
let toastId = 0

const addToast = (message: string, type: ToastMessage['type'], duration = 3000) => {
  const id = ++toastId
  messages.value.push({ id, message, type })

  setTimeout(() => {
    removeToast(id)
  }, duration)
}

const removeToast = (id: number) => {
  const index = messages.value.findIndex(m => m.id === id)
  if (index > -1) {
    messages.value.splice(index, 1)
  }
}

const getAlertClass = (type: ToastMessage['type']) => {
  const classes: Record<string, string> = {
    success: 'alert-success',
    error: 'alert-error',
    warning: 'alert-warning',
    info: 'alert-info'
  }
  return classes[type] || 'alert-info'
}

const getIcon = (type: ToastMessage['type']) => {
  const icons: Record<string, string> = {
    success: 'mdi:check-circle',
    error: 'mdi:alert-circle',
    warning: 'mdi:alert',
    info: 'mdi:information'
  }
  return icons[type] || 'mdi:information'
}

// Export toast methods
const toast = {
  success: (message: string, duration?: number) => addToast(message, 'success', duration),
  error: (message: string, duration?: number) => addToast(message, 'error', duration),
  warning: (message: string, duration?: number) => addToast(message, 'warning', duration),
  info: (message: string, duration?: number) => addToast(message, 'info', duration)
}

onMounted(() => {
  // Make toast available globally when component is mounted
  ;(window as unknown as { toast: typeof toast }).toast = toast
})

onUnmounted(() => {
  // Clean up when component is unmounted
  delete (window as unknown as { toast?: typeof toast }).toast
})

defineExpose({ toast })
</script>

<template>
  <div class="toast toast-top toast-center z-50">
    <TransitionGroup name="toast">
      <div
        v-for="msg in messages"
        :key="msg.id"
        class="alert"
        :class="getAlertClass(msg.type)"
      >
        <Icon :icon="getIcon(msg.type)" class="text-xl" />
        <span>{{ msg.message }}</span>
      </div>
    </TransitionGroup>
  </div>
</template>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}

.toast-enter-from,
.toast-leave-to {
  opacity: 0;
  transform: translateY(-20px);
}
</style>
