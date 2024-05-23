import { defineStore } from 'pinia'
import type { InternalGetValidatorDashboardRewardsResponse } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import type { TableQueryParams } from '~/types/datatable'
import { DAHSHBOARDS_NEXT_EPOCH_ID } from '~/types/dashboard'
import { API_PATH } from '~/types/customFetch'

const validatorDashboardRewardsStore = defineStore('validator_dashboard_rewards_store', () => {
  const data = ref < InternalGetValidatorDashboardRewardsResponse>()
  const query = ref < TableQueryParams>()

  return { data, query }
})

export function useValidatorDashboardRewardsStore () {
  const { fetch } = useCustomFetch()
  const { data, query: storedQuery } = storeToRefs(validatorDashboardRewardsStore())

  const { slotViz } = useValidatorSlotVizStore()

  const rewards = computed(() => data.value)
  const query = computed(() => storedQuery.value)

  async function getRewards (dashboardKey: DashboardKey, query?: TableQueryParams) {
    if (!dashboardKey) {
      data.value = undefined
      return undefined
    }
    storedQuery.value = query
    const res = await fetch<InternalGetValidatorDashboardRewardsResponse>(API_PATH.DASHBOARD_VALIDATOR_REWARDS, undefined, { dashboardKey }, query)

    if (JSON.stringify(storedQuery.value) !== JSON.stringify(query)) {
      return // in case some query params change while loading
    }

    // If we are on the first page we get the next Epoch slot viz data and create a future entry
    if (!query?.cursor && slotViz.value && res.data?.length) {
      const searchEpoch = res.data[0].epoch
      const nextEpoch = slotViz.value?.findLast(e => e.epoch > searchEpoch)

      if (nextEpoch) {
        res.data = [{ epoch: nextEpoch.epoch, group_id: DAHSHBOARDS_NEXT_EPOCH_ID, duty: { attestation: 0, proposal: 0, slashing: 0, sync: 0 }, reward: { cl: '0', el: '0' } }, ...res.data]
      }
    }

    data.value = res
    return res
  }

  return { rewards, query, getRewards }
}
