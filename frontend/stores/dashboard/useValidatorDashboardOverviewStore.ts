import { defineStore } from 'pinia'
import type { VDBOverviewData, InternalGetValidatorDashboardResponse } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'

export const useValidatorDashboardOverviewStore = defineStore('validator_overview', () => {
  const { fetch } = useCustomFetch()
  const overview = ref<VDBOverviewData | undefined | null>()
  async function getOverview (dashboardKey: DashboardKey) {
    const res = await fetch<InternalGetValidatorDashboardResponse>(API_PATH.DASHBOARD_OVERVIEW, undefined, { dashboardKey })
    overview.value = res.data
    return overview.value
  }

  return { overview, getOverview }
})
