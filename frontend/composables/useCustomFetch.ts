import type { NitroFetchOptions } from 'nitropack'
import type { LoginResponse } from '~/types/user'

const APIcallTimeout = 30 * 1000 // 30 seconds

export enum API_PATH {
  AD_CONFIGURATIONs = '/adConfigurations',
  USER_DASHBOARDS = '/user/dashboards',
  DASHBOARD_CREATE_ACCOUNT = '/dashboard/createAccount',
  DASHBOARD_CREATE_VALIDATOR = '/dashboard/createValidator',
  DASHBOARD_SUMMARY = '/dashboard/validatorSummary',
  DASHBOARD_SUMMARY_DETAILS = '/dashboard/validatorSummaryDetails',
  DASHBOARD_SUMMARY_CHART = '/dashboard/validatorSummaryChart',
  DASHBOARD_OVERVIEW = '/dashboard/overview',
  DASHBOARD_SLOTVIZ = '/dashboard/slotViz',
  LATEST_STATE = '/latestState',
  LOGIN = '/login',
  REFRESH_TOKEN = '/refreshToken',
  SEARCH = '/search'
}

const pathNames = Object.values(API_PATH)
type PathName = typeof pathNames[number]

export type PathValues = Record<string, string | number>

interface MockFunction {
  (body?: Record<string, any>, param?: PathValues, query?: PathValues) : any
}

type MappingData = {
  path: string,
  getPath?: (values?: PathValues) => string,
  noAuth?: boolean,
  mock?: boolean,
  mockFunction?: MockFunction,
  legacy?: boolean
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE' // 'GET' will be used as default
}

function addQueryParams (path: string, query?: PathValues) {
  if (!query) {
    return path
  }
  const q = Object.entries(query).filter(([_, value]) => value !== undefined).map(([key, value]) => `${key}=${value}`).join('&')
  return `${path}?${q}`
}

const mapping: Record<string, MappingData> = {
  [API_PATH.AD_CONFIGURATIONs]: {
    path: '/ad-configurations?={keys}',
    getPath: values => `/ad-configurations?keys=${values?.keys}`,
    mock: true
  },
  [API_PATH.USER_DASHBOARDS]: {
    path: '/users/me/dashboards',
    mock: false
  },
  [API_PATH.DASHBOARD_CREATE_ACCOUNT]: {
    path: '/account-dashboards',
    mock: true,
    method: 'POST'
  },
  [API_PATH.DASHBOARD_CREATE_VALIDATOR]: {
    path: '/validator-dashboards',
    mock: true,
    method: 'POST'
  },
  [API_PATH.DASHBOARD_SUMMARY_DETAILS]: {
    path: '/validator-dashboards/{dashboardKey}/groups/{group_id}/summary',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/groups/${values?.groupId}/summary`,
    mock: false
  },
  [API_PATH.DASHBOARD_SUMMARY]: {
    path: '/validator-dashboards/{dashboardKey}/summary?',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/summary`,
    mock: false
  },
  [API_PATH.DASHBOARD_SUMMARY_CHART]: {
    path: '/validator-dashboards/{dashboardKey}/summary-chart?',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/summary-chart`,
    mock: false
  },
  [API_PATH.DASHBOARD_OVERVIEW]: {
    path: '/validator-dashboards/{dashboardKey}',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}`,
    mock: false
  },
  [API_PATH.DASHBOARD_SLOTVIZ]: {
    path: '/validator-dashboards/{dashboardKey}/slot-viz',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/slot-viz`,
    mock: false
  },
  [API_PATH.LATEST_STATE]: {
    path: '/latestState',
    legacy: true,
    mock: true
  },
  [API_PATH.LOGIN]: {
    path: '/login',
    method: 'POST',
    noAuth: true,
    mock: true
  },
  [API_PATH.REFRESH_TOKEN]: {
    path: '/refreshToken',
    method: 'POST',
    noAuth: true,
    mock: true
  }
}

export async function useCustomFetch<T> (pathName: PathName, options: NitroFetchOptions<string & {}> = { }, pathValues?: PathValues, query?: PathValues): Promise<T> {
  // TODO : the access token stuff is only a blue-print and needs to be refined once we have api calls to test against
  const refreshToken = useCookie('refreshToken')
  const accessToken = useCookie('accessToken')

  const map = mapping[pathName]
  if (!map) {
    throw new Error(`path ${pathName} not found`)
  }

  options.signal = AbortSignal.timeout(APIcallTimeout)

  if (map.mockFunction !== undefined && map.mock) {
    return map.mockFunction(options.body as Record<string, any>, pathValues, query) as T
  }

  const url = useRequestURL()
  const { public: { apiClient, legacyApiClient, xUserId }, private: pConfig } = useRuntimeConfig()
  const path = addQueryParams(map.mock ? `${pathName}.json` : map.getPath?.(pathValues) || map.path, query)
  let baseURL = map.mock ? '../mock' : map.legacy ? legacyApiClient : apiClient

  if (process.server) {
    baseURL = map.mock ? `${url.protocol}${url.host}/mock` : map.legacy ? pConfig?.legacyApiServer : pConfig?.apiServer
  }

  const method = map.method || 'GET'
  if (pathName === API_PATH.LOGIN) {
    const res = await $fetch<LoginResponse>(path, { method, ...options, baseURL })
    refreshToken.value = res.refresh_token
    accessToken.value = res.access_token
    return res as T
  } else if (!map.noAuth) {
    if (!accessToken.value && refreshToken.value) {
      const res = await useCustomFetch<{ access_token: string }>(API_PATH.REFRESH_TOKEN, { body: { refresh_token: refreshToken.value } })
      accessToken.value = res.access_token
    }

    if (accessToken.value) {
      options.headers = new Headers({})
      options.headers.append('Authorization', `Bearer ${accessToken.value}`)
    } else if (xUserId) {
      options.headers = new Headers({})
      options.headers.append('X-User-Id', xUserId)
    }
  }

  return await $fetch<T>(path, { method, ...options, baseURL })
}
