import type { NitroFetchOptions } from 'nitropack'
import type { LoginResponse } from '~/types/user'
// import { defu } from 'defu'

export enum API_PATH {
  DASHBOARD_SUMMARY = '/dashboard/validatorSummary',
  DASHBOARD_SUMMARY_DETAILS = '/dashboard/validatorSummaryDetails',
  DASHBOARD_OVERVIEW = '/dashboard/overview',
  DASHBOARD_SLOTVIZ = '/dashboard/slotViz',
  LATEST_STATE = '/latestState',
  LOGIN = '/login',
  REFRESH_TOKEN = '/refreshToken'
}

const pathNames = Object.values(API_PATH)
type PathName = typeof pathNames[number]

export type PathValues = Record<string, string | number>

type MappingData = {
  path: string,
  getPath?: (values?: PathValues) => string,
  noAuth?: boolean,
  mock?: boolean
}

const mapping: Record<string, MappingData> = {
  [API_PATH.DASHBOARD_SUMMARY_DETAILS]: {
    path: '/validator-dashboards/{dashboard_id}/groups/{group_id}/summary',
    getPath: values => `/validator-dashboards/${values?.dashboardId}/groups/${values?.groupId}/summary`,
    mock: true
  },
  [API_PATH.DASHBOARD_SUMMARY]: {
    path: '/validator-dashboards/{dashboard_id}/summary',
    getPath: values => `/validator-dashboards/${values?.dashboardId}/summary`,
    mock: true
  },
  [API_PATH.DASHBOARD_OVERVIEW]: {
    path: '/validator-dashboards/{dashboard_id}',
    getPath: values => `/validator-dashboards/${values?.validatorId}`,
    mock: true
  },
  [API_PATH.DASHBOARD_SLOTVIZ]: {
    path: '/validator-slot-viz/{dashboard_id}',
    getPath: values => `/validator-slot-viz/${values?.validatorId}`,
    mock: true
  },
  [API_PATH.LATEST_STATE]: {
    path: '/latestState',
    mock: true
  },
  [API_PATH.LOGIN]: {
    path: '/login',
    noAuth: true,
    mock: true
  },
  [API_PATH.REFRESH_TOKEN]: {
    path: '/refreshToken',
    noAuth: true,
    mock: true
  }
}

export async function useCustomFetch<T> (pathName: PathName, options: NitroFetchOptions<string & {}> = {}, pathValues?: PathValues): Promise<T> {
  // the access token stuff is only a blue-print and needs to be refined once we have api calls to test against
  const refreshToken = useCookie('refreshToken')
  const accessToken = useCookie('accessToken')

  const map = mapping[pathName]
  if (!map) {
    throw new Error(`path ${pathName} not found`)
  }

  const url = useRequestURL()
  const { public: { apiClient }, private: pConfig } = useRuntimeConfig()
  const path = map.mock ? `${pathName}.json` : map.getPath?.(pathValues) || map.path
  let baseURL = map.mock ? './mock' : apiClient

  if (process.server) {
    baseURL = map.mock ? `${url.protocol}${url.host}/mock` : pConfig?.apiServer
  }

  if (pathName === API_PATH.LOGIN) {
    const res = await $fetch<LoginResponse>(path, { ...options, baseURL })
    refreshToken.value = res.refresh_token
    accessToken.value = res.access_token
    return res as T
  } else if (!map.noAuth) {
    if (!accessToken.value && refreshToken.value) {
      const res = await useCustomFetch<{ access_token: string }>(API_PATH.REFRESH_TOKEN, { method: 'POST', body: { refresh_token: refreshToken.value } })
      accessToken.value = res.access_token
    }

    if (accessToken.value) {
      options.headers = new Headers({})
      options.headers.append('Authorization', `Bearer ${accessToken.value}`)
    }
  }
  return await $fetch<T>(path, { ...options, baseURL })
}
