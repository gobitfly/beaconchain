import { defineStore } from 'pinia'
import type { InternalGetValidatorDashboardWithdrawalsResponse, InternalGetValidatorDashboardTotalWithdrawalsResponse } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import type { TableQueryParams } from '~/types/datatable'

const validatorDashboardWithdrawalsStore = defineStore('validator_dashboard_withdrawals', () => {
  const data = ref < InternalGetValidatorDashboardWithdrawalsResponse>()
  const query = ref < TableQueryParams>()

  return { data, query }
})

const validatorDashboardTotalWithdrawalStore = defineStore('validator_dashboard_total_withdrawal', () => {
  const data = ref < InternalGetValidatorDashboardTotalWithdrawalsResponse>()

  return { data }
})

export function useValidatorDashboardWithdrawalsStore () {
  const { fetch } = useCustomFetch()
  const { data: withdrawalData, query: storedWithdrawalQuery } = storeToRefs(validatorDashboardWithdrawalsStore())
  const { data: totalWithdrawalData } = storeToRefs(validatorDashboardTotalWithdrawalStore())

  const withdrawals = computed(() => withdrawalData.value)
  const query = computed(() => storedWithdrawalQuery.value)

  const totalWithdrawals = computed(() => totalWithdrawalData.value)

  async function getWithdrawals (dashboardKey: DashboardKey, query?: TableQueryParams) {
    if (dashboardKey === undefined) {
      return undefined
    }
    storedWithdrawalQuery.value = query
    const res = await fetch<InternalGetValidatorDashboardWithdrawalsResponse>(API_PATH.DASHBOARD_VALIDATOR_WITHDRAWALS, undefined, { dashboardKey }, query)

    if (JSON.stringify(storedWithdrawalQuery.value) !== JSON.stringify(query)) {
      return // in case some query params change while loading
    }

    withdrawalData.value = res
    return res
  }

  async function getTotalWithdrawals (dashboardKey: DashboardKey) {
    if (dashboardKey === undefined) {
      return undefined
    }

    const res = await fetch<InternalGetValidatorDashboardTotalWithdrawalsResponse>(API_PATH.DASHBOARD_VALIDATOR_TOTAL_WITHDRAWALS, undefined, { dashboardKey })

    totalWithdrawalData.value = res
    return res
  }

  return { withdrawals, query, getWithdrawals, totalWithdrawals, getTotalWithdrawals }
}
