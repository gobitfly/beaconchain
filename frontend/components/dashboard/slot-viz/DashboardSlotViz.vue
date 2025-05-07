<script setup lang="ts">
import { useValidatorSlotVizStore } from '~/stores/dashboard/useValidatorSlotVizStore'
import type { SlotVizCategories } from '~/types/dashboard/slotViz'

const {
  dashboardKey,
} = useDashboardKey()
const { networkInfo } = useNetwork()
const {
  loading: loadingSlotViz,
  refreshSlotViz,
  slotViz,
} = useValidatorSlotVizStore()
const { secondsPerSlot = 12 } = networkInfo.value
const {
  counter,
  reset: resetIntervalCounter,
} = useInterval(secondsPerSlot)
const { getSlotFromTimestamp } = useNetwork()
const validatorDashboardOverviewStore = useValidatorDashboardOverviewStore()
const {
  loading: loadingOverview,
  overview,
} = storeToRefs(validatorDashboardOverviewStore)

const selectedCategories = ref<SlotVizCategories[]>([])
const selectedGroupIds = ref<number[]>([])
const refetchingSlotViz = ref(false)

const activeValidatorGroups = computed(() =>
  overview.value?.groups.filter(group => !!group.count) || [],
)
const mostRecentScheduledSlotId = computed(() => {
  if (!slotViz.value?.length) {
    return
  }
  let id = -1

  for (let i = 0; i < slotViz.value.length; i++) {
    const row = slotViz.value[i]
    if (!row.slots?.length) {
      continue
    }
    for (let j = row.slots.length - 1; j >= 0; j--) {
      if (row.slots[j].status === 'scheduled') {
        id = row.slots[j].slot
      }
      else {
        return id
      }
    }
  }
  return id
})
const currentSlotId = computed(() => {
  // in case of some backend issues Inan want's us to counter in the future ... so let's counter
  return Math.max(
    mostRecentScheduledSlotId.value ?? 0,
    getSlotFromTimestamp((counter.value ?? 0) / 1000) - 1)
})

watch(
  () =>
    activeValidatorGroups.value,
  () => {
    selectedGroupIds.value = activeValidatorGroups.value.length > 1
      ? activeValidatorGroups.value.map(group => group.id)
      : []
  },
  { immediate: true },
)
watch(
  () => selectedGroupIds.value,
  () => {
    useAsyncData('validator_dashboard_slot_viz', () =>
      refreshSlotViz(dashboardKey.value, selectedGroupIds.value),
    )
    resetIntervalCounter()
  },
  { immediate: true },
)
watch(
  () => counter.value,
  async () => {
    refetchingSlotViz.value = true
    await refreshSlotViz(dashboardKey.value, selectedGroupIds.value)
    refetchingSlotViz.value = false
  },
)
</script>

<template>
  <section class="dashboard-slot-viz">
    <div class="dashboard-slot-viz-header">
      <BcTooltip
        class="dashboard-slot-viz-info"
        :text="$t('slot_viz.info_tootlip')"
        dont-open-permanently
      >
        <BcLink
          to="https://kb.beaconcha.in/v2beta/slot-visualization#how-does-it-work"
          target="_blank"
          class="link"
        >
          <BcIcon name="circle-info" />
        </BcLink>
      </BcTooltip>

      <DashboardSlotVizDutyVisibilityToggle
        class="dashboard-slot-viz-toggle"
        @update-categories="(categories) => selectedCategories = categories"
      />

      <BcText
        variant="lg"
        class="dashboard-slot-viz-heading"
      >
        {{ networkInfo?.name }}
      </BcText>

      <DashboardSlotVizGroupSelector
        v-if="activeValidatorGroups.length > 1"
        :validator-groups="activeValidatorGroups"
        class="dashboard-slot-viz-group-selector"
        @update-selected-group-ids="(newGroupIdSelection) => selectedGroupIds = newGroupIdSelection"
      />
    </div>

    <div
      v-if="(loadingSlotViz && !refetchingSlotViz) || loadingOverview"
      class="dashboard-slot-viz-grid-loading-skeleton"
    >
      <BcLoadingSpinner
        loading
        alignment="center"
      />
    </div>
    <div
      v-else
      class="dashboard-slot-viz-grid"
    >
      <template
        v-for="row in slotViz"
        :key="row.epoch"
      >
        <div class="dashboard-slot-viz-grid-epoch">
          <BcFormatNumber
            :text="
              row.state === 'head'
                ? $t('slot_viz.head')
                : formatNumber(`${row.epoch}`)
            "
          />
        </div>
        <div class="dashboard-slot-viz-grid-row">
          <DashboardSlotVizTile
            v-for="slot in row.slots"
            :key="slot.slot"
            :data="slot"
            :selected-categories
            :current-slot-id
          />
        </div>
      </template>
    </div>
  </section>
</template>

<style lang="scss" scoped>
@use "~/assets/css/main.scss";
@use "~/assets/css/fonts.scss";

.dashboard-slot-viz {
  @include main.container;
  position: relative;
  padding-top: var(--padding-large);
  padding-bottom: var(--padding);
  padding-left: var(--padding-small);
  padding-right: var(--padding);

  &-header {
    display: grid;
    justify-content: center;
    align-items: center;
    grid-template: [row1-start] "network network network" [row1-end] [row2-start] "info filter-row header-right" [row2-end] / max-content max-content 1fr;
    gap: var(--padding);
    column-gap: var(--padding-small);
    margin-bottom: var(--padding);

    @media (min-width: 800px) {
      grid-template: [row1-start] "info filter-row network header-right" [row1-end] / 4rem max-content 1fr 196px;
    }
  }

  &-info {
    display: flex;
    justify-content: center;
    align-items: center;
    margin-right: var(--padding-small);
  }

  &-toggle {
    grid-area: filter-row;
  }

  &-heading {
    grid-area: network;
    text-align: center;
  }

  &-group-selector {
    grid-area: header-right;
    margin-left: auto;
  }

  &-grid {
    display: grid;
    gap: var(--padding);
    overflow-x: auto;
    overflow-y: hidden;
    grid-template-columns: 3.75rem auto;
    padding-bottom: var(--padding);

    @media (max-width: 490px) {
      padding-right: var(--padding);
    }

    &-epoch {
      @include fonts.small_text_bold;
      margin-top: auto;
      margin-bottom: auto;
    }

    &-row {
      display: flex;
      align-items: center;
      justify-content: flex-start;
      height: 30px;
      gap: var(--padding);
      justify-self: center;
    }

    &-loading-skeleton {
      height: 165px
    }
  }
}
</style>
