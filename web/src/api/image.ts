import request, { type ApiResponse } from './request'

// 镜像类型
export interface Image {
  id: string
  repo_tags: string[]
  size: number
  created: number
  labels: Record<string, string>
}

// 镜像搜索结果
export interface SearchResult {
  name: string
  description: string
  is_official: boolean
  stars: number
}

// 镜像 API
export const imageApi = {
  // 获取镜像列表
  list: (hostId: string) =>
    request.get<ApiResponse<Image[]>>('/images', {
      params: { host_id: hostId },
    }),

  // 获取镜像详情
  get: (hostId: string, id: string) =>
    request.get<ApiResponse<Image>>(`/images/${id}`, {
      params: { host_id: hostId },
    }),

  // 拉取镜像
  pull: (hostId: string, image: string) =>
    request.post<ApiResponse<{ status: string }>>('/images/pull', {
      host_id: hostId,
      image,
    }),

  // 搜索镜像
  search: (hostId: string, term: string, limit = 25) =>
    request.get<ApiResponse<SearchResult[]>>('/images/search', {
      params: { host_id: hostId, term, limit },
    }),

  // 删除镜像
  remove: (hostId: string, id: string, force = false) =>
    request.delete<ApiResponse<void>>(`/images/${id}`, {
      params: { host_id: hostId, force },
    }),

  // 标记镜像
  tag: (hostId: string, id: string, repo: string, tag?: string) =>
    request.post<ApiResponse<void>>(`/images/${id}/tag`, {
      host_id: hostId,
      repo,
      tag,
    }),
}
