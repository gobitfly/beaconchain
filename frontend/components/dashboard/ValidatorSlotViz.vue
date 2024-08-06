<script setup lang="ts">

import { orderBy } from 'lodash-es'
import { useValidatorSlotVizStore } from '~/stores/dashboard/useValidatorSlotVizStore'
import { getGroupLabel } from '~/utils/dashboard/group'

const { t: $t } = useTranslation()
const { dashboardKey } = useDashboardKey()
const { validatorCount, overview } = useValidatorDashboardOverviewStore()
const { networkInfo } = useNetworkStore()
const selectedGroups = ref<number[]>([])

const { tick, resetTick } = useInterval(12)

const { slotViz, refreshSlotViz } = useValidatorSlotVizStore()

await useAsyncData('validator_dashboard_slot_viz', () => refreshSlotViz(dashboardKey.value))

watch(() => [dashboardKey.value, selectedGroups.value, tick.value], (newValue, oldValue) => {
  if (oldValue && (newValue[0] !== oldValue[0] || (newValue[1] as number[]).length !== (oldValue[1] as number[]).length)) {
    resetTick()
  }
  refreshSlotViz(dashboardKey.value, selectedGroups.value)
}, { immediate: true })

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
  return orderBy(overview.value.groups.filter(g => !!g.count), [g => g.name.toLowerCase()], 'asc')
})

const selectAll = () => {
  selectedGroups.value = groups.value.map(g => g.id)
}

const toggleAll = () => {
  if (selectedGroups.value.length < groups.value.length) {
    selectAll()
  } else {
    selectedGroups.value = []
  }
}

watch(groups, (newGroups, oldGroups) => {
  if (!newGroups || newGroups.length <= 0) {
    selectedGroups.value = []
  }
  if (!oldGroups || JSON.stringify(newGroups) !== JSON.stringify(oldGroups)) {
    selectAll()
  }
}, { immediate: true })

const selectedLabel = computed(() => {
  if (selectedGroups.value.length === 0 || selectedGroups.value.length === groups.value.length) {
    return $t('dashboard.group.selection.all')
  }
  return orderBy(selectedGroups.value.map(id => getGroupLabel($t, id, groups.value)), [g => g.toLowerCase()], 'asc').join(', ')
})

</script>

<template>
  <SlotVizViewer
    v-if="slotViz"
    :data="slotViz"
    :network-info="networkInfo"
    :timestamp="tick"
    :initially-hide-visible="initiallyHideVisible"
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
          <span class="pointer" @click="toggleAll">
            {{ $t('dashboard.group.selection.all') }}
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
