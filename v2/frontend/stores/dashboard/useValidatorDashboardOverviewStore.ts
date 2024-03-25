import { defineStore } from 'pinia'
import { warn } from 'vue'
import type { InternalGetValidatorDashboardResponse } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'

export const useValidatorDashboardOverviewStore = defineStore('validator_overview', () => {
  const { fetch } = useCustomFetch()

  warn('DIECE: Creating useValidatorDashboardOverviewStore')
  const key = inject<Ref<DashboardKey>>('dashboardKey')
  warn('DIECE: currently key is set to:', key?.value)

  async function getOverview () {
    warn('DIECE: Fetching for useValidatorDashboardOverviewStore <InternalGetValidatorDashboardResponse>, dashboardKey:', key?.value)

    if (key === undefined) {
      return undefined
    }

    const res = await fetch<InternalGetValidatorDashboardResponse>(API_PATH.DASHBOARD_OVERVIEW, undefined, { dashboardKey: key.value! })
    warn('DIECE: Fetching DONE in store and got', res.data)
    return res.data
  }

  const { data, pending, refresh } = useAsyncData('validator_overview', getOverview, { watch: [key!] }) // This store can only be used within dashboard/index and its children

  return { key, validatorDashboardOverview: data, refreshValidatorDashboardOverview: refresh, pending }
})
