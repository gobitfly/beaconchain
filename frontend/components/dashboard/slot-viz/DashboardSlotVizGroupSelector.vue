<script setup lang="ts">
import { orderBy } from 'lodash-es'
import { getGroupLabel } from '~/utils/dashboard/group'
import type { VDBOverviewGroup } from '~/types/api/validator_dashboard'

const props = defineProps<{ validatorGroups: VDBOverviewGroup[] }>()

const { t: $t } = useTranslation()

const selectedGroups = defineModel<number[]>('selectedGroups', {
  default: [],
})

const groups = computed(() => {
  return orderBy(
    props.validatorGroups,
    [ g => g.name.toLowerCase() ],
    'asc',
  )
})
const selectedLabel = computed(() => {
  if (
    selectedGroups.value.length === 0
    || selectedGroups.value.length === groups.value.length
  ) {
    return $t('dashboard.group.selection.all')
  }
  return orderBy(
    selectedGroups.value.map(id => getGroupLabel($t, id, groups.value)),
    [ g => g.toLowerCase() ],
    'asc',
  ).join(', ')
})

const selectAll = () => {
  selectedGroups.value = groups.value.map(g => g.id)
}
const toggleAll = () => {
  if (selectedGroups.value.length < groups.value.length) {
    selectAll()
  }
  else {
    selectedGroups.value = []
  }
}

watch(
  groups,
  (newGroups, oldGroups) => {
    if (!newGroups || newGroups.length <= 0) {
      selectedGroups.value = []
    }
    if (!oldGroups || JSON.stringify(newGroups) !== JSON.stringify(oldGroups)) {
      selectAll()
    }
  },
  { immediate: true },
)
</script>

<template>
  <BcMultiSelect
    v-model="selectedGroups"
    :options="groups"
    option-label="name"
    option-value="id"
    :placeholder="$t('dashboard.group.selection.all')"
    class="slot-viz-group-selector"
  >
    <template #header>
      <span
        class="pointer"
        @click="toggleAll"
      >
        {{ $t("dashboard.group.selection.all") }}
      </span>
    </template>
    <template #value>
      {{ selectedLabel }}
    </template>
  </BcMultiSelect>
</template>

<style lang="scss" scoped>
  .slot-viz-group-selector {
    margin-left: auto;
    width: 120px;
    height: 46px;

    @media (min-width: 800px) {
      width: 196px;
      height: 30px;
    }
  }
</style>
