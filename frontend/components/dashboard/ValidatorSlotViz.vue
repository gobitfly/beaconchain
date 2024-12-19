<script setup lang="ts">
import { orderBy } from 'lodash-es'
import type { SlotVizEpoch } from '~/types/api/slot_viz'
import { getGroupLabel } from '~/utils/dashboard/group'

const { t: $t } = useTranslation()
const {
  groups: dashboardGroups, validatorCount,
} = storeToRefs(useValidatorDashboardStore())
const { networkInfo } = useNetworkStore()

const { data } = defineProps<{
  data: SlotVizEpoch[],
}>()

const emit = defineEmits<{
  (e: 'update', groupIds: number[]): void,
}>()

const {
  resetTick, tick,
} = useInterval(12)
watch(
  tick,
  () => {
    emit('update', selectedGroups.value)
  },
)

const isAllSelected = (groupList: number[]) => {
  return groupList.length === groups.value.length || groupList.length === 0
}
const selectedGroups = ref<number[]>([])
watch (
  selectedGroups,
  (newSelectedGroups, oldSelectedGroups) => {
    if (isAllSelected(newSelectedGroups) && isAllSelected(oldSelectedGroups)) {
      // avoids redundant emit on toggleAll or dashboard switch
      return
    }
    emit('update', newSelectedGroups)
    resetTick()
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
