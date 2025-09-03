import { defineStore } from 'pinia'
import { useAllValidatorDashboardRewardsDetailsStore } from './useValidatorDashboardRewardsDetailsStore'
import type {
  GetValidatorDashboardResponse,
  VDBOverviewData,
} from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'

export const useValidatorDashboardOverviewStore = defineStore('validator-dashboard-overview', () => {
  const overview = ref<undefined | VDBOverviewData>()
  const loading = ref(false)
  const { fetch } = useCustomFetch()
  const { clearCache: clearRewardDetails }
    = useAllValidatorDashboardRewardsDetailsStore()

  async function refreshOverview(key: DashboardKey) {
    if (!key) {
      overview.value = undefined
      return
    }
    try {
      loading.value = true
      const res = await fetch<GetValidatorDashboardResponse>(
        'DASHBOARD_OVERVIEW',
        undefined,
        { dashboardKey: key },
      )
      overview.value = res.data
      loading.value = false

      clearOverviewDependentCaches()

      return overview.value
    }
    catch (e) {
      overview.value = undefined
      clearOverviewDependentCaches()

      throw e
    }
  }

  function clearOverviewDependentCaches() {
    clearRewardDetails()
  }

  const hasValidators = computed<boolean>(() => {
    if (!overview.value?.validators) {
      return false
    }
    return (
      !!overview.value.validators.online
      || !!overview.value.validators.exited
      || !!overview.value.validators.offline
      || !!overview.value.validators.pending
      || !!overview.value.validators.slashed
    )
  })

  const validatorCount = computed(() => {
    if (!overview.value) {
      return undefined
    }
    if (!overview.value.validators) {
      return 0
    }
    return (
      overview.value.validators.exited
      + overview.value.validators.offline
      + overview.value.validators.online
      + overview.value.validators.pending
      + overview.value.validators.slashed
    )
  })

  const isLargeDashboard = computed(() => {
    if (!validatorCount.value) return false

    // This amount is a product decision
    const VALIDATOR_DASHBOARD_SIZE_THRESHOLD = 64

    return validatorCount.value > VALIDATOR_DASHBOARD_SIZE_THRESHOLD
  })

  const groups = computed(() => overview.value?.groups)

  const hasAbilityRewardsChartHistory = computed(() => ({
    daily: (overview.value?.rewards_chart_history_seconds?.daily ?? 0) > 0,
    epoch: (overview.value?.rewards_chart_history_seconds?.epoch ?? 0) > 0,
    hourly: (overview.value?.rewards_chart_history_seconds?.hourly ?? 0) > 0,
    weekly: (overview.value?.rewards_chart_history_seconds?.weekly ?? 0) > 0,
  }))

  const hasAbilityChartHistory = computed(() => ({
    daily: (overview.value?.chart_history_seconds?.daily ?? 0) > 0,
    epoch: (overview.value?.chart_history_seconds?.epoch ?? 0) > 0,
    hourly: (overview.value?.chart_history_seconds?.hourly ?? 0) > 0,
    weekly: (overview.value?.chart_history_seconds?.weekly ?? 0) > 0,
  }))

  return {
    groups,
    hasAbilityChartHistory,
    hasAbilityRewardsChartHistory,
    hasValidators,
    isLargeDashboard,
    loading,
    overview,
    refreshOverview,
    validatorCount,
  }
},
)
