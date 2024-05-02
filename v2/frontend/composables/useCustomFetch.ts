import type { NitroFetchOptions } from 'nitropack'
import { warn } from 'vue'
import { useCsrfStore } from '~/stores/useCsrfStore'
import type { LoginResponse } from '~/types/user'
import { mapping, type PathValues } from '~/types/customFetch'

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
  DASHBOARD_EL_DEPOSITS = '/dashboard/elDeposits',
  DASHBOARD_EL_DEPOSITS_TOTAL = '/dashboard/elDepositsTotal',
  DASHBOARD_CL_DEPOSITS = '/dashboard/clDeposits',
  DASHBOARD_CL_DEPOSITS_TOTAL = '/dashboard/clDepositsTotal',
  DASHBOARD_OVERVIEW = '/dashboard/overview',
  DASHBOARD_SLOTVIZ = '/dashboard/slotViz',
  LATEST_STATE = '/latestState',
  LOGIN = '/login',
  SEARCH = '/search'
}

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
    const method = options.method || map.method || 'GET'

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
