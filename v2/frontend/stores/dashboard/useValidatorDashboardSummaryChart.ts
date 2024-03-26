import { defineStore } from 'pinia'
import { type ChartData } from '~/types/api/common'
import { type InternalGetValidatorDashboardSummaryChartResponse } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'

const validatorDashboardSummaryChartStore = defineStore('useValidatorDashboardSummaryChartStore', () => {
  const data = ref<ChartData<number> | undefined >()
  return { data }
})

export function useValidatorDashboardSummaryChartStore () {
  const { fetch } = useCustomFetch()
  const { data } = storeToRefs(validatorDashboardSummaryChartStore())

  const chartData = readonly(data)

  async function refreshChartData (dashboardKey: DashboardKey) {
    const response = await fetch<InternalGetValidatorDashboardSummaryChartResponse>(API_PATH.DASHBOARD_SUMMARY_CHART, undefined, { dashboardKey })
    data.value = response.data

    return data.value
  }

  return { chartData, refreshChartData }
}
