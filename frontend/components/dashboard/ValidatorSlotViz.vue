<script setup lang="ts">

import { useValidatorSlotVizStore } from '~/stores/dashboard/useValidatorSlotVizStore'

const { dashboardKey } = useDashboardKey()
const { validatorCount } = useValidatorDashboardOverviewStore()
const { networkInfo } = useNetworkStore()

const { tick, resetTick } = useInterval(12)

const { slotViz, refreshSlotViz } = useValidatorSlotVizStore()
await useAsyncData('validator_dashboard_slot_viz', () => refreshSlotViz(dashboardKey.value))

watch(() => [dashboardKey.value, tick.value], (newValue, oldValue) => {
  if (oldValue && newValue[0] !== oldValue[0]) {
    resetTick()
  }
  refreshSlotViz(dashboardKey.value)
}, { immediate: true })

const initiallyHideVisible = computed(() => {
  if (validatorCount.value === undefined) {
    return undefined
  }
  return validatorCount.value > 60
})
</script>

<template>
  <SlotVizViewer v-if="slotViz" :data="slotViz" :network-info="networkInfo" :timestamp="tick" :initially-hide-visible="initiallyHideVisible" />
</template>

<style lang="scss" scoped>
</style>
