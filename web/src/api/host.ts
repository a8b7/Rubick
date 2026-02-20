import request, { type ApiResponse } from './request'

// 主机类型
export interface Host {
  id: string
  name: string
  type: 'local' | 'tcp' | 'ssh'
  host: string
  is_default: boolean
  is_active: boolean
  description: string
  created_at: string
  updated_at: string
  ssh_user?: string
  ssh_auth_type?: string
  ssh_port?: number
  docker_port?: number
  skip_tls_verify?: boolean
}

// 创建主机请求
export interface CreateHostRequest {
  name: string
  type: 'local' | 'tcp' | 'ssh'
  host?: string
  ssh_user?: string
  ssh_auth_type?: string
  ssh_private_key?: string
  ssh_password?: string
  ssh_port?: number
  docker_port?: number
  skip_tls_verify?: boolean
  description?: string
}

// 主机 API
export const hostApi = {
  // 获取主机列表
  list: () => request.get<ApiResponse<Host[]>>('/hosts'),

  // 获取主机详情
  get: (id: string) => request.get<ApiResponse<Host>>(`/hosts/${id}`),

  // 创建主机
  create: (data: CreateHostRequest) => request.post<ApiResponse<Host>>('/hosts', data),

  // 更新主机
  update: (id: string, data: Partial<CreateHostRequest>) =>
    request.put<ApiResponse<Host>>(`/hosts/${id}`, data),

  // 删除主机
  delete: (id: string) => request.delete<ApiResponse<void>>(`/hosts/${id}`),

  // 测试连接
  test: (id: string) => request.post<ApiResponse<{ success: boolean; message: string }>>(`/hosts/${id}/test`),
}
