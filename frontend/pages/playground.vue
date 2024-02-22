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

// Validator Subset Modal
const modalVisibility = ref(false)
const modalValidators = ref([...Array(1000).fill(0).map((_, i) => i)])

onMounted(async () => {
  const res = await $fetch<SlotVizData>('./mock/dashboard/slotViz.json')
  slotVizData.value = res
})

</script>

<template>
  <div class="content">
    <DashboardValidatorSubsetModal
      v-model="modalVisibility"
      context="sync"
      time-frame="details_total"
      dashboard-name="Old Validators"
      group-name="Hetzner"
      :validators="modalValidators"
    />

    <h1>Playground for testing UI components</h1>
    <NuxtLink to="/" class="row">
      <Button class="row">
        <IconBeaconchainLogo alt="Beaconcha.in logo" />
      </Button>
    </NuxtLink>
    <div class="row">
      Latest Epoch: {{ formatNumber(latest?.currentEpoch) }}
    </div>

    <TabView :lazy="true">
      <TabPanel header="Validator Subset Modal">
        <Button label="Open Modal" @click="modalVisibility = true" />
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
      <TabPanel header="Summary">
        <PlaygroundDashboardValidatorSummary />
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
