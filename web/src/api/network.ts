import request, { type ApiResponse } from './request'

// 网络类型
export interface Network {
  id: string
  name: string
  driver: string
  scope: string
  ipam: {
    driver: string
    config: Array<{
      subnet: string
      gateway?: string
    }>
  }
  created: string
  labels: Record<string, string>
  internal: boolean
  attachable: boolean
  ingress: boolean
  enable_ipv6: boolean
}

// 网络创建请求
export interface CreateNetworkRequest {
  host_id: string
  name: string
  driver?: string
  subnet?: string
  gateway?: string
  labels?: Record<string, string>
  internal?: boolean
  attachable?: boolean
}

// 网络 API
export const networkApi = {
  // 获取网络列表
  list: (hostId: string) =>
    request.get<ApiResponse<Network[]>>('/networks', {
      params: { host_id: hostId },
    }),

  // 获取网络详情
  get: (hostId: string, id: string) =>
    request.get<ApiResponse<Network>>(`/networks/${id}`, {
      params: { host_id: hostId },
    }),

  // 创建网络
  create: (data: CreateNetworkRequest) =>
    request.post<ApiResponse<Network>>('/networks', data),

  // 删除网络
  remove: (hostId: string, id: string) =>
    request.delete<ApiResponse<void>>(`/networks/${id}`, {
      params: { host_id: hostId },
    }),
}
