<script setup lang="ts">

import { useValidatorSlotVizStore } from '~/stores/dashboard/useValidatorSlotVizStore'

interface Props {
  dashboardId: number
}
const props = defineProps<Props>()
const { tick } = useInterval(12000)

const store = useValidatorSlotVizStore()

const { getSlotViz } = store
const { slotViz } = storeToRefs(store)
await useAsyncData('validator_dashboard_slot_viz', () => getSlotViz(props.dashboardId))

watch(() => [props.dashboardId, tick.value], () => {
  getSlotViz(props.dashboardId)
}, { immediate: true })

</script>
<template>
  <SlotVizViewer v-if="slotViz" :data="slotViz" />
</template>

<style lang="scss" scoped>
</style>
