import { defineStore } from 'pinia'
import { useCustomFetch } from '~/composables/useCustomFetch'
import type { InternalGetValidatorDashboardSummaryResponse } from '~/types/api/validator_dashboard'
import type { TableQueryParams } from '~/types/datatable'

export const useValidatorDashboardSummaryStore = defineStore('validator_dashboard_sumary_store', () => {
  const summaryMap = ref < Record<number, InternalGetValidatorDashboardSummaryResponse >>({})
  const queryMap = ref < Record<number, TableQueryParams | undefined >>({})

  async function getSummary (dashboardId: number, query?: TableQueryParams) {
    queryMap.value = { ...queryMap.value, [dashboardId]: query }

    const res = await useCustomFetch<InternalGetValidatorDashboardSummaryResponse>(API_PATH.DASHBOARD_SUMMARY, undefined, { dashboardId }, query)

    if (JSON.stringify(queryMap.value[dashboardId]) !== JSON.stringify(query)) {
      return // in case some query params change while loading
    }
    summaryMap.value = { ...summaryMap.value, [dashboardId]: res }
    return res
  }

  return { summaryMap, queryMap, getSummary }
})
