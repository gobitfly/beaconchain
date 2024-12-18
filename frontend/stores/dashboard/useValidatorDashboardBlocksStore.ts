import { defineStore } from 'pinia'
import type { GetValidatorDashboardBlocksResponse } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import type { TableQueryParams } from '~/types/datatable'

const validatorDashboardBlocksStore = defineStore(
  'validator_dashboard_blocks_store',
  () => {
    const data = ref<GetValidatorDashboardBlocksResponse>()
    const query = ref<TableQueryParams>()

    return {
      data,
      query,
    }
  },
)

export function useValidatorDashboardBlocksStore() {
  const { fetch } = useCustomFetch()
  const {
    data, query: storedQuery,
  } = storeToRefs(
    validatorDashboardBlocksStore(),
  )
  const isLoading = ref(false)

  const blocks = computed(() => data.value)
  const query = computed(() => storedQuery.value)

  async function getBlocks(
    dashboardKey: DashboardKey,
    query?: TableQueryParams,
  ) {
    if (!dashboardKey) {
      data.value = undefined
      isLoading.value = false
      storedQuery.value = undefined
      return undefined
    }
    isLoading.value = true
    storedQuery.value = query
    const res = await fetch<GetValidatorDashboardBlocksResponse>(
      'DASHBOARD_VALIDATOR_BLOCKS',
      undefined,
      { dashboardKey },
      query,
    )
    isLoading.value = false

    if (JSON.stringify(storedQuery.value) !== JSON.stringify(query)) {
      return // in case some query params change while loading
    }

    data.value = res
    return res
  }

  return {
    blocks,
    getBlocks,
    isLoading,
    query,
  }
}
