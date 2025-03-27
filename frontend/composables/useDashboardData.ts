import type {
  GetValidatorDashboardExecutionLayerDepositsResponse,
  GetValidatorDashboardTotalExecutionDepositsResponse,
} from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import type { TableQueryParams } from '~/types/datatable'

export const useDashboardData = () => {
  const { fetch } = useCustomFetch()

  async function fetchELDeposits(
    dashboardKey: DashboardKey,
    query?: TableQueryParams,
  ) {
    const res
      = await fetch<GetValidatorDashboardExecutionLayerDepositsResponse>(
        'DASHBOARD_EL_DEPOSITS',
        undefined,
        { dashboardKey },
        query,
      )

    return res
  }
  async function fetchELDpositsTotalAmount(dashboardKey: DashboardKey) {
    const res
      = await fetch<GetValidatorDashboardTotalExecutionDepositsResponse>(
        'DASHBOARD_EL_DEPOSITS_TOTAL',
        undefined,
        { dashboardKey },
      )

    return res
  }

  return {
    fetchELDeposits,
    fetchELDpositsTotalAmount,
  }
}
