# 主机右键菜单功能实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 为左侧主机树节点添加右键菜单，支持查看、编辑、测试连接和删除操作。

**Architecture:** 创建通用 ContextMenu 组件处理右键菜单逻辑（定位、显示/隐藏），创建 HostDetailDialog 显示只读主机详情，修改 HostTreeNode 集成右键菜单。

**Tech Stack:** Vue 3 + TypeScript + Tailwind CSS + DaisyUI

---

### Task 1: 创建 ContextMenu 通用组件

**Files:**
- Create: `web/src/components/ContextMenu.vue`

**Step 1: 创建 ContextMenu 组件**

```vue
<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue'
import { Icon } from '@iconify/vue'

export interface MenuItem {
  label: string
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
```

**Step 2: 验证组件创建**

Run: `ls -la web/src/components/ContextMenu.vue`
Expected: 文件存在

**Step 3: Commit**

```bash
git add web/src/components/ContextMenu.vue
git commit -m "feat: add ContextMenu component for right-click menus"
```

---

### Task 2: 创建 HostDetailDialog 组件

**Files:**
- Create: `web/src/components/HostDetailDialog.vue`

**Step 1: 创建主机详情对话框组件**

```vue
<script setup lang="ts">
import { computed } from 'vue'
import { Icon } from '@iconify/vue'
import type { Host } from '@/api'

const props = defineProps<{
  visible: boolean
  host: Host | null
}>()

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void
}>()

const getTypeLabel = (type: string): string => {
  const labels: Record<string, string> = {
    local: '本地',
    tcp: 'TCP',
    ssh: 'SSH'
  }
  return labels[type] || type
}

const getTypeIcon = (type: string): string => {
  const icons: Record<string, string> = {
    local: 'mdi:desktop-classic',
    tcp: 'mdi:server-network',
    ssh: 'mdi:server'
  }
  return icons[type] || 'mdi:server'
}

const formatDate = (dateStr: string): string => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

const handleClose = () => {
  emit('update:visible', false)
}

const infoItems = computed(() => {
  if (!props.host) return []

  const items = [
    { label: '名称', value: props.host.name },
    { label: '类型', value: getTypeLabel(props.host.type), icon: getTypeIcon(props.host.type) },
    { label: '状态', value: props.host.is_active ? '在线' : '离线', isStatus: true, isActive: props.host.is_active },
    { label: '默认主机', value: props.host.is_default ? '是' : '否' },
  ]

  if (props.host.type !== 'local') {
    items.push({ label: '主机地址', value: props.host.host || '-' })
    items.push({ label: 'Docker 端口', value: props.host.docker_port?.toString() || '-' })
  }

  if (props.host.type === 'ssh') {
    items.push({ label: 'SSH 用户', value: props.host.ssh_user || '-' })
    items.push({ label: 'SSH 端口', value: props.host.ssh_port?.toString() || '-' })
    items.push({ label: '认证方式', value: props.host.ssh_auth_type === 'password' ? '密码' : '私钥' })
  }

  items.push({ label: '描述', value: props.host.description || '-' })
  items.push({ label: '创建时间', value: formatDate(props.host.created_at) })
  items.push({ label: '更新时间', value: formatDate(props.host.updated_at) })

  return items
})
</script>

<template>
  <dialog :class="['modal', { 'modal-open': visible }]">
    <div class="modal-box w-full max-w-md">
      <div class="flex items-center justify-between mb-4">
        <h3 class="font-bold text-lg">主机详情</h3>
        <button class="btn btn-sm btn-ghost btn-circle" @click="handleClose">
          <Icon icon="mdi:close" class="text-lg" />
        </button>
      </div>

      <div v-if="host" class="space-y-3">
        <div
          v-for="(item, index) in infoItems"
          :key="index"
          class="flex items-start gap-3 py-2 border-b border-base-200 last:border-0"
        >
          <span class="text-base-content/60 w-24 shrink-0 text-sm">{{ item.label }}</span>
          <div class="flex-1 flex items-center gap-2">
            <Icon v-if="item.icon" :icon="item.icon" class="text-lg text-base-content/60" />
            <span
              v-if="item.isStatus"
              class="flex items-center gap-1.5"
              :class="item.isActive ? 'text-success' : 'text-error'"
            >
              <span class="w-2 h-2 rounded-full" :class="item.isActive ? 'bg-success' : 'bg-error'"></span>
              {{ item.value }}
            </span>
            <span v-else class="text-sm">{{ item.value }}</span>
          </div>
        </div>
      </div>

      <div class="modal-action">
        <button class="btn btn-primary" @click="handleClose">关闭</button>
      </div>
    </div>
    <form method="dialog" class="modal-backdrop">
      <button @click="handleClose">close</button>
    </form>
  </dialog>
</template>
```

