<script setup lang="ts">
import { DAHSHBOARDS_ALL_GROUPS_ID } from '~/types/dashboard'

interface Props {
  includeAll?: boolean,
}
const props = defineProps<Props>()

const emit = defineEmits<{(e: 'setGroup', value: number): void}>()

const { overview } = storeToRefs(useValidatorDashboardOverviewStore())

const list = computed(() => {
  const groups = overview.value?.groups ?? []
  if (props.includeAll) {
    return [{ id: DAHSHBOARDS_ALL_GROUPS_ID, name: '', count: 0 }].concat(groups)
  }
  return groups
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
    @update:model-value="(value: number)=>emit('setGroup', value)"
  >
    <template v-if="selectedGroup" #value>
      <span>{{ $t('dashboard.group.selection.group') }}:
        <DashboardGroupLabel :group="selectedGroup" />
      </span>
    </template>

    <template #option="slotProps">
      <DashboardGroupLabel :group="slotProps" />
    </template>
  </BcDropdown>
</template>

<style lang="scss" scoped></style>
