// 统一 API client。VITE_API_BASE_URL 表示后端 origin，例如 http://localhost:8080。
// 业务 API 的 /api 前缀只在本模块集中拼接，页面和业务封装不直接写 /api。

const API_PREFIX = '/api'
const rawBaseURL = import.meta.env.VITE_API_BASE_URL ?? ''

export const apiBaseURL = normalizeBaseURL(rawBaseURL)

export type ApiMethod = 'GET' | 'POST' | 'PATCH' | 'PUT' | 'DELETE'

export interface ApiSuccessResponse<T> {
  success: true
  message: string
  data: T
}

export interface ApiErrorResponse {
  success: false
  message: string
  errorCode: string
}

export type ApiResponse<T> = ApiSuccessResponse<T> | ApiErrorResponse

export interface ApiRequestOptions<TBody = undefined> extends Omit<RequestInit, 'body' | 'headers' | 'method'> {
  method?: ApiMethod
  body?: TBody
  headers?: HeadersInit
  token?: string
  teacherToken?: string
  studentToken?: string
}

export class ApiRequestError extends Error {
  readonly status: number
  readonly errorCode?: string

  constructor(message: string, status: number, errorCode?: string) {
    super(message)
    this.name = 'ApiRequestError'
    this.status = status
    this.errorCode = errorCode
  }
}

export async function request<TResponse, TBody = undefined>(
  path: string,
  options: ApiRequestOptions<TBody> = {},
): Promise<TResponse> {
  const res = await fetch(buildApiURL(path), buildRequestInit(options))
  const envelope = await readResponseEnvelope<TResponse>(res)

  if (!res.ok || !envelope.success) {
    const message = envelope.message || `请求失败：${res.status} ${res.statusText}`
    const errorCode = envelope.success ? undefined : envelope.errorCode
    throw new ApiRequestError(message, res.status, errorCode)
  }

  return envelope.data
}

export async function download(path: string, options: ApiRequestOptions = {}) {
  const res = await fetch(buildApiURL(path), buildRequestInit(options))
  if (!res.ok) {
    const envelope = await readResponseEnvelope<unknown>(res)
    const message = envelope.message || `下载失败：${res.status} ${res.statusText}`
    const errorCode = envelope.success ? undefined : envelope.errorCode
    throw new ApiRequestError(message, res.status, errorCode)
  }

  return {
    blob: await res.blob(),
    fileName: readFileName(res.headers.get('Content-Disposition')),
    contentType: res.headers.get('Content-Type') ?? 'application/octet-stream',
  }
}

export function buildApiURL(path: string) {
  const apiPath = normalizeApiPath(path)
  if (!apiBaseURL) {
    return apiPath
  }
  return `${apiBaseURL}${apiPath}`
}

function buildRequestInit<TBody>(options: ApiRequestOptions<TBody>): RequestInit {
  const { body, headers: inputHeaders, token, teacherToken, studentToken, ...rest } = options
  const headers = new Headers(inputHeaders)
  const init: RequestInit = { ...rest, headers }

  if (token) {
    headers.set('Authorization', `Bearer ${token}`)
  }
  if (teacherToken) {
    headers.set('X-Teacher-Token', teacherToken)
  }
  if (studentToken) {
    headers.set('X-Student-Token', studentToken)
  }

  if (body !== undefined) {
    if (isBodyInit(body)) {
      init.body = body
    } else {
      if (!headers.has('Content-Type')) {
        headers.set('Content-Type', 'application/json')
      }
      init.body = JSON.stringify(body)
    }
  }

  return init
}

async function readResponseEnvelope<T>(res: Response): Promise<ApiResponse<T>> {
  try {
    return (await res.json()) as ApiResponse<T>
  } catch {
    return {
      success: false,
      message: res.statusText,
      errorCode: 'INVALID_RESPONSE',
    }
  }
}

function normalizeBaseURL(value: string) {
  const trimmed = value.trim().replace(/\/+$/, '')
  return trimmed.endsWith('/api') ? trimmed.slice(0, -4) : trimmed
}

function normalizeApiPath(path: string) {
  const trimmed = path.trim()
  const withSlash = trimmed.startsWith('/') ? trimmed : `/${trimmed}`
  return withSlash.startsWith(`${API_PREFIX}/`) ? withSlash : `${API_PREFIX}${withSlash}`
}

function isBodyInit(value: unknown): value is BodyInit {
  return (
    typeof value === 'string' ||
    value instanceof Blob ||
    value instanceof FormData ||
    value instanceof URLSearchParams ||
    value instanceof ArrayBuffer ||
    ArrayBuffer.isView(value) ||
    value instanceof ReadableStream
  )
}

function readFileName(contentDisposition: string | null) {
  if (!contentDisposition) {
    return ''
  }

  const match = contentDisposition.match(/filename="?([^";]+)"?/)
  return match?.[1] ?? ''
}

export const apiClient = {
  baseURL: apiBaseURL,
  request,
  download,
  get: <TResponse>(path: string, options?: ApiRequestOptions) =>
    request<TResponse>(path, { ...options, method: 'GET' }),
  post: <TResponse, TBody = undefined>(path: string, body?: TBody, options?: ApiRequestOptions<TBody>) =>
    request<TResponse, TBody>(path, { ...options, method: 'POST', body }),
  patch: <TResponse, TBody = undefined>(path: string, body?: TBody, options?: ApiRequestOptions<TBody>) =>
    request<TResponse, TBody>(path, { ...options, method: 'PATCH', body }),
}
