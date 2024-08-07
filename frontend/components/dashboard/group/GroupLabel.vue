<script setup lang="ts">
import type { VDBOverviewGroup } from '~/types/api/validator_dashboard'
import { DAHSHBOARDS_ALL_GROUPS_ID } from '~/types/dashboard'

interface Props {
  group: VDBOverviewGroup,
}

const props = defineProps<Props>()

const { t: $t } = useTranslation()

const name = computed(() => {
  if (props.group.id === DAHSHBOARDS_ALL_GROUPS_ID) {
    return $t('dashboard.group.selection.all')
  } else if (props.group.id === 0) {
    return props.group.name && props.group.name !== 'default' ? props.group.name : $t('dashboard.group.selection.default')
  }

  return props.group.name
})
</script>

<template>
  <span>{{ name }}</span>
</template>
