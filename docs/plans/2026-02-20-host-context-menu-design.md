# 主机右键菜单功能设计

## 概述

为左侧主机树节点添加右键菜单，支持查看、编辑、测试连接和删除操作。

## 需求

- 右键点击主机节点时显示上下文菜单
- 菜单位置跟随鼠标位置
- 菜单项：查看详情、编辑、测试连接、删除主机

## 组件结构

```
web/src/components/
├── ContextMenu.vue        # 新建：通用右键菜单组件
├── HostDetailDialog.vue   # 新建：主机详情对话框（只读）
├── HostFormDialog.vue     # 已有：编辑主机表单（复用）
└── Confirm.vue            # 已有：删除确认（复用）

web/src/components/layout/
└── HostTreeNode.vue       # 修改：添加右键事件和菜单集成
```

## ContextMenu 组件设计

### Props

```typescript
interface Props {
  items: MenuItem[]  // 菜单项列表
  visible: boolean   // 是否显示
  x: number          // 鼠标 X 坐标
  y: number          // 鼠标 Y 坐标
}

interface MenuItem {
  label: string           // 显示文本
  icon?: string           // 图标（可选）
  disabled?: boolean      // 是否禁用
  danger?: boolean        // 危险样式（红色）
  divider?: boolean       // 是否为分隔线
  action?: () => void     // 点击回调
}
```

### 功能

- 根据 x/y 定位菜单
- 点击菜单项后自动关闭
- 点击菜单外部自动关闭
- 支持 Escape 键关闭
- 边界检测（超出视口时自动调整位置）

## 右键菜单项

| 菜单项 | 图标 | 操作 |
|--------|------|------|
| 查看详情 | `mdi:eye` | 打开 `HostDetailDialog`（只读） |
| 编辑 | `mdi:pencil` | 打开 `HostFormDialog`（编辑模式） |
| 测试连接 | `mdi:lan-connect` | 调用 `hostApi.test(id)`，显示结果 Toast |
| — | — | （分隔线） |
| 删除 | `mdi:delete` | 打开 `Confirm` 确认后调用 `hostApi.delete(id)` |

## 交互流程

```
用户右键点击主机节点
        ↓
阻止默认右键菜单 + 记录坐标
        ↓
显示 ContextMenu（跟随鼠标位置）
        ↓
用户点击菜单项
        ↓
执行对应操作 + 关闭菜单
```

### 编辑流程

打开 `HostFormDialog`，传入 `hostId`，保存后刷新主机列表

### 删除流程

打开 `Confirm` 对话框（danger 类型），确认后调用删除 API，成功后刷新列表

### 测试连接流程

调用 API，通过 Toast 显示成功/失败消息

## 文件变更清单

| 文件 | 操作 | 说明 |
|------|------|------|
| `web/src/components/ContextMenu.vue` | 新建 | 通用右键菜单组件 |
| `web/src/components/HostDetailDialog.vue` | 新建 | 主机详情对话框 |
| `web/src/components/layout/HostTreeNode.vue` | 修改 | 集成右键菜单 |
