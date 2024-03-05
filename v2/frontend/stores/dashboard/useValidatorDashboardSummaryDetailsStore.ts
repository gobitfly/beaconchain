import { defineStore } from 'pinia'
import { useCustomFetch } from '~/composables/useCustomFetch'
import type { VDBGroupSummaryData, InternalGetValidatorDashboardGroupSummaryResponse } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'

export const useValidatorDashboardSummaryDetailsStore = defineStore('validator_dashboard_sumary_details_store', () => {
  const detailsMap = ref < Record<string, VDBGroupSummaryData >>({})

  function getKey (dashboardKey: DashboardKey, groupId: number) {
    return `${dashboardKey}_${groupId}`
  }

  async function getDetails (dashboardKey: DashboardKey, groupId: number) {
    const res = await useCustomFetch<InternalGetValidatorDashboardGroupSummaryResponse>(API_PATH.DASHBOARD_SUMMARY_DETAILS, undefined, { dashboardKey, groupId })
    detailsMap.value = { ...detailsMap.value, [getKey(dashboardKey, groupId)]: res.data }
    return res.data
  }

  return { detailsMap, getDetails, getKey }
})
