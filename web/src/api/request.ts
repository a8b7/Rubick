import axios, { type AxiosInstance, type AxiosResponse } from 'axios'

// API 响应结构
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

// 分页响应
export interface PageResponse<T> {
  items: T[]
  total: number
  page: number
  page_size: number
}

// 创建 axios 实例
const request: AxiosInstance = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求拦截器
request.interceptors.request.use(
  (config) => {
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response: AxiosResponse<ApiResponse>) => {
    const res = response.data
    if (res.code !== 0 && res.code !== 200) {
      showToast(res.message || '请求失败', 'error')
      return Promise.reject(new Error(res.message || '请求失败'))
    }
    return response
  },
  (error) => {
    const message = error.response?.data?.message || error.message || '网络错误'
    showToast(message, 'error')
    return Promise.reject(error)
  }
)

// Toast 显示函数
function showToast(message: string, type: 'success' | 'error' | 'warning' | 'info') {
  const toast = (window as unknown as { toast?: { [key: string]: (msg: string) => void } }).toast
  if (toast && toast[type]) {
    toast[type](message)
  } else {
    console.log(`[${type.toUpperCase()}] ${message}`)
  }
}

export default request
