import { defineStore } from 'pinia'
import type { InternalGetValidatorDashboardSummaryResponse } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import type { TableQueryParams } from '~/types/datatable'

export const useValidatorDashboardSummaryStore = defineStore('validator_dashboard_sumary_store', () => {
  const { fetch } = useCustomFetch()
  const summaryMap = ref < Record<DashboardKey, InternalGetValidatorDashboardSummaryResponse >>({})
  const queryMap = ref < Record<DashboardKey, TableQueryParams | undefined >>({})

  async function getSummary (dashboardKey: DashboardKey, query?: TableQueryParams) {
    if (!dashboardKey) {
      return
    }
    queryMap.value = { ...queryMap.value, [dashboardKey]: query }

    const res = await fetch<InternalGetValidatorDashboardSummaryResponse>(API_PATH.DASHBOARD_SUMMARY, undefined, { dashboardKey }, query)

    if (JSON.stringify(queryMap.value[dashboardKey]) !== JSON.stringify(query)) {
      return // in case some query params change while loading
    }
    summaryMap.value = { ...summaryMap.value, [dashboardKey]: res }
    return res
  }

  return { summaryMap, queryMap, getSummary }
})
