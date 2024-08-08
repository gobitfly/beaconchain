<script setup lang="ts">
import { useLatestStateStore } from '~/stores/useLatestStateStore'
import {
  type InternalGetValidatorDashboardSlotVizResponse,
  type SlotVizEpoch,
} from '~/types/api/slot_viz'
import { formatNumber } from '~/utils/format'

const { dashboardKey } = useDashboardKeyProvider(undefined, '100')

useBcSeo()

const { latestState, refreshLatestState } = useLatestStateStore()
const slotVizData = ref<null | SlotVizEpoch[]>(null)
const { refreshOverview } = useValidatorDashboardOverviewStore()

await Promise.all([
  useAsyncData('latest_state', () => refreshLatestState()),
  useAsyncData('test_slot_viz_data', async () => {
    const res = await $fetch<InternalGetValidatorDashboardSlotVizResponse>(
      './mock/dashboard/slotViz.json',
    )
    slotVizData.value = res.data
  }),
  useAsyncData('validator_dashboard_overview', () =>
    refreshOverview(dashboardKey.value),
  ),
])

onMounted(async () => {
  const res = await $fetch<InternalGetValidatorDashboardSlotVizResponse>(
    './mock/dashboard/slotViz.json',
  )
  slotVizData.value = res.data
})
</script>

<template>
  <div class="content">
    <h1>Playground for testing UI components</h1>
    <BcLink
      to="/"
      class="row"
    >
      <Button class="row">
        <IconBeaconchainLogo alt="Beaconcha.in logo" />
      </Button>
    </BcLink>
    <div class="row">
      Latest Slot: {{ formatNumber(latestState?.current_slot) }}
    </div>

    <TabView :lazy="true">
      <TabPanel header="Components">
        <PlaygroundComponents />
      </TabPanel>
      <TabPanel header="Styling">
        <PlaygroundStyling />
      </TabPanel>
      <TabPanel header="Composable">
        <PlaygroundComposable />
      </TabPanel>
      <TabPanel header="Ads">
        <PlaygroundAds />
      </TabPanel>
      <TabPanel header="Slot Viz">
        <SlotVizViewer
          v-if="slotVizData"
          :data="slotVizData"
        />
      </TabPanel>
      <TabPanel header="Subset Validators">
        <PlaygroundSubsetList />
      </TabPanel>
      <TabPanel header="Manage Validators">
        <PlaygroundDashboardValidatorManageValidators />
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
