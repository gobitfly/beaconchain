<script setup lang="ts">
import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import type { SlotVizCategories } from '~/types/dashboard/slotViz'
import type { SlotVizEpoch } from '~/types/api/slot_viz'
import type { VDBOverviewData } from '~/types/api/validator_dashboard'

const {
  epochsData, isLoading, overviewData, refetchingSlotVizData, timestamp,
} = defineProps<{
  epochsData: SlotVizEpoch[],
  isLoading: boolean,
  overviewData: VDBOverviewData,
  refetchingSlotVizData: boolean,
  timestamp: number,
}>()

const { networkInfo } = useNetworkStore()
const { getSlotFromTimestamp } = useNetworkStore()

const selectedCategories = ref<SlotVizCategories[]>([])

const selectedGroups = defineModel<number[]>('selectedGroups', {
  default: [],
})

const mostRecentScheduledSlotId = computed(() => {
  if (!epochsData || !epochsData.length) {
    return
  }
  let id = -1

  for (let i = 0; i < epochsData.length; i++) {
    const row = epochsData[i]
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
  // in case of some backend issues Inan want's us to tick in the future ... so let's tick
  return Math.max(
    mostRecentScheduledSlotId.value ?? 0,
    getSlotFromTimestamp((timestamp ?? 0) / 1000) - 1)
})

const activeValidatorGroups = computed(() =>
  overviewData.groups.filter(group => !!group.count) || [],
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
          <FontAwesomeIcon :icon="faInfoCircle" />
        </BcLink>
      </BcTooltip>

      <DashboardSlotVizDutyVisibilityToggle
        class="dashboard-slot-viz-toggle"
        :overview-data
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
        v-model:selected-groups="selectedGroups"
        :validator-groups="activeValidatorGroups"
        class="dashboard-slot-viz-group-selector"
      />
    </div>

    <div
      class="dashboard-slot-viz-grid"
    >
      <BcLoadingSpinner
        v-if="isLoading && !refetchingSlotVizData"
        class="dashboard-slot-viz-grid-loading-spinner"
        loading
        has-backdrop
        alignment="center"
      />

      <template
        v-for="row in epochsData"
        :key="row.epoch"
      >
        <div class="dashboard-slot-viz-grid-epoch">
          <BcFormatNumber
            :text="
              row.state === 'head'
                ? $t('slot_viz.head')
                : formatNumber(row.epoch)
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
    position: relative;

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

    &-loading-spinner {
        position: absolute;
        width: 100%;
        height: 100%;
    }
  }
}
</style>
