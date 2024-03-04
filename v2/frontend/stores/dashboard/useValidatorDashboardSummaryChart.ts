import { defineStore } from 'pinia'
import { useCustomFetch } from '~/composables/useCustomFetch'
import { type ChartData } from '~/types/api/common'
import { type InternalGetValidatorDashboardSummaryChartResponse } from '~/types/api/validator_dashboard'

export const useValidatorDashboardSummaryChartStore = defineStore('useValidatorDashboardSummaryChartStore', () => {
  const chartData = ref<ChartData<number> | undefined >()

  async function getDashboardSummaryChart (dashboardId: number) {
    const response = await useCustomFetch<InternalGetValidatorDashboardSummaryChartResponse>(API_PATH.DASHBOARD_SUMMARY_CHART, undefined, { dashboardId })
    chartData.value = response.data
    return chartData.value
  }

  return { chartData, getDashboardSummaryChart }
})
