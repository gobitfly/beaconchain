import { defineStore } from 'pinia'
import type { InternalGetValidatorDashboardRewardsResponse } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import type { TableQueryParams } from '~/types/datatable'

const validatorDashboardRewardsStore = defineStore('validator_dashboard_rewards', () => {
  const dataMap = ref < Record<DashboardKey, InternalGetValidatorDashboardRewardsResponse >>({})
  const queryMap = ref < Record<DashboardKey, TableQueryParams | undefined >>({})

  return { dataMap, queryMap }
})

export function useValidatorDashboardRewardsStore (dashboardKey: DashboardKey) {
  const { fetch } = useCustomFetch()
  const { dataMap, queryMap } = storeToRefs(validatorDashboardRewardsStore())

  const rewards = computed(() => dataMap.value[dashboardKey])
  const query = computed(() => queryMap.value[dashboardKey])

  async function getRewards (query?: TableQueryParams) {
    if (dashboardKey === undefined) {
      return undefined
    }
    queryMap.value = { ...queryMap.value, [dashboardKey]: query }

    const res = await fetch<InternalGetValidatorDashboardRewardsResponse>(API_PATH.DASHBOARD_VALIDATOR_REWARDS, undefined, { dashboardKey }, query)

    dataMap.value = { ...dataMap.value, [dashboardKey]: res }
    return res
  }

  return { rewards, query, getRewards }
}
