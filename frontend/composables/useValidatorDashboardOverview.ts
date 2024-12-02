import type { GetValidatorDashboardResponse } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import { API_PATH } from '~/types/customFetch'

export function useValidatorDashboardOverview() {
  const { fetch } = useCustomFetch()
  async function fetchOverviewData(key: DashboardKey) {
    if (!key) {
      throw new Error('No key provided when fetching overview')
    }
    const res = await fetch<GetValidatorDashboardResponse>(
      API_PATH.DASHBOARD_OVERVIEW,
      undefined,
      { dashboardKey: key },
    )
    return res.data
  }

  return {
    fetchOverviewData,
  }
}
