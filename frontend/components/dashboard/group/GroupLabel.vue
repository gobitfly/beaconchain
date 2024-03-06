<script setup lang="ts">
import type { VDBOverviewGroup } from '~/types/api/validator_dashboard'

interface Props {
  group: VDBOverviewGroup,
}

const props = defineProps<Props>()

const { t: $t } = useI18n()

const name = computed(() => {
  if (props.group.id === -1) {
    return $t('dashboard.group.selection.all')
  } else if (props.group.id === 0) {
    return $t('dashboard.group.selection.default')
  }

  return props.group.name
})
</script>

<template>
  <span><span>{{ name }}</span> <span v-if="group.id >= 0" class="id">(ID: {{ group.id }})</span></span>
</template>

<style lang="scss" scoped>
.id{
  color: var(--text-color-discreet);
}
</style>
