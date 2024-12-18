<script setup lang="ts">
import { orderBy } from 'lodash-es'
import type { SlotVizEpoch } from '~/types/api/slot_viz'
import { getGroupLabel } from '~/utils/dashboard/group'

const { t: $t } = useTranslation()
const {
  groups: dashboardGroups, validatorCount,
} = storeToRefs(useValidatorDashboardStore())
const { networkInfo } = useNetworkStore()
const selectedGroups = ref<number[]>([])

const {
  tick,
} = useInterval(12)

const props = defineProps<{
  data?: SlotVizEpoch[],
}>()
const { data } = toRefs(props)

const emit = defineEmits<{
  (e: 'update', groupIds: number[]): void,
}>()

watch(
  [
    selectedGroups,
    tick,
  ],
  ([
    newSelectedGroups,
    newTick,
  ], [
    oldSelectedGroups,
    oldTick,
  ]) => {
    if (oldTick === newTick && isAllSelected(newSelectedGroups) && isAllSelected(oldSelectedGroups)) {
      // when toggleAll is called or dashboard groups are updated, don't emit redundantly
      return
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
  if (!dashboardGroups.value) {
    return []
  }
  return orderBy(
    dashboardGroups.value.filter(g => !!g.count),
    [ g => g.name.toLowerCase() ],
    'asc',
  )
})

const isAllSelected = (groupList: number[]) => {
  return groupList.length === groups.value.length || groupList.length === 0
}

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
    v-if="data"
    :data
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
