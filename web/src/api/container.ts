import request, { type ApiResponse } from './request'

// 容器类型
export interface Container {
  id: string
  name: string
  image: string
  state: string
  status: string
  created: number
  ports: PortMapping[]
  labels: Record<string, string>
  mounts: MountInfo[]
  networks: NetworkInfo[]
}

// 挂载信息
export interface MountInfo {
  type: string
  source: string
  destination: string
  mode: string
  rw: boolean
}

// 网络信息
export interface NetworkInfo {
  name: string
  network_id: string
  ip_address: string
  mac_address: string
  gateway: string
}

// 端口映射
export interface PortMapping {
  ip: string
  private_port: number
  public_port: number
  type: string
}

// 创建容器请求
export interface CreateContainerRequest {
  name: string
  image: string
  host_id: string
  ports?: PortMapping[]
  env?: string[]
  volumes?: VolumeBinding[]
  command?: string[]
  labels?: Record<string, string>
}

// 卷绑定
export interface VolumeBinding {
  source: string
  target: string
  mode: string
}

// 容器 API
export const containerApi = {
  // 获取容器列表
  list: (hostId: string, all = true) =>
    request.get<ApiResponse<Container[]>>('/containers', {
      params: { host_id: hostId, all },
    }),

  // 获取容器详情
  get: (hostId: string, id: string) =>
    request.get<ApiResponse<Container>>(`/containers/${id}`, {
      params: { host_id: hostId },
    }),

  // 创建容器
  create: (data: CreateContainerRequest) =>
    request.post<ApiResponse<{ id: string; warnings: string[] }>>('/containers', data),

  // 启动容器
  start: (hostId: string, id: string) =>
    request.post<ApiResponse<void>>(`/containers/${id}/start`, null, {
      params: { host_id: hostId },
    }),

  // 停止容器
  stop: (hostId: string, id: string, timeout?: number) =>
    request.post<ApiResponse<void>>(`/containers/${id}/stop`, { timeout }, {
      params: { host_id: hostId },
    }),

  // 重启容器
  restart: (hostId: string, id: string, timeout?: number) =>
    request.post<ApiResponse<void>>(`/containers/${id}/restart`, { timeout }, {
      params: { host_id: hostId },
    }),

  // 删除容器
  remove: (hostId: string, id: string, force = false, removeVolumes = false, removeImage = false) =>
    request.delete<ApiResponse<void>>(`/containers/${id}`, {
      params: { host_id: hostId, force, remove_volumes: removeVolumes, remove_image: removeImage },
    }),

  // 获取容器日志
  logs: (hostId: string, id: string, params: { tail?: number; follow?: boolean }) =>
    request.get<string>(`/containers/${id}/logs`, {
      params: { host_id: hostId, ...params },
    }),

  // 获取容器统计
  stats: (hostId: string, id: string) =>
    request.get<ApiResponse<ContainerStats>>(`/containers/${id}/stats`, {
      params: { host_id: hostId },
    }),

  // 创建 exec 实例
  createExec: (hostId: string, id: string, data: ExecCreateRequest) =>
    request.post<ApiResponse<{ id: string }>>(`/containers/${id}/exec`, data, {
      params: { host_id: hostId },
    }),
}

// Exec 创建请求
export interface ExecCreateRequest {
  cmd?: string[]
  user?: string
  tty?: boolean
  stdin?: boolean
  stdout?: boolean
  stderr?: boolean
}

// 容器统计
export interface ContainerStats {
  cpu_percent: number
  memory_usage: number
  memory_limit: number
  memory_percent: number
  network_rx: number
  network_tx: number
  block_read: number
  block_write: number
}
