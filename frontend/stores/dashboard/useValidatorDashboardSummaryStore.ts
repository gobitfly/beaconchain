import { defineStore } from 'pinia'
import type { InternalGetValidatorDashboardSummaryResponse } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import type { TableQueryParams } from '~/types/datatable'
import { API_PATH } from '~/types/customFetch'
import type { SummaryTimeFrame } from '~/types/dashboard/summary'

const validatorDashboardSummaryStore = defineStore('validator_dashboard_sumary_store', () => {
  const data = ref < InternalGetValidatorDashboardSummaryResponse>()
  const query = ref < TableQueryParams>()

  return { data, query }
})

export function useValidatorDashboardSummaryStore () {
  const { fetch } = useCustomFetch()

  const { data, query: storedQuery } = storeToRefs(validatorDashboardSummaryStore())
  const isLoading = ref(false)

  const summary = computed(() => data.value)
  const query = computed(() => storedQuery.value)

  async function getSummary (dashboardKey: DashboardKey, timeFrame: SummaryTimeFrame, query?: TableQueryParams) {
    if (!dashboardKey) {
      data.value = undefined
      return undefined
    }
    isLoading.value = true
    storedQuery.value = query

    const res = await fetch<InternalGetValidatorDashboardSummaryResponse>(API_PATH.DASHBOARD_SUMMARY, { query: { timeFrame } }, { dashboardKey }, query)
    isLoading.value = false
    if (JSON.stringify(storedQuery.value) !== JSON.stringify(query)) {
      return // in case some query params change while loading
    }
    data.value = res
    return res
  }

  return { summary, query, isLoading, getSummary }
}
