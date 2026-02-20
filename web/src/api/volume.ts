import request, { type ApiResponse } from './request'

// 卷类型
export interface Volume {
  name: string
  driver: string
  mountpoint: string
  created_at: string
  labels: Record<string, string>
  scope: string
  options: Record<string, string>
  usage_data?: {
    size: number
    ref_count: number
  }
}

// 卷 API
export const volumeApi = {
  // 获取卷列表
  list: (hostId: string) =>
    request.get<ApiResponse<Volume[]>>('/volumes', {
      params: { host_id: hostId },
    }),

  // 获取卷详情
  get: (hostId: string, name: string) =>
    request.get<ApiResponse<Volume>>(`/volumes/${name}`, {
      params: { host_id: hostId },
    }),

  // 创建卷
  create: (hostId: string, data: { name: string; driver?: string; labels?: Record<string, string> }) =>
    request.post<ApiResponse<Volume>>('/volumes', {
      host_id: hostId,
      ...data,
    }),

  // 删除卷
  remove: (hostId: string, name: string, force = false) =>
    request.delete<ApiResponse<void>>(`/volumes/${name}`, {
      params: { host_id: hostId, force },
    }),
}
