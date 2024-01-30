import { defineStore } from 'pinia'
import type { SlotVizData } from '~/types/dashboard/slotViz'
import { useCustomFetch } from '~/composables/useCustomFetch'

export const useValidatorSlotVizStore = defineStore('validator_slotViz', () => {
  const slotViz = ref<SlotVizData | undefined | null>()
  async function getSlotViz () {
    const res = await useCustomFetch<SlotVizData>(API_PATH.DASHBOARD_SLOTVIZ)
    slotViz.value = res
    return slotViz.value
  }

  return { slotViz, getSlotViz }
})
