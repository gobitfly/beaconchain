import type { ChartHistorySeconds } from '~/types/api/common'
import type {
  VDBOverviewData, VDBOverviewGroup,
} from '~/types/api/validator_dashboard'

export const useValidatorDashboardStore = defineStore(
  'validator_dashboard_store',
  () => {
    const { t: $t } = useTranslation()
    // TODO: bring ID in here and remove the provider composable!!
    const validatorCount = ref<number>()
    const chainId = ref<number>()
    const rawGroups = ref<VDBOverviewGroup[]>()
    const chartHistorySeconds = ref<ChartHistorySeconds>()

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

    function setByOverviewData(data: VDBOverviewData) {
      const v = data.validators
      validatorCount.value = v.exited + v.offline + v.online + v.pending + v.slashed
      chainId.value = data.network
      rawGroups.value = data.groups
      chartHistorySeconds.value = data.chart_history_seconds
    }

    return {
      chainId,
      chartHistorySeconds,
      groups,
      hasAbilityChartHistory,
      hasValidators,
      setByOverviewData,
      validatorCount,
    }
  },
)
