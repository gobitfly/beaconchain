import { defineStore } from 'pinia'
import type { InternalGetValidatorDashboardBlocksResponse } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import type { TableQueryParams } from '~/types/datatable'

const validatorDashboardBlocksStore = defineStore('validator_dashboard_blocks', () => {
  const data = ref < InternalGetValidatorDashboardBlocksResponse>()
  const query = ref < TableQueryParams>()

  return { data, query }
})

export function useValidatorDashboardBlocksStore () {
  const { fetch } = useCustomFetch()
  const { data, query: storedQuery } = storeToRefs(validatorDashboardBlocksStore())

  const blocks = computed(() => data.value)
  const query = computed(() => storedQuery.value)

  async function getBlocks (dashboardKey: DashboardKey, query?: TableQueryParams) {
    if (!dashboardKey) {
      data.value = undefined
      return undefined
    }
    storedQuery.value = query
    const res = await fetch<InternalGetValidatorDashboardBlocksResponse>(API_PATH.DASHBOARD_VALIDATOR_BLOCKS, undefined, { dashboardKey }, query)

    if (JSON.stringify(storedQuery.value) !== JSON.stringify(query)) {
      return // in case some query params change while loading
    }

    data.value = res
    return res
  }

  return { blocks, query, getBlocks }
}
