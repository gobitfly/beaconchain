import { defineStore } from 'pinia'
import { useAllValidatorDashboardRewardsDetailsStore } from './useValidatorDashboardRewardsDetailsStore'
import type { VDBOverviewData, InternalGetValidatorDashboardResponse } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import { API_PATH } from '~/types/customFetch'

const validatorOverviewStore = defineStore('validator_overview_store', () => {
  const data = ref<VDBOverviewData | undefined | null>()
  return { data }
})

export function useValidatorDashboardOverviewStore () {
  const { fetch } = useCustomFetch()
  const { data } = storeToRefs(validatorOverviewStore())
  const { clearCache: clearRewardDetails } = useAllValidatorDashboardRewardsDetailsStore()

  const overview = computed(() => data.value)

  async function refreshOverview (key: DashboardKey) {
    if (!key) {
      data.value = undefined
      return
    }
    const res = await fetch<InternalGetValidatorDashboardResponse>(API_PATH.DASHBOARD_OVERVIEW, undefined, { dashboardKey: key })
    data.value = res.data

    clearOverviewDependentCaches()

    return overview.value
  }

  function clearOverviewDependentCaches () {
    clearRewardDetails()
  }

  const hasValidators = computed<boolean>(() => {
    if (!overview.value?.validators) {
      return false
    }
    return !!overview.value.validators.online || !!overview.value.validators.exited || !!overview.value.validators.offline || !!overview.value.validators.pending || !!overview.value.validators.slashed
  })

  return { overview, refreshOverview, hasValidators }
}
