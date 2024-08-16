<script setup lang="ts">
import type { VDBOverviewGroup } from '~/types/api/validator_dashboard'
import { DAHSHBOARDS_ALL_GROUPS_ID } from '~/types/dashboard'

interface Props {
  includeAll?: boolean,
}
const props = defineProps<Props>()

const emit = defineEmits<{ (e: 'setGroup', value: number): void }>()

const { groups } = useValidatorDashboardGroups()

const list = computed<VDBOverviewGroup[]>(() => {
  if (props.includeAll) {
    return [ {
      count: 0,
      id: DAHSHBOARDS_ALL_GROUPS_ID,
      name: '',
    } ].concat(
      groups.value,
    )
  }
  return groups.value
})

const selected = defineModel<number | undefined>({ required: true })

const selectedGroup = computed(() => {
  return list.value.find(item => item.id === selected.value)
})
</script>

<template>
  <BcDropdown
    v-model="selected"
    :options="list"
    option-value="id"
    option-label="name"
    :placeholder="$t('dashboard.group.selection.placeholder')"
    @update:model-value="(value: number) => emit('setGroup', value)"
  >
    <template
      v-if="selectedGroup"
      #value
    >
      <DashboardGroupLabel :group="selectedGroup" />
    </template>

    <template #option="slotProps">
      <DashboardGroupLabel :group="slotProps" />
    </template>
  </BcDropdown>
</template>

<style lang="scss" scoped></style>
