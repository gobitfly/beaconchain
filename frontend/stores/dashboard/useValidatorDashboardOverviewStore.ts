import { defineStore } from 'pinia'
import { warn } from 'vue'
import type { VDBOverviewData, InternalGetValidatorDashboardResponse } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'

const validatorOverviewStore = defineStore('validator_overview', () => {
  const data = ref<VDBOverviewData | undefined | null>()
  warn('DIECE: validatorOverviewStore = defineStore, data', data)
  const overview = readonly(data)

  const setOverview = (d: VDBOverviewData | undefined | null) => {
    data.value = d
  }

  return { overview, setOverview }
})

// this only works if called within a components script setup
// but that's good enough since the data returned below is reactive
export function useValidatorDashboardOverviewStore () {
  const { fetch } = useCustomFetch()
  const store = validatorOverviewStore()
  const { setOverview } = store
  const { overview } = storeToRefs(store)

  const key = inject<Ref<DashboardKey>>('dashboardKey')
  warn('DIECE: useValidatorDashboardOverviewStore, currently key is set to:', key?.value)

  async function getOverview () {
    warn('DIECE: getOverview() called')

    if (key === undefined) {
      return undefined
    }

    const res = await fetch<InternalGetValidatorDashboardResponse>(API_PATH.DASHBOARD_OVERVIEW, undefined, { dashboardKey: key.value })
    setOverview(res.data)
    return overview.value
  }

  // TODO: Have watch on key here?

  const { data, refresh } = useAsyncData('validator_overview', getOverview, { watch: [key!] }) // This store can only be used within dashboard/index and its children

  return { validatorDashboardOverview: data, refreshValidatorDashboardOverview: refresh }
}
