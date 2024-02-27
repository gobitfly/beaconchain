<script setup lang="ts">
import { useLatestStateStore } from '~/stores/useLatestStateStore'
import { type SlotVizData } from '~/types/dashboard/slotViz'
import { formatNumber } from '~/utils/format'
const { getLatestState } = useLatestStateStore()
await useAsyncData('latest_state', () => getLatestState())

const { latest } = storeToRefs(useLatestStateStore())

const slotVizData = ref<SlotVizData | null>(null)

await useAsyncData('test_slot_viz_data', async () => {
  const res = await $fetch<SlotVizData>('./mock/dashboard/slotViz.json')
  slotVizData.value = res
})

onMounted(async () => {
  const res = await $fetch<SlotVizData>('./mock/dashboard/slotViz.json')
  slotVizData.value = res
})

</script>

<template>
  <div class="content">
    <h1>Playground for testing UI components</h1>
    <NuxtLink to="/" class="row">
      <Button class="row">
        <IconBeaconchainLogo alt="Beaconcha.in logo" />
      </Button>
    </NuxtLink>
    <div class="row">
      Latest Epoch: {{ formatNumber(latest?.currentEpoch) }}
    </div>

    <TabView>
      <TabPanel header="Chart">
        <PlaygroundCharts />
      </TabPanel>
      <TabPanel header="Icons">
        <PlaygroundIcons />
      </TabPanel>
      <TabPanel header="Conversion">
        <PlaygroundConversion />
      </TabPanel>
      <TabPanel header="Styling">
        <PlaygroundStyling />
      </TabPanel>
      <TabPanel header="Ads">
        <PlaygroundAds />
      </TabPanel>
      <TabPanel header="Slot Viz">
        <SlotVizViewer v-if="slotVizData" :data="slotVizData" />
      </TabPanel>
    </TabView>

    <BcFooterMainFooter />
  </div>
</template>

<style lang="scss" scoped>
.content {
  padding: var(--padding-large);
}

.row {
  margin-bottom: var(--padding);
}
</style>
