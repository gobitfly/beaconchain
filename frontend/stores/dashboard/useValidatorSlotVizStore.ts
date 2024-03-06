import { defineStore } from 'pinia'
import { useCustomFetch } from '~/composables/useCustomFetch'
import type { SlotVizEpoch, InternalGetValidatorDashboardSlotVizResponse } from '~/types/api/slot_viz'
import type { DashboardKey } from '~/types/dashboard'

export const useValidatorSlotVizStore = defineStore('validator_slotViz', () => {
  const slotViz = ref<SlotVizEpoch[] | undefined | null>()
  async function getSlotViz (dashboardKey: DashboardKey | string) {
    const res = await useCustomFetch<InternalGetValidatorDashboardSlotVizResponse>(API_PATH.DASHBOARD_SLOTVIZ, { headers: {} }, { dashboardKey })
    slotViz.value = res.data
    return slotViz.value
  }

  return { slotViz, getSlotViz }
})
