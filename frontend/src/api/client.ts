// 统一 API client。base URL 必须来自环境变量，禁止硬编码后端地址。
// 注意：当前阶段不接入真实后端，本模块仅作封装预备，业务页面尚未调用。

const baseURL = import.meta.env.VITE_API_BASE_URL ?? ''

// 后端统一响应结构（对齐 backend 的 response 包，待联调时核对）。
export interface ApiResponse<T> {
  code: number
  message: string
  data: T
}

export async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const res = await fetch(`${baseURL}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers ?? {}),
    },
    ...options,
  })

  if (!res.ok) {
    throw new Error(`请求失败：${res.status} ${res.statusText}`)
  }

  const json = (await res.json()) as ApiResponse<T>
  return json.data
}

export const apiClient = {
  baseURL,
  request,
  get: <T>(path: string) => request<T>(path, { method: 'GET' }),
  post: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: 'POST', body: body ? JSON.stringify(body) : undefined }),
}
