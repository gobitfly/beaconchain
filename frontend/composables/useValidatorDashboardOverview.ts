import type {
  GetValidatorDashboardResponse,
  VDBOverviewData,
} from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import { API_PATH } from '~/types/customFetch'

export function useValidatorDashboardOverview() {
  const { fetch } = useCustomFetch()

  const overviewData = ref<VDBOverviewData>()
  async function fetchOverviewData(key: DashboardKey) {
    if (!key) {
      throw new Error('No key provided when fetching overview')
    }
    const res = await fetch<GetValidatorDashboardResponse>(
      API_PATH.DASHBOARD_OVERVIEW,
      undefined,
      { dashboardKey: key },
    )
    return res.data
  }
  /* const hasValidators = computed<boolean>(() => {
    if (!overviewData.value?.validators) {
      return false
    }
    return (
      !!overviewData.value.validators.online
      || !!overviewData.value.validators.exited
      || !!overviewData.value.validators.offline
      || !!overviewData.value.validators.pending
      || !!overviewData.value.validators.slashed
    )
  })

  const validatorCount = computed(() => {
    if (!overviewData.value) {
      return undefined
    }
    if (!overviewData.value.validators) {
      return 0
    }
    return (
      overviewData.value.validators.exited
      + overviewData.value.validators.offline
      + overviewData.value.validators.online
      + overviewData.value.validators.pending
      + overviewData.value.validators.slashed
    )
  })

  const hasAbilityCharthistory = computed(() => ({
    daily: (overviewData.value?.chart_history_seconds?.daily ?? 0) > 0,
    epoch: (overviewData.value?.chart_history_seconds?.epoch ?? 0) > 0,
    hourly: (overviewData.value?.chart_history_seconds?.hourly ?? 0) > 0,
    weekly: (overviewData.value?.chart_history_seconds?.weekly ?? 0) > 0,
  })) */

  return {
    fetchOverviewData,
    overviewData,
  }
}
