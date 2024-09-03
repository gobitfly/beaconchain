import { defineStore } from 'pinia'
import { useAllValidatorDashboardRewardsDetailsStore } from './useValidatorDashboardRewardsDetailsStore'
import type {
  GetValidatorDashboardResponse,
  VDBOverviewData,
} from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import { API_PATH } from '~/types/customFetch'

const validatorOverviewStore = defineStore('validator_overview_store', () => {
  const data = ref<null | undefined | VDBOverviewData>()
  return { data }
})

export function useValidatorDashboardOverviewStore() {
  const { fetch } = useCustomFetch()
  const { data } = storeToRefs(validatorOverviewStore())
  const { clearCache: clearRewardDetails }
    = useAllValidatorDashboardRewardsDetailsStore()

  const overview = computed(() => data.value)

  async function refreshOverview(key: DashboardKey) {
    if (!key) {
      data.value = undefined
      return
    }
    try {
      const res = await fetch<GetValidatorDashboardResponse>(
        API_PATH.DASHBOARD_OVERVIEW,
        undefined,
        { dashboardKey: key },
      )
      data.value = res.data

      clearOverviewDependentCaches()

      return overview.value
    }
    catch (e) {
      data.value = undefined
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

  return {
    hasValidators,
    overview,
    refreshOverview,
    validatorCount,
  }
}
