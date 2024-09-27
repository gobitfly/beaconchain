import { defineStore } from 'pinia'
import type {
  GetValidatorDashboardExecutionLayerDepositsResponse,
  GetValidatorDashboardTotalExecutionDepositsResponse,
} from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import type { TableQueryParams } from '~/types/datatable'
import { API_PATH } from '~/types/customFetch'

const validatorDashboardElDepositsStore = defineStore(
  'validator_dashboard_el_deposits_store',
  () => {
    const data
      = ref<GetValidatorDashboardExecutionLayerDepositsResponse>()
    const total = ref<string>()
    const query = ref<TableQueryParams>()

    return {
      data,
      query,
      total,
    }
  },
)

export function useValidatorDashboardElDepositsStore() {
  const { fetch } = useCustomFetch()
  const {
    data,
    query: storedQuery,
    total,
  } = storeToRefs(validatorDashboardElDepositsStore())

  const deposits = computed(() => data.value)
  const totalAmount = computed(() => total.value)
  const query = computed(() => storedQuery.value)
  const isLoadingDeposits = ref(false)
  const isLoadingTotal = ref(false)

  async function getDeposits(
    dashboardKey: DashboardKey,
    query?: TableQueryParams,
  ) {
    if (!dashboardKey) {
      data.value = undefined
      isLoadingDeposits.value = false
      storedQuery.value = undefined
      return undefined
    }
    storedQuery.value = query
    isLoadingDeposits.value = true
    const res
      = await fetch<GetValidatorDashboardExecutionLayerDepositsResponse>(
        API_PATH.DASHBOARD_EL_DEPOSITS,
        undefined,
        { dashboardKey },
        query,
      )

    if (JSON.stringify(storedQuery.value) !== JSON.stringify(query)) {
      return // in case some query params change while loading
    }
    isLoadingDeposits.value = false

    data.value = res
    return res
  }

  async function getTotalAmount(dashboardKey: DashboardKey) {
    if (!dashboardKey) {
      total.value = undefined
      isLoadingTotal.value = false
      return undefined
    }
    isLoadingTotal.value = true
    const res
      = await fetch<GetValidatorDashboardTotalExecutionDepositsResponse>(
        API_PATH.DASHBOARD_EL_DEPOSITS_TOTAL,
        undefined,
        { dashboardKey },
      )
    isLoadingTotal.value = false
    total.value = res?.data?.total_amount
    return total.value
  }

  return {
    deposits,
    getDeposits,
    getTotalAmount,
    isLoadingDeposits,
    isLoadingTotal,
    query,
    totalAmount,
  }
}
