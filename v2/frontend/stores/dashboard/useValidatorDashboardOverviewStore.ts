import { defineStore } from 'pinia'
import { warn } from 'vue'
import type { VDBOverviewData, InternalGetValidatorDashboardResponse } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'

const validatorOverviewStore = defineStore('validator_overview', () => {
  const data = ref<VDBOverviewData | undefined | null>()
  warn('DIECE: validatorOverviewStore = defineStore')
  const overview = readonly(data)

  const setOverview = (d: VDBOverviewData | undefined | null) => {
    warn('DIECE: setOverview() called')
    data.value = d
  }

  return { overview, setOverview }
})

// NOTE: this only works if called within a components script setup
// but that's good enough since the data returned below is reactive
export function useValidatorDashboardOverviewStore () {
  const { fetch } = useCustomFetch()
  const store = validatorOverviewStore()
  const { setOverview } = store
  const { overview } = storeToRefs(store) // NOTE: this is the REACTIVE (and read only) data

  // NOTE: function to UPDATE the data, use overview if you just want to access the data
  async function getOverview (key: DashboardKey) {
    warn('DIECE: getOverview() called, key:', key)

    if (key === undefined) {
      return undefined
    }

    const res = await fetch<InternalGetValidatorDashboardResponse>(API_PATH.DASHBOARD_OVERVIEW, undefined, { dashboardKey: key })
    setOverview(res.data)
    return overview.value
  }

  // TODO: maybe this should just be data and refresh and every component that uses that chooses a name to their liking?
  return { validatorDashboardOverview: overview, refreshValidatorDashboardOverview: getOverview }
}
