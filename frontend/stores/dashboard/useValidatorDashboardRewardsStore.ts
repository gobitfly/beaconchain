import { defineStore } from 'pinia'
import type { InternalGetValidatorDashboardRewardsResponse } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import type { TableQueryParams } from '~/types/datatable'
import { DAHSHBOARDS_NEXT_EPOCH_ID } from '~/types/dashboard'

const validatorDashboardRewardsStore = defineStore('validator_dashboard_rewards', () => {
  const dataMap = ref < Record<DashboardKey, InternalGetValidatorDashboardRewardsResponse >>({})
  const queryMap = ref < Record<DashboardKey, TableQueryParams | undefined >>({})

  return { dataMap, queryMap }
})

export function useValidatorDashboardRewardsStore (dashboardKey: DashboardKey) {
  const { fetch } = useCustomFetch()
  const { dataMap, queryMap } = storeToRefs(validatorDashboardRewardsStore())

  const { slotViz } = useValidatorSlotVizStore()

  const rewards = computed(() => dataMap.value[dashboardKey])
  const query = computed(() => queryMap.value[dashboardKey])

  async function getRewards (query?: TableQueryParams) {
    if (dashboardKey === undefined) {
      return undefined
    }
    queryMap.value = { ...queryMap.value, [dashboardKey]: query }

    const res = await fetch<InternalGetValidatorDashboardRewardsResponse>(API_PATH.DASHBOARD_VALIDATOR_REWARDS, undefined, { dashboardKey }, query)

    // If we are on the first page we get the next Epoch slot viz data and create a future entry
    if (!query?.cursor && slotViz.value && res.data?.length) {
      const searchEpoch = res.data[0].epoch
      const nextEpoch = slotViz.value?.findLast(e => e.epoch > searchEpoch)

      if (nextEpoch) {
        res.data = [{ epoch: nextEpoch.epoch, group_id: DAHSHBOARDS_NEXT_EPOCH_ID, duty: { attestation: 0, proposal: 0, slashing: 0, sync: 0 }, reward: { cl: '0', el: '0' } }, ...res.data]
      }
    }

    dataMap.value = { ...dataMap.value, [dashboardKey]: res }
    return res
  }

  return { rewards, query, getRewards }
}
