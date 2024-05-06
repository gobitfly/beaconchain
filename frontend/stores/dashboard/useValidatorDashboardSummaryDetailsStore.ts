import { defineStore } from 'pinia'
import type { VDBGroupSummaryData, InternalGetValidatorDashboardGroupSummaryResponse } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import { API_PATH } from '~/types/customFetch'

const validatorDashboardSummaryDetailsStore = defineStore('validator_dashboard_sumary_details_store', () => {
  const data = ref < Record<string, VDBGroupSummaryData >>({})
  return { data }
})

export function useValidatorDashboardSummaryDetailsStore (dashboardKey: DashboardKey, groupId: number) {
  const { fetch } = useCustomFetch()
  const { data } = storeToRefs(validatorDashboardSummaryDetailsStore())

  function getKey () {
    return `${dashboardKey}_${groupId}`
  }

  async function getDetails () {
    // values might change so we reload whenever requested
    // values are cached in store to avoid loading spinner on expanding/collapsing rows though

    const res = await fetch<InternalGetValidatorDashboardGroupSummaryResponse>(API_PATH.DASHBOARD_SUMMARY_DETAILS, undefined, { dashboardKey, groupId })
    data.value = { ...data.value, [getKey()]: res.data }
    return res.data
  }

  // in the component where the store is used the properties will not change so we just need to load the data initially
  getDetails()

  const details = computed<VDBGroupSummaryData | undefined>(() => {
    return data.value[getKey()]
  })

  return { details, getDetails, getKey }
}
