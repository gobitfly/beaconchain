import { defineStore } from 'pinia'
import { type ChartData } from '~/types/api/common'
import { type InternalGetValidatorDashboardSummaryChartResponse } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'

// this is a store (even though the data is only used by a single component) because the component should not reload the data everytime the component is created

const validatorDashboardSummaryChartStore = defineStore('useValidatorDashboardSummaryChartStore', () => {
  const data = ref<ChartData<number> | undefined >()
  return { data }
})

export function useValidatorDashboardSummaryChartStore () {
  const { fetch } = useCustomFetch()
  const { data } = storeToRefs(validatorDashboardSummaryChartStore())

  const chartData = computed(() => data.value)

  async function refreshChartData (dashboardKey: DashboardKey) {
    const res = await fetch<InternalGetValidatorDashboardSummaryChartResponse>(API_PATH.DASHBOARD_SUMMARY_CHART, undefined, { dashboardKey })
    data.value = res.data

    return chartData.value
  }

  return { chartData, refreshChartData }
}
