<script setup lang="ts">
import {
  PlaygroundAds,
  PlaygroundComponents,
  PlaygroundComposable,
  PlaygroundDashboardValidatorManageValidators,
  PlaygroundStyling,
  PlaygroundSubsetList,
} from '#components'
import { useLatestStateStore } from '~/stores/useLatestStateStore'
import {
  type InternalGetValidatorDashboardSlotVizResponse,
  type SlotVizEpoch,
} from '~/types/api/slot_viz'
import { formatNumber } from '~/utils/format'
import type { HashTabs } from '~/types/hashTabs'

const { dashboardKey } = useDashboardKeyProvider(undefined, '100')

useBcSeo()

const {
  latestState, refreshLatestState,
} = useLatestStateStore()
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

const tabs: HashTabs = {
  addSafe: {
    component: PlaygroundAds,
    index: '3',
    title: 'Ads',
  },
  components: {
    component: PlaygroundComponents,
    index: '0',
    title: 'Components',
  },
  composables: {
    component: PlaygroundComposable,
    index: '2',
    title: 'Composables',
  },
  manage: {
    component: PlaygroundDashboardValidatorManageValidators,
    index: '6',
    title: 'Manage Validators',
  },
  slotviz: {
    index: '4',
    title: 'Slot Viz',
  },
  styling: {
    component: PlaygroundStyling,
    index: '1',
    title: 'Styling',
  },
  subset: {
    component: PlaygroundSubsetList,
    index: '5',
    title: 'Subset Validators',
  },
}
</script>

<template>
  <div class="content">
    <h1>Playground for testing UI components</h1>
    <BcLink to="/" class="row">
      <Button class="row">
        <IconBeaconchainLogo alt="Beaconcha.in logo" />
      </Button>
    </BcLink>
    <div class="row">
      Latest Slot: {{ formatNumber(latestState?.current_slot) }}
    </div>

    <BcTabList :tabs default-tab="components" :use-route-hash="true">
      <template #tab-panel-slotviz>
        <SlotVizViewer v-if="slotVizData" :data="slotVizData" />
      </template>
    </BcTabList>
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
