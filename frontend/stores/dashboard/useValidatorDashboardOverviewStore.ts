import { defineStore } from 'pinia'
import type { VDBOverviewData, InternalGetValidatorDashboardResponse } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'

const validatorOverviewStore = defineStore('validator_overview_store', () => {
  const data = ref<VDBOverviewData | undefined | null>()
  return { data }
})

export function useValidatorDashboardOverviewStore () {
  const { fetch } = useCustomFetch()
  const { data } = storeToRefs(validatorOverviewStore())

  const overview = computed(() => data.value)

  async function refreshOverview (key: DashboardKey) {
    if (!key) {
      return
    }
    const res = await fetch<InternalGetValidatorDashboardResponse>(API_PATH.DASHBOARD_OVERVIEW, undefined, { dashboardKey: key })
    data.value = res.data

    return overview.value
  }

  return { overview, refreshOverview }
}
