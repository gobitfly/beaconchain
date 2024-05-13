import type { VDBOverviewGroup } from '~/types/api/validator_dashboard'

export function useValidatorDashboardGroups () {
  const { overview } = useValidatorDashboardOverviewStore()
  const { t: $t } = useI18n()

  const groups = computed<VDBOverviewGroup[]>(() => {
    if (!overview.value?.groups) {
      return [{
        id: 0,
        name: $t('dashboard.group.selection.default'),
        count: 0
      }]
    }

    return overview.value.groups
  })

  return { groups }
}
