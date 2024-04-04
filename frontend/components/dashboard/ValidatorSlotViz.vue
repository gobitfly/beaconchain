<script setup lang="ts">

import { useValidatorSlotVizStore } from '~/stores/dashboard/useValidatorSlotVizStore'

const { dashboardKey } = useDashboardKey()

const { tick } = useInterval(12)

const { slotViz, refreshSlotViz } = useValidatorSlotVizStore()
await useAsyncData('validator_dashboard_slot_viz', () => refreshSlotViz(dashboardKey.value))

watch(() => [dashboardKey.value, tick.value], () => {
  refreshSlotViz(dashboardKey.value)
}, { immediate: true })

</script>
<template>
  <SlotVizViewer v-if="slotViz" :data="slotViz" />
</template>

<style lang="scss" scoped>
</style>
