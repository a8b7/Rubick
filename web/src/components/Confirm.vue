<script setup lang="ts">
import { ref, reactive } from 'vue'
import { Icon } from '@iconify/vue'

interface ConfirmOption {
  key: string
  label: string
  default?: boolean
}

interface ConfirmOptions {
  title?: string
  message: string
  confirmText?: string
  cancelText?: string
  type?: 'danger' | 'warning' | 'info'
  options?: ConfirmOption[]
}

interface ConfirmResult {
  confirmed: boolean
  options?: Record<string, boolean>
}

const visible = ref(false)
const options = ref<ConfirmOptions>({
  title: '',
  message: '',
  confirmText: '确认',
  cancelText: '取消',
  type: 'info',
  options: []
})

const optionValues = reactive<Record<string, boolean>>({})

let resolvePromise: ((value: ConfirmResult) => void) | null = null

const confirm = (opts: ConfirmOptions): Promise<ConfirmResult> => {
  options.value = {
    title: opts.title || '确认',
    message: opts.message,
    confirmText: opts.confirmText || '确认',
    cancelText: opts.cancelText || '取消',
    type: opts.type || 'info',
    options: opts.options || []
  }

  // 初始化选项值
  Object.keys(optionValues).forEach(key => delete optionValues[key])
  opts.options?.forEach(opt => {
    optionValues[opt.key] = opt.default || false
  })

  visible.value = true

  return new Promise((resolve) => {
    resolvePromise = resolve
  })
}

const handleConfirm = () => {
  visible.value = false
  const result: ConfirmResult = { confirmed: true }
  if (options.value.options && options.value.options.length > 0) {
    result.options = { ...optionValues }
  }
  resolvePromise?.(result)
  resolvePromise = null
}

const handleCancel = () => {
  visible.value = false
  resolvePromise?.({ confirmed: false })
  resolvePromise = null
}

const getIconClass = () => {
  const classes: Record<string, string> = {
    danger: 'text-error',
    warning: 'text-warning',
    info: 'text-info'
  }
  return classes[options.value.type || 'info']
}

const getIcon = (): string => {
  const icons: Record<string, string> = {
    danger: 'mdi:alert-circle',
    warning: 'mdi:alert',
    info: 'mdi:information'
  }
  return icons[options.value.type || 'info'] || 'mdi:information'
}

const getButtonClass = () => {
  const classes: Record<string, string> = {
    danger: 'btn-error',
    warning: 'btn-warning',
    info: 'btn-primary'
  }
  return classes[options.value.type || 'info']
}

// Make confirm available globally
if (typeof window !== 'undefined') {
  (window as unknown as { confirm: typeof confirm }).confirm = confirm
}

defineExpose({ confirm })
</script>

<template>
  <dialog
    :class="['modal', { 'modal-open': visible }]"
  >
    <div class="modal-box">
      <div class="flex items-start gap-4">
        <Icon
          :icon="getIcon()"
          class="text-3xl shrink-0"
          :class="getIconClass()"
        />
        <div class="flex-1">
          <h3 class="font-bold text-lg mb-2">{{ options.title }}</h3>
          <p class="py-2 text-base-content/80">{{ options.message }}</p>

          <!-- 额外选项 -->
          <div v-if="options.options && options.options.length > 0" class="mt-4 space-y-2">
            <label
              v-for="opt in options.options"
              :key="opt.key"
              class="flex items-center gap-2 cursor-pointer"
            >
              <input
                type="checkbox"
                v-model="optionValues[opt.key]"
                class="checkbox checkbox-sm"
              />
              <span class="text-sm">{{ opt.label }}</span>
            </label>
          </div>
        </div>
      </div>
      <div class="modal-action">
        <button
          class="btn btn-ghost"
          @click="handleCancel"
        >
          {{ options.cancelText }}
        </button>
        <button
          class="btn"
          :class="getButtonClass()"
          @click="handleConfirm"
        >
          {{ options.confirmText }}
        </button>
      </div>
    </div>
    <div class="modal-backdrop" @click="handleCancel"></div>
  </dialog>
</template>
