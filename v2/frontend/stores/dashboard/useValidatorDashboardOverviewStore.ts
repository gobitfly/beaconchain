import { defineStore } from 'pinia'
import type { InternalGetValidatorDashboardResponse } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'

export const useValidatorDashboardOverviewStore = defineStore('validator_overview', () => {
  const { fetch } = useCustomFetch()
  const key = ref<DashboardKey>()

  watch(key, (newKey) => {
    console.log('DIECE: New key found in useValidatorDashboardOverviewStore', newKey)
  })

  async function getOverview () {
    console.log('DIECE: Fetching for useValidatorDashboardOverviewStore <InternalGetValidatorDashboardResponse>, dashboardKey:', key.value)

    if (key === undefined) {
      return undefined
    }

    const res = await fetch<InternalGetValidatorDashboardResponse>(API_PATH.DASHBOARD_OVERVIEW, undefined, { dashboardKey: key.value! })
    return res.data
  }

  const { data, refresh } = useAsyncData('validator_dashboards', getOverview, { watch: [key] }) // TODO: Why doesn't this trigger getOverview()?
  return { key, validatorDashboardOverview: data, refreshValidatorDashboardOverview: refresh }
})
