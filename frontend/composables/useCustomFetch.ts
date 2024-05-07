import type { NitroFetchOptions } from 'nitropack'
import { warn } from 'vue'
import { useCsrfStore } from '~/stores/useCsrfStore'
import type { LoginResponse } from '~/types/user'
import { mapping, type PathValues, API_PATH } from '~/types/customFetch'

const APIcallTimeout = 30 * 1000 // 30 seconds

const pathNames = Object.values(API_PATH)
type PathName = typeof pathNames[number]

function addQueryParams (path: string, query?: PathValues) {
  if (!query) {
    return path
  }
  const q = Object.entries(query).filter(([_, value]) => value !== undefined).map(([key, value]) => `${key}=${value}`).join('&')
  return `${path}?${q}`
}

export function useCustomFetch () {
  const headers = useRequestHeaders(['cookie'])
  const xForwardedFor = useRequestHeader('x-forwarded-for')
  const xRealIp = useRequestHeader('x-real-ip')
  const { csrfHeader, setCsrfHeader } = useCsrfStore()
  const { showError } = useBcToast()
  const { t: $t } = useI18n()

  async function fetch<T> (pathName: PathName, options: NitroFetchOptions<string & {}> = { }, pathValues?: PathValues, query?: PathValues, dontShowError = false): Promise<T> {
    const map = mapping[pathName]
    if (!map) {
      throw new Error(`path ${pathName} not found`)
    }

    if (options.signal === undefined) {
      options.signal = AbortSignal.timeout(APIcallTimeout)
    }

    if (map.mockFunction !== undefined && map.mock) {
      return map.mockFunction(options.body, pathValues, query) as T
    }

    const url = useRequestURL()
    const { public: { apiClient, legacyApiClient, apiKey, domain, logIp }, private: pConfig } = useRuntimeConfig()
    const path = addQueryParams(map.mock ? `${pathName}.json` : map.getPath?.(pathValues) || map.path, query)
    let baseURL = map.mock ? '../mock' : map.legacy ? legacyApiClient : apiClient

    if (process.server) {
      baseURL = map.mock ? `${domain || url.origin.replace('http:', 'https:')}/mock` : map.legacy ? pConfig?.legacyApiServer : pConfig?.apiServer
    }

    options.headers = new Headers({ ...options.headers, ...headers })
    if (apiKey) {
      options.headers.append('Authorization', `Bearer ${apiKey}`)
    }
    options.credentials = 'include'
    const method = options.method || map.method || 'GET'

    if (process.server && logIp === 'LOG') {
      warn(`x-forwarded-for: ${xForwardedFor}, x-real-ip: ${xRealIp} | ${method} -> ${pathName}, hasAuth: ${!!apiKey}`, headers)
    }

    // For non GET method's we need to set the csrf header for security
    if (method !== 'GET') {
      if (csrfHeader.value) {
        options.headers.append(csrfHeader.value[0], csrfHeader.value[1])
      } else {
        warn('missing csrf header!')
      }
    }

    if (pathName === API_PATH.LOGIN) {
      const res = await $fetch<LoginResponse>(path, {
        method,
        ...options,
        baseURL
      })
      return res as T
    }

    try {
      const res = await $fetch.raw<T>(path, { method, ...options, baseURL })
      if (method === 'GET') {
        // We get the csrf header from GET requests
        setCsrfHeader(res.headers)
      }
      return res._data as T
    } catch (e: any) {
      if (!dontShowError) {
        showError({ group: e.statusCode, summary: $t('error.ws_error'), detail: `${options.method}: ${baseURL}${path}` })
      }
      throw (e)
    }
  }
  return { fetch }
}
