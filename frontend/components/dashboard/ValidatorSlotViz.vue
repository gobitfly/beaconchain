<script setup lang="ts">
import { orderBy } from 'lodash-es'
import type { SlotVizEpoch } from '~/types/api/slot_viz'
import { getGroupLabel } from '~/utils/dashboard/group'

const { t: $t } = useTranslation()
// TODO: REMOVE THIS
const {
  overview, validatorCount,
} = useValidatorDashboardOverviewStore()
const { networkInfo } = useNetworkStore()
const selectedGroups = ref<number[]>([])

const {
  resetTick, tick,
} = useInterval(12)

const {
  slotVizData,
} = defineProps<{
  slotVizData?: SlotVizEpoch[],
}>()

const emit = defineEmits<{
  (e: 'update', groupIds?: number[]): void,
}>()

watch(
  () => [
    selectedGroups.value,
    tick.value,
  ],
  (newValue, oldValue) => {
    if (
      oldValue
      && (newValue[0] !== oldValue[0]
      || (newValue[1] as number[]).length !== (oldValue[1] as number[]).length)
    ) {
      resetTick()
    }
    emit('update', selectedGroups.value)
  },
)

const initiallyHideVisible = computed(() => {
  if (validatorCount.value === undefined) {
    return undefined
  }
  return validatorCount.value > 60
})

const groups = computed(() => {
  if (!overview.value?.groups) {
    return []
  }
  return orderBy(
    overview.value.groups.filter(g => !!g.count),
    [ g => g.name.toLowerCase() ],
    'asc',
  )
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
)

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
</script>

<template>
  <SlotVizViewer
    v-if="slotVizData"
    :data="slotVizData"
    :network-info
    :timestamp="tick"
    :initially-hide-visible
  >
    <template #header-right>
      <MultiSelect
        v-if="groups.length > 1"
        v-model="selectedGroups"
        :options="groups"
        option-label="name"
        option-value="id"
        :placeholder="$t('dashboard.group.selection.all')"
        class="group-selection"
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
      </MultiSelect>
    </template>
  </SlotVizViewer>
</template>

<style lang="scss" scoped>
@media (max-width: 800px) {
  .group-selection {
    height: 46px;
  }
}
</style>
