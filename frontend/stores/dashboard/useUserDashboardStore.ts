import { defineStore } from 'pinia'
import { warn } from 'vue'
import type { GetUserDashboardsResponse, UserDashboardsData } from '~/types/api/dashboard'
import type { VDBPostReturnData } from '~/types/api/validator_dashboard'
import type { DashboardType, ExtendedDashboard, ValidatorDashboardNetwork } from '~/types/dashboard'

const userDashboardStore = defineStore('user_dashboards_store', () => {
  const data = ref<UserDashboardsData | undefined | null>()
  return { data }
})

export function useUserDashboardStore () {
  const { fetch } = useCustomFetch()
  const { data } = storeToRefs(userDashboardStore())
  const { isLoggedIn } = useUserStore()
  const dashboardCookie = useCookie('user-dashboards')

  const dashboards = computed(() => data.value)

  async function refreshDashboards () {
    if (isLoggedIn.value) {
      const res = await fetch<GetUserDashboardsResponse>(API_PATH.USER_DASHBOARDS)
      data.value = res.data
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

  async function createValidatorDashboard (name: string, network: ValidatorDashboardNetwork, dashboardKey?: string):Promise<ExtendedDashboard |undefined> {
    // TODO: implement real mapping of network id's once backend is ready for it (will not be part of first release)
    warn(`we are currently ignoring the network ${network}`)

    if (!isLoggedIn.value) {
      // Create public Validator dashboard
      const db:ExtendedDashboard = { id: 0, name, hash: dashboardKey ?? '' }
      data.value = {
        account_dashboards: dashboards.value?.account_dashboards || [],
        validator_dashboards: [db]
      }
      saveToCookie()
      return db
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

  async function createAccountDashboard (name: string, dashboardKey?: string):Promise<ExtendedDashboard |undefined> {
    if (!isLoggedIn.value) {
      // Create public account dashboard
      const db:ExtendedDashboard = { id: 0, name, hash: dashboardKey ?? '' }
      data.value = {
        validator_dashboards: dashboards.value?.validator_dashboards || [],
        account_dashboards: [db]
      }
      saveToCookie()
      return db
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

  // Update the hash (=hashed list of id's) of a specific public dashboard
  function updateHash (type: DashboardType, hash: string) {
    if (type === 'validator') {
      const db:ExtendedDashboard = { id: 0, name: 'default', ...dashboards.value?.validator_dashboards?.[0], hash }
      data.value = {
        account_dashboards: dashboards.value?.account_dashboards || [],
        validator_dashboards: [db]
      }
    } else {
      const db:ExtendedDashboard = { id: 0, name: 'default', ...dashboards.value?.account_dashboards?.[0], hash }
      data.value = {
        validator_dashboards: dashboards.value?.validator_dashboards || [],
        account_dashboards: [db]
      }
    }
    saveToCookie()
  }

  return { dashboards, refreshDashboards, createValidatorDashboard, createAccountDashboard, updateHash }
}
