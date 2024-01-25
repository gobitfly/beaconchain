import type { NitroFetchOptions } from 'nitropack'
import type { LoginResponse } from '~/types/user'
// import { defu } from 'defu'

export enum API_PATH {
  LATEST_STATE= '/latestState',
  LOGIN= '/login',
  REFRESH_TOKEN= '/refreshToken'
}

const pathNames = Object.values(API_PATH)
type PathName = typeof pathNames[number]

const mapping:Record<string, {path:string, noAuth?:boolean, mock?: boolean}> = {
  [API_PATH.LATEST_STATE]: {
    path: '/latestState',
    mock: false
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

export async function useCustomFetch<T> (pathName: PathName, options: NitroFetchOptions<string & {}> = {}): Promise<T> {
  // the access token stuff is only a blue-print and needs to be refined once we have api calls to test against
  const refreshToken = useCookie('refreshToken')
  const accessToken = useCookie('accessToken')

  const map = mapping[pathName]
  if (!map) {
    throw new Error(`path ${pathName} not found`)
  }

  const url = useRequestURL()
  const { public: { apiClient }, private: pConfig } = useRuntimeConfig()
  const path = map.mock ? `${pathName}.json` : map.path
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
      const res = await useCustomFetch<{access_token:string}>(API_PATH.REFRESH_TOKEN, { method: 'POST', body: { refresh_token: refreshToken.value } })
      accessToken.value = res.access_token
    }

    if (accessToken.value) {
      options.headers = new Headers({})
      options.headers.append('Authorization', `Bearer ${accessToken.value}`)
    }
  }
  return await $fetch<T>(path, { ...options, baseURL })
}
