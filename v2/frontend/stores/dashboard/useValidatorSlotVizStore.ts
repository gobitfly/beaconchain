import { defineStore } from 'pinia'
import type { SlotVizEpoch, InternalGetValidatorDashboardSlotVizResponse } from '~/types/api/slot_viz'
import type { DashboardKey } from '~/types/dashboard'

const validatorSlotVizStore = defineStore('validator_slotViz', () => {
  const data = ref<SlotVizEpoch[] | undefined | null>()
  return { data }
})

export function useValidatorSlotVizStore () {
  const { fetch } = useCustomFetch()
  const { data } = storeToRefs(validatorSlotVizStore())

  const slotViz = computed(() => data.value)

  async function refreshSlotViz (dashboardKey: DashboardKey | string) {
    const res = await fetch<InternalGetValidatorDashboardSlotVizResponse>(API_PATH.DASHBOARD_SLOTVIZ, { headers: {} }, { dashboardKey })
    data.value = res.data

    return slotViz.value
  }

  return { slotViz, refreshSlotViz }
}
