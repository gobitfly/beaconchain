<script setup lang="ts">

import { useValidatorSlotVizStore } from '~/stores/dashboard/useValidatorSlotVizStore'

// TODO: implement dashboard switching
const { getSlotViz } = useValidatorSlotVizStore()
await useAsyncData('validator_dashboard_slot_viz', () => getSlotViz())

const { slotViz } = storeToRefs(useValidatorSlotVizStore())

</script>
<template>
  <SlotVizViewer v-if="slotViz" :data="slotViz" />
</template>

<style lang="scss" scoped>
</style>
