import { defineStore } from 'pinia'
import type { InternalGetValidatorDashboardGroupHeatmapResponse, InternalGetValidatorDashboardHeatmapResponse, VDBHeatmap, VDBHeatmapTooltipData } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import { API_PATH } from '~/types/customFetch'

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

  async function getHeatmap (dashboardKey: DashboardKey) {
    if (!dashboardKey) {
      data.value = undefined
      return
    }
    tooltipDataMap.value = {}
    isLoading.value = true

    const res = await fetch<InternalGetValidatorDashboardHeatmapResponse>(API_PATH.DASHBOARD_HEATMAP, undefined, { dashboardKey })
    isLoading.value = false

    data.value = res.data
    return res.data
  }

  async function getHeatmapTooltip (dashboardKey: DashboardKey, groupId: number, epoch: number) {
    const key = getKey(groupId, epoch)
    if (tooltipDataMap.value[key]) {
      return tooltipDataMap.value[key]
    }
    const res = await fetch<InternalGetValidatorDashboardGroupHeatmapResponse>(API_PATH.DASHBOARD_HEATMAP_DETAILS, undefined, { dashboardKey, groupId, epoch })

    tooltipDataMap.value[key] = res.data
    return res.data
  }

  return { heatmap, isLoading, getHeatmap, getHeatmapTooltip }
}
