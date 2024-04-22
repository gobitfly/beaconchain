import { defineStore } from 'pinia'
import type { InternalGetValidatorDashboardWithdrawalsResponse } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import type { TableQueryParams } from '~/types/datatable'

const validatorDashboardWithdrawalsStore = defineStore('validator_dashboard_withdrawals', () => {
  const data = ref < InternalGetValidatorDashboardWithdrawalsResponse>()
  const query = ref < TableQueryParams>()

  return { data, query }
})

export function useValidatorDashboardWithdrawalsStore () {
  const { fetch } = useCustomFetch()
  const { data, query: storedQuery } = storeToRefs(validatorDashboardWithdrawalsStore())

  const withdrawals = computed(() => data.value)
  const query = computed(() => storedQuery.value)

  async function getWithdrawals (dashboardKey: DashboardKey, query?: TableQueryParams) {
    if (dashboardKey === undefined) {
      return undefined
    }
    storedQuery.value = query
    const res = await fetch<InternalGetValidatorDashboardWithdrawalsResponse>(API_PATH.DASHBOARD_VALIDATOR_WITHDRAWALS, undefined, { dashboardKey }, query)

    if (JSON.stringify(storedQuery.value) !== JSON.stringify(query)) {
      return // in case some query params change while loading
    }

    data.value = res
    return res
  }

  return { withdrawals, query, getWithdrawals }
}
