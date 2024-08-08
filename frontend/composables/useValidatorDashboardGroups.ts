import type { VDBOverviewGroup } from '~/types/api/validator_dashboard'

export function useValidatorDashboardGroups() {
  const { overview } = useValidatorDashboardOverviewStore()
  const { t: $t } = useTranslation()

  const groups = computed<VDBOverviewGroup[]>(() => {
    if (!overview.value?.groups) {
      return [
        {
          count: 0,
          id: 0,
          name: $t('dashboard.group.selection.default'),
        },
      ]
    }

    return overview.value.groups
  })

  return { groups }
}