**Step 2: 验证组件创建**

Run: `ls -la web/src/components/HostDetailDialog.vue`
Expected: 文件存在

**Step 3: Commit**

```bash
git add web/src/components/HostDetailDialog.vue
git commit -m "feat: add HostDetailDialog component for viewing host details"
```

---

### Task 3: 修改 HostTreeNode 集成右键菜单

**Files:**
- Modify: `web/src/components/layout/HostTreeNode.vue`

**Step 1: 更新 HostTreeNode 组件，添加右键菜单支持**

完整替换文件内容为：

```vue
<script setup lang="ts">
import { computed, ref } from 'vue'
import { Icon } from '@iconify/vue'
import { useResourceStore, useHostStore } from '@/stores'
import { hostApi, type Host } from '@/api'
import { showToast } from '@/utils/toast'
import ResourceGroup from './ResourceGroup.vue'
import ContextMenu, { type MenuItem } from '@/components/ContextMenu.vue'
import HostDetailDialog from '@/components/HostDetailDialog.vue'
import HostFormDialog from '@/components/HostFormDialog.vue'
import Confirm from '@/components/Confirm.vue'

const props = defineProps<{
  host: Host
}>()

defineEmits<{
  (e: 'select', selection: { hostId: string; type: string; id: string; name: string }): void
  (e: 'create-compose', hostId: string): void
}>()

const resourceStore = useResourceStore()
const hostStore = useHostStore()

// 右键菜单状态
const contextMenuVisible = ref(false)
const contextMenuPosition = ref({ x: 0, y: 0 })

// 对话框状态
const detailDialogVisible = ref(false)
const editDialogVisible = ref(false)
const confirmRef = ref<InstanceType<typeof Confirm> | null>(null)

const isExpanded = computed(() => resourceStore.isHostExpanded(props.host.id))
const cache = computed(() => resourceStore.getHostCache(props.host.id))

function getHostIcon(): string {
  switch (props.host.type) {
    case 'local':
      return 'mdi:desktop-classic'
    case 'tcp':
      return 'mdi:server-network'
    case 'ssh':
      return 'mdi:server'
    default:
      return 'mdi:server'
  }
}

function toggleExpand() {
  resourceStore.toggleHostExpand(props.host.id)
}

function isGroupExpanded(type: string): boolean {
  return resourceStore.isGroupExpanded(`${props.host.id}:${type}`)
}

function toggleGroup(type: string) {
  resourceStore.toggleGroupExpand(`${props.host.id}:${type}`)
}

// 右键菜单处理
function handleContextMenu(e: MouseEvent) {
  e.preventDefault()
  contextMenuPosition.value = { x: e.clientX, y: e.clientY }
  contextMenuVisible.value = true
}

const menuItems = computed<MenuItem[]>(() => [
  {
    label: '查看详情',
    icon: 'mdi:eye',
    action: () => {
      detailDialogVisible.value = true
    }
  },
  {
    label: '编辑',
    icon: 'mdi:pencil',
    action: () => {
      editDialogVisible.value = true
    }
  },
  {
    label: '测试连接',
    icon: 'mdi:lan-connect',
    action: handleTestConnection
  },
  {
    label: '设为默认',
    icon: 'mdi:star',
    disabled: props.host.is_default,
    action: handleSetDefault
  },
  { divider: true },
  {
    label: '删除',
    icon: 'mdi:delete',
    danger: true,
    action: handleDelete
  }
])

function closeContextMenu() {
  contextMenuVisible.value = false
}

// 测试连接
async function handleTestConnection() {
  try {
    const res = await hostApi.test(props.host.id)
    if (res.data.success) {
      showToast('连接成功', 'success')
    } else {
      showToast(res.data.message || '连接失败', 'error')
    }
  } catch {
    showToast('连接测试失败', 'error')
  }
}

// 设为默认
async function handleSetDefault() {
  try {
    await hostApi.update(props.host.id, { is_default: true } as never)
    showToast('已设为默认主机', 'success')
    hostStore.loadHosts()
  } catch {
    showToast('操作失败', 'error')
  }
}

// 删除主机
async function handleDelete() {
  if (!confirmRef.value) return

  const result = await confirmRef.value.confirm({
    title: '删除主机',
    message: `确定要删除主机「${props.host.name}」吗？此操作不可恢复。`,
    confirmText: '删除',
    cancelText: '取消',
    type: 'danger'
  })

  if (result.confirmed) {
    try {
      await hostApi.delete(props.host.id)
      showToast('删除成功', 'success')
      hostStore.loadHosts()
    } catch {
      showToast('删除失败', 'error')
    }
  }
}
</script>

<template>
  <div class="host-tree-node">
    <!-- 主机节点 -->
    <div
      :class="[
        'flex items-center gap-2 px-2 py-1.5 rounded-lg cursor-pointer transition-colors',
        isExpanded ? 'bg-base-200' : 'hover:bg-base-200/50'
      ]"
      @click="toggleExpand"
      @contextmenu="handleContextMenu"
    >
      <Icon
        :icon="isExpanded ? 'mdi:chevron-down' : 'mdi:chevron-right'"
        class="text-lg text-base-content/60"
      />
      <Icon
        :icon="getHostIcon()"
        :class="['text-lg', host.is_active ? 'text-success' : 'text-error']"
      />
      <span class="flex-1 truncate text-sm">{{ host.name }}</span>
      <span v-if="cache?.loading" class="loading loading-spinner loading-xs"></span>
      <span
        v-else-if="host.is_active"
        class="w-2 h-2 rounded-full bg-success"
        title="在线"
      ></span>
      <span
        v-else
        class="w-2 h-2 rounded-full bg-error"
        title="离线"
      ></span>
    </div>

    <!-- 资源分组 -->
    <div v-if="isExpanded && cache" class="ml-4 mt-1 space-y-0.5">
      <!-- 容器 -->
      <ResourceGroup
        :host-id="host.id"
        type="container"
        :items="cache.containers"
        :expanded="isGroupExpanded('container')"
        @toggle="toggleGroup('container')"
        @select="$emit('select', $event)"
      />

      <!-- 镜像 -->
      <ResourceGroup
        :host-id="host.id"
        type="image"
        :items="cache.images"
        :expanded="isGroupExpanded('image')"
        @toggle="toggleGroup('image')"
        @select="$emit('select', $event)"
      />

      <!-- 卷 -->
      <ResourceGroup
        :host-id="host.id"
        type="volume"
        :items="cache.volumes"
        :expanded="isGroupExpanded('volume')"
        @toggle="toggleGroup('volume')"
        @select="$emit('select', $event)"
      />

      <!-- 网络 -->
      <ResourceGroup
        :host-id="host.id"
        type="network"
        :items="cache.networks"
        :expanded="isGroupExpanded('network')"
        @toggle="toggleGroup('network')"
        @select="$emit('select', $event)"
      />

      <!-- Compose 项目 -->
      <ResourceGroup
        :host-id="host.id"
        type="compose"
        :items="cache.composeProjects"
        :expanded="isGroupExpanded('compose')"
        @toggle="toggleGroup('compose')"
        @select="$emit('select', $event)"
        @create="$emit('create-compose', $event)"
      />
    </div>

    <!-- 右键菜单 -->
    <ContextMenu
      :visible="contextMenuVisible"
      :x="contextMenuPosition.x"
      :y="contextMenuPosition.y"
      :items="menuItems"
      @close="closeContextMenu"
    />

    <!-- 查看详情对话框 -->
    <HostDetailDialog
      v-model:visible="detailDialogVisible"
      :host="host"
    />

    <!-- 编辑对话框 -->
    <HostFormDialog
      v-model:visible="editDialogVisible"
      :host="host"
      @saved="hostStore.loadHosts()"
    />

    <!-- 确认对话框 -->
    <Confirm ref="confirmRef" />
  </div>
</template>
```

