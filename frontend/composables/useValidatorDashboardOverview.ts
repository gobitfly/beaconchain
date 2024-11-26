import type { GetValidatorDashboardResponse } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'

export function useValidatorDashboardOverview() {
  const { fetch } = useCustomFetch()
  async function fetchOverviewData(key: DashboardKey) {
    if (!key) {
      throw new Error('No key provided when fetching overview')
    }
    const res = await fetch<GetValidatorDashboardResponse>(
      'DASHBOARD_OVERVIEW',
      undefined,
      { dashboardKey: key },
    )
    return res.data
  }

  return {
    fetchOverviewData,
  }
}
