import type { NitroFetchOptions } from 'nitropack'
import type {
  GetValidatorDashboardBlocksResponse,
  GetValidatorDashboardConsensusLayerDepositsResponse,
  GetValidatorDashboardExecutionLayerDepositsResponse,
  GetValidatorDashboardResponse,
  GetValidatorDashboardRewardsResponse,
  GetValidatorDashboardSummaryResponse,
  GetValidatorDashboardTotalConsensusDepositsResponse,
  GetValidatorDashboardTotalExecutionDepositsResponse,
  GetValidatorDashboardTotalWithdrawalsResponse,
  GetValidatorDashboardWithdrawalsResponse,
} from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import type { TableQueryParams } from '~/types/datatable'
import type { GetValidatorDashboardSlotVizResponse } from '~/types/api/slot_viz'
import type { SummaryTimeFrame } from '~/types/dashboard/summary'
import type { ApiPagingResponse } from '~/types/api/common'
import type { API_PATH } from '~/types/customFetch'

export const useValidatorDashboard = () => {
  const { fetch } = useCustomFetch()

  const fetchOverview = async (key: DashboardKey) => {
    if (!key) {
      return undefined
    }
    const res = await fetch<GetValidatorDashboardResponse>(
      'DASHBOARD_OVERVIEW',
      undefined,
      { dashboardKey: key },
    )
    return res.data
  }

  const fetchSlotViz = async (dashboardKey: DashboardKey, groupIds?: number[]) => {
    const query = groupIds?.length ? { group_ids: groupIds.join(',') } : undefined

    const res = await fetch<GetValidatorDashboardSlotVizResponse>(
      'DASHBOARD_SLOTVIZ',
      {
        headers: {},
        query,
      },
      { dashboardKey: dashboardKey || 'MQ' }, // If guest dashboard has no validators yet (= empty dashboardKey), load small guest dashboard with 1 validator (MQ)
    )
    const data = res.data
    if (!dashboardKey) {
      // remove validator info for empty dashboard
      data.forEach((epoch) => {
        epoch.slots?.forEach((slot) => {
          Object.assign(slot, {
            attestations: undefined, proposal: undefined, slashing: undefined, sync: undefined,
          })
        })
      })
    }

    return data
  }

  // Generic func for fetching table data as an ApiPagingResponse
  type ExtractedRowType<T> = T extends ApiPagingResponse<infer E> ? E : never
  const fetchTable = async <ResponseType extends ApiPagingResponse<any>>(
    dashboardKey: DashboardKey,
    apiPath: API_PATH,
    options: NitroFetchOptions<string & {}> = {},
    query?: TableQueryParams,
  ) => {
    if (!dashboardKey) {
      return undefined
    }
    const res = await fetch<ResponseType>(
      apiPath,
      options,
      { dashboardKey: dashboardKey },
      query,
    )
    return res as ApiPagingResponse<ExtractedRowType<ResponseType>>
  }

  const fetchSummary = async (
    dashboardKey: DashboardKey,
    timeFrame: SummaryTimeFrame,
    query?: TableQueryParams,
  ) => {
    return fetchTable<GetValidatorDashboardSummaryResponse>(
      dashboardKey,
      'DASHBOARD_SUMMARY',
      { query: { period: timeFrame } },
      query,
    )
  }

  const fetchRewards = async (
    dashboardKey: DashboardKey,
    query?: TableQueryParams,
  ) => {
    return fetchTable<GetValidatorDashboardRewardsResponse>(
      dashboardKey,
      'DASHBOARD_VALIDATOR_REWARDS',
      undefined,
      query,
    )
  }

  const fetchBlocks = async (
    dashboardKey: DashboardKey,
    query?: TableQueryParams,
  ) => {
    return fetchTable<GetValidatorDashboardBlocksResponse>(
      dashboardKey,
      'DASHBOARD_VALIDATOR_BLOCKS',
      undefined,
      query,
    )
  }

  const fetchClDeposits = async (
    dashboardKey: DashboardKey,
    query?: TableQueryParams,
  ) => {
    return fetchTable<GetValidatorDashboardConsensusLayerDepositsResponse>(
      dashboardKey,
      'DASHBOARD_CL_DEPOSITS',
      undefined,
      query,
    )
  }

  const fetchTotalClDeposits = async (dashboardKey: DashboardKey) => {
    if (!dashboardKey) {
      return undefined
    }
    const res = await fetch<GetValidatorDashboardTotalConsensusDepositsResponse>(
      'DASHBOARD_CL_DEPOSITS_TOTAL',
      undefined,
      { dashboardKey: dashboardKey },
    )
    return res.data
  }

  const fetchElDeposits = async (
    dashboardKey: DashboardKey,
    query?: TableQueryParams,
  ) => {
    return fetchTable<GetValidatorDashboardExecutionLayerDepositsResponse>(
      dashboardKey,
      'DASHBOARD_EL_DEPOSITS',
      undefined,
      query,
    )
  }

  const fetchTotalElDeposits = async (dashboardKey: DashboardKey) => {
    if (!dashboardKey) {
      return undefined
    }
    const res = await fetch<GetValidatorDashboardTotalExecutionDepositsResponse>(
      'DASHBOARD_EL_DEPOSITS_TOTAL',
      undefined,
      { dashboardKey: dashboardKey },
    )
    return res.data
  }

  const fetchWithdrawals = async (
    dashboardKey: DashboardKey,
    query?: TableQueryParams,
  ) => {
    return fetchTable<GetValidatorDashboardWithdrawalsResponse>(
      dashboardKey,
      'DASHBOARD_VALIDATOR_WITHDRAWALS',
      undefined,
      query,
    )
  }

  const fetchTotalWithdrawals = async (dashboardKey: DashboardKey) => {
    if (!dashboardKey) {
      return undefined
    }
    const res = await fetch<GetValidatorDashboardTotalWithdrawalsResponse>(
      'DASHBOARD_VALIDATOR_TOTAL_WITHDRAWALS',
      undefined,
      { dashboardKey: dashboardKey },
    )
    return res.data
  }

  return {
    fetchBlocks,
    fetchClDeposits,
    fetchElDeposits,
    fetchOverview,
    fetchRewards,
    fetchSlotViz,
    fetchSummary,
    fetchTotalClDeposits,
    fetchTotalElDeposits,
    fetchTotalWithdrawals,
    fetchWithdrawals,
  }
}
