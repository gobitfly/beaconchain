import type {
  GetValidatorDashboardConsensusLayerConsolidationsResponse,
  GetValidatorDashboardConsensusLayerDepositsResponse,
  GetValidatorDashboardExecutionLayerConsolidationsResponse,
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
  async function fetchElConsolidations(
    dashboardKey: DashboardKey,
    query?: TableQueryParams,
  ) {
    const res
      = await fetch<GetValidatorDashboardExecutionLayerConsolidationsResponse>(
        'DASHBOARD_EL_CONSOLIDATIONS',
        undefined,
        { dashboardKey },
        query,
      )

    return res
  }
  async function fetchClConsolidations(
    dashboardKey: DashboardKey,
    query?: TableQueryParams,
  ) {
    const res
      = await fetch<GetValidatorDashboardConsensusLayerConsolidationsResponse>(
        'DASHBOARD_CL_CONSOLIDATIONS',
        undefined,
        { dashboardKey },
        query,
      )

    return res
  }

  return {
    fetchClConsolidations,
    fetchClDeposits,
    fetchClDpositsTotalAmount,
    fetchElConsolidations,
    fetchELDeposits,
    fetchELDpositsTotalAmount,
  }
}