**Step 2: 验证前端编译**

Run: `cd /Users/yuu/Developer/saide/Rubick/web && pnpm run build`
Expected: 编译成功，无错误

**Step 3: Commit**

```bash
git add web/src/components/layout/HostTreeNode.vue
git commit -m "feat: integrate context menu into HostTreeNode with view/edit/test/delete actions"
```

---

### Task 4: 手动验证功能

**Step 1: 启动开发服务器**

Run: `cd /Users/yuu/Developer/saide/Rubick/web && pnpm run dev`

**Step 2: 手动测试清单**

1. 在浏览器打开 http://localhost:3000
2. 右键点击左侧任意主机节点
3. 验证右键菜单出现在鼠标位置
4. 点击「查看详情」→ 验证对话框显示主机信息
5. 点击「编辑」→ 验证编辑表单打开且数据正确
6. 点击「测试连接」→ 验证 Toast 显示结果
7. 点击「设为默认」→ 验证状态变化（如果主机不是默认的话）
8. 点击「删除」→ 验证确认对话框出现，确认后主机被删除
9. 点击菜单外部 → 验证菜单关闭
10. 按 ESC 键 → 验证菜单关闭

**Step 3: 最终提交（如需修复）**

```bash
git add -A
git commit -m "fix: resolve any issues found during testing"
```

---

## Summary

| Task | Description | Files Changed |
|------|-------------|---------------|
| 1 | Create ContextMenu component | +1 file |
| 2 | Create HostDetailDialog component | +1 file |
| 3 | Integrate into HostTreeNode | 1 file modified |
| 4 | Manual testing | - |
