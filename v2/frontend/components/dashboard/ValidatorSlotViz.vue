<script setup lang="ts">

import { useValidatorSlotVizStore } from '~/stores/dashboard/useValidatorSlotVizStore'
import type { DashboardKey } from '~/types/dashboard'

interface Props {
  dashboardKey: DashboardKey
}
const props = defineProps<Props>()
const { tick } = useInterval(12000)

const { slotViz, refreshSlotViz } = useValidatorSlotVizStore()
await useAsyncData('validator_dashboard_slot_viz', () => refreshSlotViz(props.dashboardKey))

watch(() => [props.dashboardKey, tick.value], () => {
  refreshSlotViz(props.dashboardKey)
}, { immediate: true })

</script>
<template>
  <SlotVizViewer v-if="slotViz" :data="slotViz" />
</template>

<style lang="scss" scoped>
</style>
