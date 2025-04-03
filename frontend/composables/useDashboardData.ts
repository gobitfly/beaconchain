import type {
  GetValidatorDashboardConsensusLayerDepositsResponse,
  GetValidatorDashboardExecutionLayerDepositsResponse,
  GetValidatorDashboardTotalConsensusDepositsResponse,
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

  async function fetchClDeposits(
    dashboardKey: DashboardKey,
    query?: TableQueryParams,
  ) {
    const res
    = await fetch<GetValidatorDashboardConsensusLayerDepositsResponse>(
      'DASHBOARD_CL_DEPOSITS',
      undefined,
      { dashboardKey },
      query,
    )

    return res
  }

  async function fetchClDpositsTotalAmount(dashboardKey: DashboardKey) {
    const res
      = await fetch<GetValidatorDashboardTotalConsensusDepositsResponse>(
        'DASHBOARD_CL_DEPOSITS_TOTAL',
        undefined,
        { dashboardKey },
      )

    return res
  }

  return {
    fetchClDeposits,
    fetchClDpositsTotalAmount,
    fetchELDeposits,
    fetchELDpositsTotalAmount,
  }
}
