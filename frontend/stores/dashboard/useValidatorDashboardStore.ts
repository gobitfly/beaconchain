import type { ChartHistorySeconds } from '~/types/api/common'
import type {
  VDBOverviewData, VDBOverviewGroup,
} from '~/types/api/validator_dashboard'

export const useValidatorDashboardStore = defineStore(
  'validator_dashboard_store',
  () => {
    const { t: $t } = useTranslation()
    // TODO: bring ID in here and remove the provider composable!!
    const validatorCount = ref<null | number>(null)
    const chainId = ref<null | number>(null)
    const rawGroups = ref<null | VDBOverviewGroup[]>(null)
    const chartHistorySeconds = ref<ChartHistorySeconds | null>(null)

    const hasValidators = computed<boolean>(() => {
      if (!validatorCount.value) {
        return false
      }
      return validatorCount.value > 0
    })
    const hasAbilityChartHistory = computed(() => ({
      daily: (chartHistorySeconds.value?.daily ?? 0) > 0,
      epoch: (chartHistorySeconds.value?.epoch ?? 0) > 0,
      hourly: (chartHistorySeconds.value?.hourly ?? 0) > 0,
      weekly: (chartHistorySeconds.value?.weekly ?? 0) > 0,
    }))
    const groups = computed<VDBOverviewGroup[]>(() => {
      if (!rawGroups.value) {
        return [ {
          count: 0,
          id: 0,
          name: $t('dashboard.group.selection.default'),
        } ]
      }

      return rawGroups.value
    })
    const populatedGroups = computed(() =>
      groups.value.filter(group => !!group.count),
    )

    const setByOverviewData = (data: VDBOverviewData) => {
      const v = data.validators
      validatorCount.value = v.exited + v.offline + v.online + v.pending + v.slashed
      chainId.value = data.network
      rawGroups.value = data.groups
      chartHistorySeconds.value = data.chart_history_seconds
    }

    const isLargeDashboard = computed(() => {
      if (!validatorCount.value) return false

      // This amount is a product decision
      const VALIDATOR_DASHBOARD_SIZE_THRESHOLD = 64

      return validatorCount.value > VALIDATOR_DASHBOARD_SIZE_THRESHOLD
    })

    const reset = () => {
      validatorCount.value = null
      chainId.value = null
      rawGroups.value = null
      chartHistorySeconds.value = null
    }

    return {
      chainId,
      chartHistorySeconds,
      groups,
      hasAbilityChartHistory,
      hasValidators,
      isLargeDashboard,
      populatedGroups,
      reset,
      setByOverviewData,
      validatorCount,
    }
  },
)
