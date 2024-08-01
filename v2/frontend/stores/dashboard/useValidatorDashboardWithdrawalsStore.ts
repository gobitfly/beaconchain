import { defineStore } from 'pinia'
import type { InternalGetValidatorDashboardWithdrawalsResponse, InternalGetValidatorDashboardTotalWithdrawalsResponse } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import type { TableQueryParams } from '~/types/datatable'
import { API_PATH } from '~/types/customFetch'

const validatorDashboardWithdrawalsStore = defineStore('validator_dashboard_withdrawals', () => {
  const data = ref<InternalGetValidatorDashboardWithdrawalsResponse>()
  const total = ref<string >()
  const query = ref<TableQueryParams>()

  return { data, query, total }
})

export function useValidatorDashboardWithdrawalsStore () {
  const { fetch } = useCustomFetch()
  const { data, total, query: storedQuery } = storeToRefs(validatorDashboardWithdrawalsStore())

  const withdrawals = computed(() => data.value)
  const totalAmount = computed(() => total.value)
  const query = computed(() => storedQuery.value)
  const isLoadingWithdrawals = ref(false)
  const isLoadingTotal = ref(false)

  async function getWithdrawals (dashboardKey: DashboardKey, query?: TableQueryParams) {
    if (!dashboardKey) {
      data.value = undefined
      isLoadingWithdrawals.value = false
      storedQuery.value = undefined
      return undefined
    }

    storedQuery.value = query
    isLoadingWithdrawals.value = true
    const res = await fetch<InternalGetValidatorDashboardWithdrawalsResponse>(API_PATH.DASHBOARD_VALIDATOR_WITHDRAWALS, undefined, { dashboardKey }, query)

    if (JSON.stringify(storedQuery.value) !== JSON.stringify(query)) {
      return // in case some query params change while loading
    }
    isLoadingWithdrawals.value = false

    data.value = res
    return res
  }

  async function getTotalAmount (dashboardKey: DashboardKey) {
    if (!dashboardKey) {
      total.value = undefined
      isLoadingTotal.value = false
      return undefined
    }

    isLoadingTotal.value = true
    const res = await fetch<InternalGetValidatorDashboardTotalWithdrawalsResponse>(API_PATH.DASHBOARD_VALIDATOR_TOTAL_WITHDRAWALS, undefined, { dashboardKey })
    isLoadingTotal.value = false
    total.value = res?.data?.total_amount
    return total.value
  }

  return { totalAmount, getTotalAmount, withdrawals, query, getWithdrawals, isLoadingTotal, isLoadingWithdrawals }
}
