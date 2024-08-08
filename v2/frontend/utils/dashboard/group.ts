import type { ComposerTranslation } from 'vue-i18n'
import type { VDBOverviewGroup } from '~/types/api/validator_dashboard'
import {
  DAHSHBOARDS_ALL_GROUPS_ID,
  DAHSHBOARDS_NEXT_EPOCH_ID,
} from '~/types/dashboard'

export function getGroupLabel(
  t: ComposerTranslation,
  groupId?: number,
  groups?: VDBOverviewGroup[],
  allGroupLabel?: string,
): string {
  if (groupId === undefined) {
    return ''
  }
  if (groupId === DAHSHBOARDS_ALL_GROUPS_ID) {
    return allGroupLabel ?? t('dashboard.validator.summary.total_group_name')
  }
  else if (groupId === DAHSHBOARDS_NEXT_EPOCH_ID) {
    return '-'
  }
  const group = groups?.find(g => g.id === groupId)
  if (!group) {
    return `${groupId}` // fallback if we could not match the group name
  }
  return `${group.name}`
}
