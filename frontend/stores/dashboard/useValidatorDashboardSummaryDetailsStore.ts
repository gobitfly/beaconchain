import { defineStore } from 'pinia'
import type {
  GetValidatorDashboardGroupSummaryResponse,
  VDBGroupSummaryData,
} from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'

import type { SummaryTimeFrame } from '~/types/dashboard/summary'

const validatorDashboardSummaryDetailsStore = defineStore(
  'validator_dashboard_sumary_details_store',
  () => {
    const data = ref<Record<string, VDBGroupSummaryData>>({})
    const timeFrame = ref<SummaryTimeFrame>()
    return {
      data,
      timeFrame,
    }
  },
)

export function useValidatorDashboardSummaryDetailsStore(
  dashboardKey: DashboardKey,
  groupId: number,
) {
  const { fetch } = useCustomFetch()
  const {
    data, timeFrame: storeTimeFrame,
  } = storeToRefs(
    validatorDashboardSummaryDetailsStore(),
  )

  function getKey() {
    return `${dashboardKey}_${groupId}`
  }

  async function getDetails(timeFrame: SummaryTimeFrame) {
    // values might change so we reload whenever requested
    // values are cached in store to avoid loading spinner on expanding/collapsing rows though
    // except when the timeFrame changed, then we clear the cache
    if (storeTimeFrame.value !== timeFrame) {
      data.value = {}
      storeTimeFrame.value = timeFrame
    }
    const res = await fetch<GetValidatorDashboardGroupSummaryResponse>(
      'DASHBOARD_SUMMARY_DETAILS',
      { query: { period: timeFrame } },
      {
        dashboardKey,
        groupId,
      },
    )
    data.value = {
      ...data.value,
      [getKey()]: res.data,
    }
    return res.data
  }

  const details = computed<undefined | VDBGroupSummaryData>(() => {
    return data.value[getKey()]
  })

  return {
    details,
    getDetails,
    getKey,
  }
}
