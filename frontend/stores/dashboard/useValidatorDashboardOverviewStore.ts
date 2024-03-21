import { defineStore } from 'pinia'
import type { InternalGetValidatorDashboardResponse } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'

export const useValidatorDashboardOverviewStore = defineStore('validator_overview', () => {
  const { fetch } = useCustomFetch()
  const key = ref<DashboardKey>('') // TODO: Key should not be here, there should be a key store and this store uses the key store, right? Alternatively use Provider / Receiver

  async function getOverview () {
    console.log('DIECE: Fetching for useValidatorDashboardOverviewStore <InternalGetValidatorDashboardResponse>, dashboardKey:', key.value)

    if (key === undefined) {
      return undefined
    }

    const res = await fetch<InternalGetValidatorDashboardResponse>(API_PATH.DASHBOARD_OVERVIEW, undefined, { dashboardKey: key.value! })
    return res.data
  }

  const { data, refresh } = useAsyncData('validator_overview', getOverview, { watch: [key] })
  return { key, validatorDashboardOverview: data, refreshValidatorDashboardOverview: refresh }
})
