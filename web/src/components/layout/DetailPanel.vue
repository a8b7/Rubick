<template>
  <div class="detail-panel flex-1 flex flex-col overflow-hidden p-4">
    <!-- 容器详情 -->
    <template v-if="selection.type === 'container'">
      <ContainerDetail
        v-if="tab === 'detail'"
        :host-id="selection.hostId"
        :container-id="selection.id"
        class="overflow-auto"
        @refresh="$emit('refresh')"
      />
      <ContainerLogs
        v-else-if="tab === 'logs'"
        :host-id="selection.hostId"
        :container-id="selection.id"
        class="flex-1 min-h-0"
      />
      <ContainerTerminal
        v-else-if="tab === 'terminal'"
        :host-id="selection.hostId"
        :container-id="selection.id"
        class="flex-1 min-h-0"
      />
    </template>

    <!-- Compose 项目详情 -->
    <template v-else-if="selection.type === 'compose'">
      <ComposeDetail
        v-if="tab === 'detail'"
        :host-id="selection.hostId"
        :project-id="selection.id"
        class="overflow-auto"
        @refresh="$emit('refresh')"
      />
      <ComposeLogs
        v-else-if="tab === 'logs'"
        :host-id="selection.hostId"
        :project-id="selection.id"
        class="flex-1 min-h-0"
      />
    </template>

    <!-- 镜像详情 -->
    <ImageDetail
      v-else-if="selection.type === 'image'"
      :host-id="selection.hostId"
      :image-id="selection.id"
      class="overflow-auto"
    />

    <!-- 卷详情 -->
    <VolumeDetail
      v-else-if="selection.type === 'volume'"
      :host-id="selection.hostId"
      :volume-name="selection.id"
      class="overflow-auto"
    />

    <!-- 网络详情 -->
    <NetworkDetail
      v-else-if="selection.type === 'network'"
      :host-id="selection.hostId"
      :network-id="selection.id"
      class="overflow-auto"
    />

    <!-- 主机详情 -->
    <HostDetail
      v-else-if="selection.type === 'host'"
      :host-id="selection.hostId"
      class="overflow-auto"
    />
  </div>
</template>

<script setup lang="ts">
import type { Selection } from '@/stores'
import ContainerDetail from '@/components/details/ContainerDetail.vue'
import ContainerLogs from '@/components/details/ContainerLogs.vue'
import ContainerTerminal from '@/components/details/ContainerTerminal.vue'
import ComposeDetail from '@/components/details/ComposeDetail.vue'
import ComposeLogs from '@/components/details/ComposeLogs.vue'
import ImageDetail from '@/components/details/ImageDetail.vue'
import VolumeDetail from '@/components/details/VolumeDetail.vue'
import NetworkDetail from '@/components/details/NetworkDetail.vue'
import HostDetail from '@/components/details/HostDetail.vue'

defineProps<{
  tab: 'detail' | 'logs' | 'terminal'
  selection: Selection
}>()

defineEmits<{
  (e: 'refresh'): void
}>()
</script>
