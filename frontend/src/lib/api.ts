import axios from 'axios'
import { ValidationError } from './types'

function getCsrfCookie(): string | undefined {
  const m = document.cookie.match(/(?:^|;\s*)csrf_token=([^;]*)/)
  return m?.[1]
}

export const api = axios.create({
  baseURL: '/api/v1',
  withCredentials: true,
})

api.interceptors.request.use((config) => {
  const method = config.method ?? 'get'
  if (method !== 'get') {
    const token = getCsrfCookie()
    if (token) config.headers['X-CSRF-Token'] = token
  }
  return config
})

let refreshPromise: Promise<unknown> | null = null

api.interceptors.response.use(undefined, async (error) => {
  const { response, config } = error
  if (response?.status === 401 && !config._retried && !config.url.includes('/auth/')) {
    config._retried = true
    refreshPromise ??= api.post('/auth/refresh').finally(() => (refreshPromise = null))
    try {
      await refreshPromise
      return api(config)
    } catch {
      return Promise.reject(error)
    }
  }
  if (response?.status === 422 && Array.isArray(response.data?.error?.details)) {
    throw new ValidationError(response.data.error.details)
  }
  return Promise.reject(error)
})

export function errMessage(e: unknown, fallback = 'Terjadi kesalahan. Coba lagi.'): string {
  const ax = e as { response?: { data?: { error?: { message?: string } }; status?: number } }
  if (ax.response?.data?.error?.message) return ax.response.data.error.message
  if (ax.response?.status === 429) return 'Terlalu banyak permintaan. Coba beberapa saat lagi.'
  return fallback
}
