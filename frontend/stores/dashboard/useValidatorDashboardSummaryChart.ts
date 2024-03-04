import { defineStore } from 'pinia'
import { useCustomFetch } from '~/composables/useCustomFetch'
import { type ChartData } from '~/types/api/common'

export const useValidatorDashboardSummaryChartStore = defineStore('useValidatorDashboardSummaryChartStore', () => {
  const chartData = ref<ChartData<number> | undefined >()

  async function getDashboardSummaryChart (dashboardId: number) {
    chartData.value = await useCustomFetch<ChartData<number>>(API_PATH.DASHBOARD_SUMMARY_CHART, undefined, { dashboardId })
    return chartData.value
  }

  return { chartData, getDashboardSummaryChart }
})
