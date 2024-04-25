import type { NitroFetchOptions } from 'nitropack'
import type { LoginResponse } from '~/types/user'
import { simulateAPIresponseForTheSearchBar } from '~/utils/mock'

const APIcallTimeout = 30 * 1000 // 30 seconds

export enum API_PATH {
  AD_CONFIGURATIONs = '/adConfigurations',
  USER_DASHBOARDS = '/user/dashboards',
  DASHBOARD_CREATE_ACCOUNT = '/dashboard/createAccount',
  DASHBOARD_CREATE_VALIDATOR = '/dashboard/createValidator',
  DASHBOARD_DELETE_ACCOUNT = '/dashboard/accountValidator',
  DASHBOARD_DELETE_VALIDATOR = '/dashboard/deleteValidator',
  DASHBOARD_VALIDATOR_MANAGEMENT = '/validator-dashboards/validators',
  DASHBOARD_VALIDATOR_GROUPS = '/validator-dashboards/groups',
  DASHBOARD_VALIDATOR_GROUP_MODIFY = '/validator-dashboards/group-modify',
  DASHBOARD_VALIDATOR_REWARDS_CHART = '/dashboard/validatorRewardsChart',
  DASHBOARD_VALIDATOR_BLOCKS = '/validator-dashboards/blocks',
  DASHBOARD_VALIDATOR_WITHDRAWALS = '/validator-dashboards/withdrawals',
  DASHBOARD_VALIDATOR_TOTAL_WITHDRAWALS = '/validator-dashboards/total-withdrawals',
  DASHBOARD_VALIDATOR_EPOCH_DUTY = '/validator-dashboards/epoch_duty',
  DASHBOARD_SUMMARY = '/dashboard/validatorSummary',
  DASHBOARD_SUMMARY_DETAILS = '/dashboard/validatorSummaryDetails',
  DASHBOARD_VALIDATOR_REWARDS = '/dashboard/validatorRewards',
  DASHBOARD_VALIDATOR_REWARDS_DETAILS = '/dashboard/validatorRewardsDetails',
  DASHBOARD_SUMMARY_CHART = '/dashboard/validatorSummaryChart',
  DASHBOARD_OVERVIEW = '/dashboard/overview',
  DASHBOARD_SLOTVIZ = '/dashboard/slotViz',
  LATEST_STATE = '/latestState',
  LOGIN = '/login',
  SEARCH = '/search'
}

const pathNames = Object.values(API_PATH)
type PathName = typeof pathNames[number]

export type PathValues = Record<string, string | number>

interface MockFunction {
  (body?: any, param?: PathValues, query?: PathValues) : any
}

type MappingData = {
  path: string,
  getPath?: (values?: PathValues) => string,
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
  [API_PATH.DASHBOARD_VALIDATOR_MANAGEMENT]: {
    path: 'validator-dashboards/{dashboard_id}/validators',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/validators`,
    mock: false
  },
  [API_PATH.DASHBOARD_VALIDATOR_BLOCKS]: {
    path: 'validator-dashboards/{dashboard_id}/blocks',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/blocks`,
    mock: false
  },
  [API_PATH.DASHBOARD_VALIDATOR_WITHDRAWALS]: {
    path: 'validator-dashboards/{dashboard_id}/withdrawals',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/withdrawals`,
    mock: false
  },
  [API_PATH.DASHBOARD_VALIDATOR_TOTAL_WITHDRAWALS]: {
    path: 'validator-dashboards/{dashboard_id}/total-withdrawals',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/total-withdrawals`,
    mock: false
  },
  [API_PATH.DASHBOARD_VALIDATOR_GROUPS]: {
    path: 'validator-dashboards/{dashboard_id}/groups',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/groups`,
    mock: false,
    method: 'POST'
  },
  [API_PATH.DASHBOARD_VALIDATOR_GROUP_MODIFY]: {
    path: 'validator-dashboards/{dashboard_id}/groups/{group_id}',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/groups/${values?.groupId}`,
    mock: false,
    method: 'PUT' // can be 'DELETE' = delete group or 'PUT' = modify group
  },
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
    mock: false,
    method: 'POST'
  },
  [API_PATH.DASHBOARD_DELETE_ACCOUNT]: {
    path: '/account-dashboards/{dashboardKey}',
    getPath: values => `/account-dashboards/${values?.dashboardKey}`,
    mock: true,
    method: 'DELETE'
  },
  [API_PATH.DASHBOARD_DELETE_VALIDATOR]: {
    path: '/validator-dashboards/{dashboardKey}',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}`,
    mock: false,
    method: 'DELETE'
  },
  [API_PATH.DASHBOARD_SUMMARY_DETAILS]: {
    path: '/validator-dashboards/{dashboardKey}/groups/{group_id}/summary',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/groups/${values?.groupId}/summary`,
    mock: false
  },
  [API_PATH.DASHBOARD_SUMMARY]: {
    path: '/validator-dashboards/{dashboardKey}/summary',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/summary`,
    mock: false
  },
  [API_PATH.DASHBOARD_VALIDATOR_REWARDS_CHART]: {
    path: '/validator-dashboards/{dashboardKey}/rewards-chart',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/rewards-chart`,
    mock: false
  },
  [API_PATH.DASHBOARD_VALIDATOR_REWARDS_DETAILS]: {
    path: '/validator-dashboards/{dashboardKey}/groups/{group_id}/rewards',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/groups/${values?.groupId}/rewards/${values?.epoch}`,
    mock: false
  },
  [API_PATH.DASHBOARD_VALIDATOR_REWARDS]: {
    path: '/validator-dashboards/{dashboardKey}/rewards',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/rewards`,
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
  [API_PATH.DASHBOARD_VALIDATOR_EPOCH_DUTY]: {
    path: '/validator-dashboards/{dashboard_id}/duties/{epoch}:',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/duties/${values?.epoch}`,
    mock: false
  },
  [API_PATH.LATEST_STATE]: {
    path: '/latestState',
    legacy: true,
    mock: false
  },
  [API_PATH.LOGIN]: {
    path: '/login',
    method: 'POST',
    mock: false
  },
  [API_PATH.SEARCH]: {
    path: '/search',
    method: 'POST',
    mock: true,
    mockFunction: simulateAPIresponseForTheSearchBar
  }
}

export function useCustomFetch () {
  const headers = useRequestHeaders(['cookie'])
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
    const { public: { apiClient, legacyApiClient, apiKey }, private: pConfig } = useRuntimeConfig()
    const path = addQueryParams(map.mock ? `${pathName}.json` : map.getPath?.(pathValues) || map.path, query)
    let baseURL = map.mock ? '../mock' : map.legacy ? legacyApiClient : apiClient

    if (process.server) {
      baseURL = map.mock ? `${url.origin.replace('http:', 'https:')}/mock` : map.legacy ? pConfig?.legacyApiServer : pConfig?.apiServer
    }

    options.headers = new Headers({ ...options.headers, ...headers })
    if (apiKey) {
      options.headers.append('Authorization', `Bearer ${apiKey}`)
    }
    options.credentials = 'include'
    const method = map.method || 'GET'
    if (pathName === API_PATH.LOGIN) {
      const res = await $fetch<LoginResponse>(path, {
        method,
        ...options,
        baseURL
      })
      return res as T
    }

    try {
      return await $fetch<T>(path, { method, ...options, baseURL })
    } catch (e: any) {
      if (!dontShowError) {
        showError({ group: e.statusCode, summary: $t('error.ws_error'), detail: `${options.method}: ${baseURL}${path}` })
      }
      throw (e)
    }
  }
  return { fetch }
}
