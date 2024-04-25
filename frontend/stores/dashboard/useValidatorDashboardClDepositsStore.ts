import { defineStore } from 'pinia'
import type { InternalGetValidatorDashboardConsensusLayerDepositsResponse, InternalGetValidatorDashboardTotalConsensusDepositsResponse } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import type { TableQueryParams } from '~/types/datatable'

const validatorDashboardClDepositsStore = defineStore('validator_dashboard_cl_deposits', () => {
  const data = ref < InternalGetValidatorDashboardConsensusLayerDepositsResponse>()
  const total = ref < string >()
  const query = ref < TableQueryParams>()

  return { data, query, total }
})

export function useValidatorDashboardClDepositsStore () {
  const { fetch } = useCustomFetch()
  const { data, total, query: storedQuery } = storeToRefs(validatorDashboardClDepositsStore())

  const deposits = computed(() => data.value)
  const totalAmount = computed(() => total.value)
  const query = computed(() => storedQuery.value)
  const isLoadingDeposits = ref(false)
  const isLoadingTotal = ref(false)

  async function getDeposits (dashboardKey: DashboardKey, query?: TableQueryParams) {
    if (dashboardKey === undefined) {
      return undefined
    }
    storedQuery.value = query
    isLoadingDeposits.value = true
    const res = await fetch<InternalGetValidatorDashboardConsensusLayerDepositsResponse>(API_PATH.DASHBOARD_CL_DEPOSITS, undefined, { dashboardKey }, query)

    if (JSON.stringify(storedQuery.value) !== JSON.stringify(query)) {
      return // in case some query params change while loading
    }
    isLoadingDeposits.value = false

    data.value = res
    return res
  }

  async function getTotalAmount (dashboardKey: DashboardKey) {
    if (dashboardKey === undefined) {
      return undefined
    }
    isLoadingTotal.value = true
    const res = await fetch<InternalGetValidatorDashboardTotalConsensusDepositsResponse>(API_PATH.DASHBOARD_CL_DEPOSITS_TOTAL, undefined, { dashboardKey })
    isLoadingTotal.value = false
    total.value = res?.data?.total_amount
    return total.value
  }

  return { totalAmount, getTotalAmount, deposits, query, getDeposits, isLoadingTotal, isLoadingDeposits }
}
