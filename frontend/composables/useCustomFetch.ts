import type { NitroFetchOptions } from 'nitropack'
import { useCsrfStore } from '~/stores/useCsrfStore'
import type { LoginResponse } from '~/types/user'
import {
  API_PATH, mapping, type PathValues,
} from '~/types/customFetch'

const APIcallTimeout = 30 * 1000 // 30 seconds

const pathNames = Object.values(API_PATH)
type PathName = (typeof pathNames)[number]

export function useCustomFetch() {
  const headers = useRequestHeaders([ 'cookie' ])
  const xForwardedFor = useRequestHeader('x-forwarded-for')
  const xRealIp = useRequestHeader('x-real-ip')
  const {
    csrfHeader, setCsrfHeader,
  } = useCsrfStore()
  const { showError } = useBcToast()
  const { t: $t } = useTranslation()
  const { $bcLogger } = useNuxtApp()
  const uuid = inject<{ value: string }>('app-uuid')

  async function fetch<T>(
    pathName: PathName,
    // eslint-disable-next-line @typescript-eslint/ban-types
    options: NitroFetchOptions<{} & string> = {},
    pathValues?: PathValues,
    query?: PathValues,
    dontShowError = false,
  ): Promise<T> {
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
    const runtimeConfig = useRuntimeConfig()
    const showInDevelopment = Boolean(runtimeConfig.public.showInDevelopment)
    const {
      private: pConfig,
      public: {
        apiClient, apiKey, domain, legacyApiClient, logIp,
      },
    } = runtimeConfig
    const path = map.mock
      ? `${pathName}.json`
      : map.getPath?.(pathValues) || map.path
    let baseURL = map.mock
      ? '../mock'
      : map.legacy
        ? legacyApiClient
        : apiClient
    const ssrSecret = pConfig?.ssrSecret

    if (isServerSide) {
      baseURL = map.mock
        ? `${domain || url.origin.replace('http:', 'https:')}/mock`
        : map.legacy
          ? pConfig?.legacyApiServer
          : pConfig?.apiServer
    }

    options.headers = new Headers({
      ...options.headers,
      ...headers,
    })
    if (apiKey) {
      options.headers.append('Authorization', `Bearer ${apiKey}`)
    }

    if (isServerSide && ssrSecret) {
      options.headers.append('x-ssr-secret', ssrSecret)
    }

    options.query = {
      ...options.query,
      ...query,
    }
    options.credentials = 'include'
    const method = options.method || map.method || 'GET'

    if (isServerSide && logIp === 'LOG') {
      $bcLogger.warn(
        `${
          uuid?.value
        } | x-forwarded-for: ${xForwardedFor}, x-real-ip: ${xRealIp} | ${method} -> ${pathName}, hasAuth: ${!!apiKey}`,
        headers,
      )
    }

    // For non GET method's we need to set the csrf header for security
    if (method !== 'GET') {
      if (csrfHeader.value) {
        options.headers.append(csrfHeader.value[0], csrfHeader.value[1])
      }
      else {
        $bcLogger.warn(`${uuid?.value} | missing csrf header!`)
      }
    }

    if (pathName === API_PATH.LOGIN) {
      const res = await $fetch<LoginResponse>(path, {
        baseURL,
        method,
        ...options,
      })
      return res as T
    }

    try {
      const res = await $fetch.raw<T>(path, {
        baseURL,
        method,
        ...options,
      })
      if (method === 'GET') {
        // We get the csrf header from GET requests
        setCsrfHeader(res.headers)
      }
      return res._data as T
    }
    catch (e: any) {
      if (!dontShowError && showInDevelopment) {
        showError({
          detail: `${options.method}: ${baseURL}${path}`,
          group: e.statusCode,
          summary: $t('error.ws_error'),
        })
      }
      throw e
    }
  }

  return { fetch }
}
