import { defineStore } from 'pinia'
import { useCustomFetch } from '~/composables/useCustomFetch'
import type { VDBGroupSummaryData, VDBGroupSummaryResponse } from '~/types/api/validator_dashboard'

export const useValidatorDashboardSummaryDetailsStore = defineStore('validator_dashboard_sumary_details_store', () => {
  const detailsMap = ref < Record<string, VDBGroupSummaryData >>({})

  function getKey (dashboardId: number, groupId: number) {
    return `${dashboardId}_${groupId}`
  }

  async function getDetails (dashboardId: number, groupId: number) {
    const res = await useCustomFetch<VDBGroupSummaryResponse>(API_PATH.DASHBOARD_SUMMARY_DETAILS, undefined, { dashboardId, groupId })
    detailsMap.value = { ...detailsMap.value, [getKey(dashboardId, groupId)]: res.data }
    return res.data
  }

  return { detailsMap, getDetails, getKey }
})
