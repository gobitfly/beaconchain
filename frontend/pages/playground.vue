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

const tabs: HashTabs = [
  {
    component: PlaygroundComponents,
    key: 'components',
    title: 'Components',
  },
  {
    component: PlaygroundStyling,
    key: 'styling',
    title: 'Styling',
  },
  {
    component: PlaygroundComposable,
    key: 'composables',
    title: 'Composables',
  },
  {
    key: 'slotviz',
    title: 'Slot Viz',
  },
  {
    component: PlaygroundAds,
    key: 'addSafe',
    title: 'Ads',
  },
  {
    component: PlaygroundSubsetList,
    key: 'subset',
    title: 'Subset Validators',
  },
  {
    component: PlaygroundDashboardValidatorManageValidators,
    key: 'manage',
    title: 'Manage Validators',
  },
]
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
