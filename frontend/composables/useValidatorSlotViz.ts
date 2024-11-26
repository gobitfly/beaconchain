import type { GetValidatorDashboardSlotVizResponse } from '~/types/api/slot_viz'
import type { DashboardKey } from '~/types/dashboard'
import { API_PATH } from '~/types/customFetch'

export function useValidatorSlotViz() {
  const { fetch } = useCustomFetch()

  // Actions
  async function fetchSlotVizData(dashboardKey: DashboardKey, groupIds?: number[]) {
    const query = groupIds?.length ? { group_ids: groupIds.join(',') } : undefined

    const res = await fetch<GetValidatorDashboardSlotVizResponse>(
      API_PATH.DASHBOARD_SLOTVIZ,
      {
        headers: {},
        query,
      },
      { dashboardKey: dashboardKey || 'MQ' }, // If guest dashboard has no validators yet (= empty dashboardKey), load small guest dashboard with 1 validator (MQ)
    )
    const data = res.data
    if (!dashboardKey) {
      data.forEach((epoch) => {
        epoch.slots?.forEach((slot) => {
          Object.assign(slot, {
            attestations: undefined, proposal: undefined, slashing: undefined, sync: undefined,
          })
        })
      })
    }

    return data
  }

  return {
    fetchSlotVizData,
  }
}
