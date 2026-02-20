import request, { type ApiResponse } from './request'

// Compose 项目类型
export interface ComposeProject {
  id: string
  name: string
  host_id: string
  source_type: 'content' | 'directory'
  content: string
  work_dir: string
  compose_file: string
  env_file: string
  status: string
  created_at: string
  updated_at: string
  host?: {
    id: string
    name: string
    type: string
  }
}

// 服务状态
export interface ServiceStatus {
  name: string
  command: string
  state: string
  status: string
  health: string
  exit_code: number
  publishers: PortPublisher[]
}

// 端口发布
export interface PortPublisher {
  url: string
  target_port: number
  published_port: number
  protocol: string
}

// 文件信息
export interface FileInfo {
  name: string
  path: string
  is_dir: boolean
  size: number
}

// 创建项目请求
export interface CreateProjectRequest {
  name: string
  host_id: string
  source_type?: 'content' | 'directory'
  content?: string
  work_dir?: string
  compose_file?: string
  env_file?: string
}

// Up 选项
export interface UpOptions {
  build?: boolean
  detach?: boolean
  remove_orphans?: boolean
  timeout?: number
  services?: string[]
}

// Down 选项
export interface DownOptions {
  remove_images?: string
  remove_volumes?: boolean
  remove_orphans?: boolean
  timeout?: number
}

// Compose API
export const composeApi = {
  // 获取项目列表
  list: (hostId?: string) =>
    request.get<ApiResponse<ComposeProject[]>>('/compose/projects', {
      params: { host_id: hostId },
    }),

  // 获取项目详情
  get: (id: string) =>
    request.get<ApiResponse<ComposeProject>>(`/compose/projects/${id}`),

  // 创建项目
  create: (data: CreateProjectRequest) =>
    request.post<ApiResponse<ComposeProject>>('/compose/projects', data),

  // 更新项目
  update: (id: string, data: Partial<CreateProjectRequest>) =>
    request.put<ApiResponse<ComposeProject>>(`/compose/projects/${id}`, data),

  // 删除项目
  delete: (id: string) =>
    request.delete<ApiResponse<void>>(`/compose/projects/${id}`),

  // 启动项目
  up: (id: string, options: UpOptions = {}) =>
    request.post<ApiResponse<void>>(`/compose/projects/${id}/up`, options),

  // 停止并删除项目
  down: (id: string, options: DownOptions = {}) =>
    request.post<ApiResponse<{ output: string }>>(`/compose/projects/${id}/down`, options),

  // 启动服务
  start: (id: string, services: string[] = []) =>
    request.post<ApiResponse<{ output: string }>>(`/compose/projects/${id}/start`, { services }),

  // 停止服务
  stop: (id: string, timeout?: number, services: string[] = []) =>
    request.post<ApiResponse<{ output: string }>>(`/compose/projects/${id}/stop`, { timeout, services }),

  // 重启服务
  restart: (id: string, timeout?: number, services: string[] = []) =>
    request.post<ApiResponse<{ output: string }>>(`/compose/projects/${id}/restart`, { timeout, services }),

  // 获取日志
  logs: (id: string, params: { tail?: string; follow?: boolean; services?: string[] }) =>
    request.get<string>(`/compose/projects/${id}/logs`, {
      params,
    }),

  // 获取服务状态
  ps: (id: string) =>
    request.get<ApiResponse<ServiceStatus[]>>(`/compose/projects/${id}/ps`),

  // 浏览目录
  browseDir: (hostId: string, path: string = '/') =>
    request.get<ApiResponse<{ path: string; files: FileInfo[] }>>('/compose/browse', {
      params: { host_id: hostId, path },
    }),

  // 扫描 compose 文件
  scanComposeFiles: (hostId: string, path: string) =>
    request.get<ApiResponse<{ path: string; compose_files: string[]; env_files: string[] }>>('/compose/scan', {
      params: { host_id: hostId, path },
    }),

  // 上传目录
  uploadDirectory: (hostId: string, targetPath: string, files: FileList) => {
    const formData = new FormData()
    formData.append('host_id', hostId)
    formData.append('target_path', targetPath)
    for (let i = 0; i < files.length; i++) {
      const file = files[i]
      if (file) {
        formData.append('files', file)
      }
    }
    return request.post<ApiResponse<{ path: string; uploaded_files: string[]; count: number }>>('/compose/upload', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
  },
}
