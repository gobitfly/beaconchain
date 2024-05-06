import { defineStore } from 'pinia'
import type { InternalGetValidatorDashboardGroupHeatmapResponse, InternalGetValidatorDashboardHeatmapResponse, VDBHeatmap, VDBHeatmapTooltipData } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import { API_PATH } from '~/types/customFetch'
import type { HeatmapTimeFrame } from '~/types/dashboard/heatmap'

const validatorDashboardHeatmapStore = defineStore('validator_dashboard_heatmap_store', () => {
  const data = ref < VDBHeatmap>()

  return { data }
})

export function useValidatorDashboardHeatmapStore () {
  const { fetch } = useCustomFetch()

  const { data } = storeToRefs(validatorDashboardHeatmapStore())
  const tooltipDataMap = ref < Record<string, VDBHeatmapTooltipData >>({})
  const isLoading = ref(false)

  const heatmap = computed(() => data.value)
  function getKey (group: number, epoch: number) {
    return `${group}_${epoch}`
  }

  async function getHeatmap (dashboardKey: DashboardKey, timeFrame: HeatmapTimeFrame) {
    if (!dashboardKey) {
      data.value = undefined
      return
    }
    tooltipDataMap.value = {}
    isLoading.value = true

    const path = timeFrame === '24h' ? API_PATH.DASHBOARD_HEATMAP_EPOCH : API_PATH.DASHBOARD_HEATMAP_DAILY

    const res = await fetch<InternalGetValidatorDashboardHeatmapResponse>(path, { query: { period: timeFrame === '24h' ? null : timeFrame } }, { dashboardKey })
    isLoading.value = false

    data.value = res.data
    return res.data
  }

  async function getHeatmapTooltip (dashboardKey: DashboardKey, groupId: number, date: number /** can either be a ts or an epoch */, timeFrame: HeatmapTimeFrame) {
    const key = getKey(groupId, date)
    if (tooltipDataMap.value[key]) {
      return tooltipDataMap.value[key]
    }
    const path = timeFrame === '24h' ? API_PATH.DASHBOARD_HEATMAP_EPOCH_DETAILS : API_PATH.DASHBOARD_HEATMAP_DAILY_DETAILS
    const res = await fetch<InternalGetValidatorDashboardGroupHeatmapResponse>(path, undefined, { dashboardKey, groupId, date })

    tooltipDataMap.value[key] = res.data
    return res.data
  }

  return { heatmap, isLoading, getHeatmap, getHeatmapTooltip }
}
