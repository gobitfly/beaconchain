import { defineStore } from 'pinia'
import type { VDBOverviewData, InternalGetValidatorDashboardResponse } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'

const validatorOverviewStore = defineStore('validator_overview_store', () => {
  const data = ref<VDBOverviewData | undefined | null>()
  return { data }
})

export function useValidatorDashboardOverviewStore () {
  const { fetch } = useCustomFetch()
  const { t: $t } = useI18n()
  const { data } = storeToRefs(validatorOverviewStore())

  const overview = computed(() => data.value)

  async function refreshOverview (key: DashboardKey) {
    if (!key) {
      data.value = undefined
      return
    }
    const res = await fetch<InternalGetValidatorDashboardResponse>(API_PATH.DASHBOARD_OVERVIEW, undefined, { dashboardKey: key })
    data.value = res.data

    if (data.value.groups === null) {
      data.value.groups = []
      data.value.groups.push({
        id: 0,
        name: $t('dashboard.group.selection.default'),
        count: 0
      })
    }

    return overview.value
  }

  return { overview, refreshOverview }
}
