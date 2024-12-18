import { defineStore } from 'pinia'
import type {
  GetValidatorDashboardSlotVizResponse,
  SlotVizEpoch,
} from '~/types/api/slot_viz'
import type { DashboardKey } from '~/types/dashboard'

const validatorSlotVizStore = defineStore('validator_slotViz', () => {
  const data = ref<null | SlotVizEpoch[] | undefined>()
  return { data }
})

export function useValidatorSlotVizStore() {
  const { fetch } = useCustomFetch()
  const { data } = storeToRefs(validatorSlotVizStore())

  const slotViz = computed(() => data.value)

  async function refreshSlotViz(dashboardKey: DashboardKey, groups?: number[]) {
    let query
    if (groups?.length) {
      query = { group_ids: groups.join(',') }
    }
    const res = await fetch<GetValidatorDashboardSlotVizResponse>(
      'DASHBOARD_SLOTVIZ',
      {
        headers: {},
        query,
      },
      { dashboardKey: dashboardKey || 'MQ' },
    )

    // We use this hacky solution as we don't have an api endpoint to load a slot viz without validators
    // So we load it for a small guest dashboard and then remove the validator informations from it.
    if (!dashboardKey) {
      data.value = res.data.map(e => ({
        ...e,
        slots: e.slots?.map(s => ({
          slot: s.slot,
          status: s.status,
        })),
      }))
    }
    else {
      data.value = res.data
    }

    return slotViz.value
  }

  return {
    refreshSlotViz,
    slotViz,
  }
}
