import { defineStore } from 'pinia'
import type { VDBGroupRewardsData, InternalGetValidatorDashboardGroupRewardsResponse } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'

const validatorDashboardRewardsDetailsStore = defineStore('validator_dashboard_rewards_details', () => {
  const data = ref < Record<string, VDBGroupRewardsData >>({})
  return { data }
})

export const useValidatorDashboardRewardsDetailsStore = (dashboardKey: DashboardKey, groupId: number, epoch: number) => {
  const { fetch } = useCustomFetch()
  const { data } = storeToRefs(validatorDashboardRewardsDetailsStore())

  function getKey () {
    return `${dashboardKey}_${groupId}`
  }

  async function getDetails () {
    const res = await fetch<InternalGetValidatorDashboardGroupRewardsResponse>(API_PATH.DASHBOARD_VALIDATOR_REWARDS_DETAILS, undefined, { dashboardKey, groupId, epoch })
    data.value = { ...data.value, [getKey()]: res.data }
    return res.data
  }

  watch(() => getKey(), getDetails, { immediate: true })

  const details = computed<VDBGroupRewardsData | undefined>(() => {
    return data.value[getKey()]
  })

  return { details }
}
