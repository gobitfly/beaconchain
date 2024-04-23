import { defineStore } from 'pinia'
import { warn } from 'vue'
import type { GetUserDashboardsResponse, UserDashboardsData } from '~/types/api/dashboard'
import type { VDBPostReturnData } from '~/types/api/validator_dashboard'
import { type DashboardKey, type DashboardType, type CookieDashboard, type ValidatorDashboardNetwork, COOKIE_DASHBOARD_ID } from '~/types/dashboard'
import { COOKIE_KEY } from '~/types/cookie'

const userDashboardStore = defineStore('user_dashboards_store', () => {
  const data = ref<UserDashboardsData | undefined | null>()
  return { data }
})

export function useUserDashboardStore () {
  const { fetch } = useCustomFetch()
  const { t: $t } = useI18n()
  const { data } = storeToRefs(userDashboardStore())
  const { isLoggedIn } = useUserStore()
  const dashboardCookie = useCookie(COOKIE_KEY.USER_DASHBOARDS)

  const dashboards = computed(() => data.value)

  async function refreshDashboards () {
    if (isLoggedIn.value) {
      const res = await fetch<GetUserDashboardsResponse>(API_PATH.USER_DASHBOARDS)
      data.value = res.data

      // add fallback names for dashboards that have no names
      if (dashboards.value) {
        dashboards.value.account_dashboards?.forEach((d) => {
          if (d.name === '') {
            d.name = `${$t('dashboard.account_dashboard')} ${d.id}`
          }
        })
        dashboards.value.validator_dashboards?.forEach((d) => {
          if (d.name === '') {
            d.name = `${$t('dashboard.validator_dashboard')} ${d.id}`
          }
        })
      }
    } else if (dashboardCookie.value) {
      if (typeof dashboardCookie.value === 'object') {
        // it seems the browser sometimes auto converts the string into an object
        data.value = dashboardCookie.value as any as UserDashboardsData
      } else {
        data.value = JSON.parse(dashboardCookie.value)
      }
    }
    return dashboards.value
  }

  // Public dashboards are saved in a cookie (so that it's accessable during SSR)
  function saveToCookie () {
    dashboardCookie.value = JSON.stringify(dashboards.value)
  }

  async function createValidatorDashboard (name: string, network: ValidatorDashboardNetwork, dashboardKey?: string):Promise<CookieDashboard |undefined> {
    // TODO: implement real mapping of network id's once backend is ready for it (will not be part of first release)
    warn(`we are currently ignoring the network ${network}`)

    if (!isLoggedIn.value) {
      // Create local Validator dashboard
      const cd:CookieDashboard = { id: COOKIE_DASHBOARD_ID.VALIDATOR, name: '', hash: dashboardKey ?? '' }
      data.value = {
        account_dashboards: dashboards.value?.account_dashboards || [],
        validator_dashboards: [cd]
      }
      saveToCookie()
      return cd
    }
    // Create user specific Validator dashboard
    const res = await fetch<{data: VDBPostReturnData}>(API_PATH.DASHBOARD_CREATE_VALIDATOR, { body: { name, network: '0' } })
    if (res.data) {
      data.value = {
        account_dashboards: dashboards.value?.account_dashboards || [],
        validator_dashboards: [
          ...(dashboards.value?.validator_dashboards || []),
          { id: res.data.id, name: res.data.name }
        ]
      }
      return res.data
    }
  }

  async function createAccountDashboard (name: string, dashboardKey?: string):Promise<CookieDashboard |undefined> {
    if (!isLoggedIn.value) {
      // Create local account dashboard
      const cd:CookieDashboard = { id: COOKIE_DASHBOARD_ID.ACCOUNT, name: '', hash: dashboardKey ?? '' }
      data.value = {
        validator_dashboards: dashboards.value?.validator_dashboards || [],
        account_dashboards: [cd]
      }
      saveToCookie()
      return cd
    }
    // Create user specific account dashboard
    const res = await fetch<{data: VDBPostReturnData}>(API_PATH.DASHBOARD_CREATE_ACCOUNT, { body: { name } })
    if (res.data) {
      data.value = {
        validator_dashboards: dashboards.value?.validator_dashboards || [],
        account_dashboards: [
          ...(dashboards.value?.account_dashboards || []),
          { id: res.data.id, name: res.data.name }
        ]
      }
      return res.data
    }
  }

  // Update the hash (=hashed list of id's) of a specific local dashboard
  function updateHash (type: DashboardType, hash: string) {
    if (type === 'validator') {
      const cd:CookieDashboard = { id: COOKIE_DASHBOARD_ID.VALIDATOR, name: '', ...dashboards.value?.validator_dashboards?.[0], hash }
      data.value = {
        account_dashboards: dashboards.value?.account_dashboards || [],
        validator_dashboards: [cd]
      }
    } else {
      const cd:CookieDashboard = { id: COOKIE_DASHBOARD_ID.ACCOUNT, name: '', ...dashboards.value?.account_dashboards?.[0], hash }
      data.value = {
        validator_dashboards: dashboards.value?.validator_dashboards || [],
        account_dashboards: [cd]
      }
    }
    saveToCookie()
  }

  function getDashboardLabel (key: DashboardKey, type:DashboardType): string {
    const isValidatorDashboard = type === 'validator'
    const list = isValidatorDashboard ? dashboards.value?.validator_dashboards : dashboards.value?.account_dashboards
    const id = parseInt(key ?? '')
    if (!isNaN(id)) {
      const userDb = list?.find(db => db.id === id)
      if (userDb) {
        return userDb.name
      }

      // in production we should not get here, but with our public api key we can also view dashboards that are not part of our list
      return `${isValidatorDashboard ? $t('dashboard.validator_dashboard') : $t('dashboard.account_dashboard')} ${id}`
    }

    const cookieDb = (list as CookieDashboard[])?.find(db => db.hash === key)
    if (cookieDb || (isLoggedIn.value && !key)) {
      return isValidatorDashboard ? $t('dashboard.validator_dashboard') : $t('dashboard.account_dashboard')
    }

    return isValidatorDashboard ? $t('dashboard.public_validator_dashboard') : $t('dashboard.public_account_dashboard')
  }

  return { dashboards, refreshDashboards, createValidatorDashboard, createAccountDashboard, updateHash, getDashboardLabel }
}
