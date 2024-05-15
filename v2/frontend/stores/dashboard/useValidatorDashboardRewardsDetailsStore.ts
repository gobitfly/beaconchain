import { defineStore } from 'pinia'
import type { VDBGroupRewardsData, InternalGetValidatorDashboardGroupRewardsResponse } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import { API_PATH } from '~/types/customFetch'

const validatorDashboardRewardsDetailsStore = defineStore('validator_dashboard_rewards_details', () => {
  const data = ref < Record<string, VDBGroupRewardsData >>({})
  return { data }
})

export const useValidatorDashboardRewardsDetailsStore = (dashboardKey: DashboardKey, groupId: number, epoch: number) => {
  const { fetch } = useCustomFetch()
  const { data } = storeToRefs(validatorDashboardRewardsDetailsStore())

  function getKey () {
    return `${dashboardKey}_${groupId}_${epoch}`
  }

  async function getDetails () {
    // Rewards of an epoch will not change, so we only need to load the data once
    if (data.value[getKey()]) {
      return data.value[getKey()]
    }
    const res = await fetch<InternalGetValidatorDashboardGroupRewardsResponse>(API_PATH.DASHBOARD_VALIDATOR_REWARDS_DETAILS, undefined, { dashboardKey, groupId, epoch })
    data.value = { ...data.value, [getKey()]: res.data }
    return res.data
  }

  // in the component where the store is used the properties will not change so we just need to load the data initially
  getDetails()

  const details = computed<VDBGroupRewardsData | undefined>(() => {
    return data.value[getKey()]
  })

  return { details }
}
