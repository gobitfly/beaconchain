import { defineStore } from 'pinia'
import { warn } from 'vue'
import type { GetUserDashboardsResponse, UserDashboardsData } from '~/types/api/dashboard'
import type { VDBPostReturnData } from '~/types/api/validator_dashboard'
import type { ValidatorDashboardNetwork } from '~/types/dashboard'

export const useUserDashboardStore = defineStore('user_dashboards', () => {
  const { fetch } = useCustomFetch()
  const dashboards = ref<UserDashboardsData | undefined | null>()
  async function getDashboards () {
    const res = await fetch<GetUserDashboardsResponse>(API_PATH.USER_DASHBOARDS)
    dashboards.value = res.data
    return dashboards.value
  }

  async function createValidatorDashboard (name: string, network: ValidatorDashboardNetwork) {
    // TODO: implement real mapping of network id's once backend is ready for it
    warn(`we are currently ignoring the network ${network} and use 0 instead`)
    const res = await fetch<{data: VDBPostReturnData}>(API_PATH.DASHBOARD_CREATE_VALIDATOR, { body: { name, network: '0' } })
    if (res.data) {
      dashboards.value = {
        account_dashboards: dashboards.value?.account_dashboards || [],
        validator_dashboards: [
          ...(dashboards.value?.validator_dashboards || []),
          { id: res.data.id, name: res.data.name }
        ]
      }
      return res.data
    }
  }

  async function createAccountDashboard (name: string) {
    const res = await fetch<{data: VDBPostReturnData}>(API_PATH.DASHBOARD_CREATE_ACCOUNT, { body: { name } })
    if (res.data) {
      dashboards.value = {
        validator_dashboards: dashboards.value?.validator_dashboards || [],
        account_dashboards: [
          ...(dashboards.value?.account_dashboards || []),
          { id: res.data.id, name: res.data.name }
        ]
      }
      return res.data
    }
  }

  return { dashboards, getDashboards, createValidatorDashboard, createAccountDashboard }
})
