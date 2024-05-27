<script setup lang="ts">
import type { SlotVizEpoch } from '~/types/api/slot_viz'
import { type SlotVizCategories } from '~/types/dashboard/slotViz'
import { formatNumber } from '~/utils/format'
import { IconSlotAttestation, IconSlotBlockProposal, IconSlotSlashing, IconSlotSync } from '#components'
import { COOKIE_KEY } from '~/types/cookie'

interface Props {
  data: SlotVizEpoch[]
}
const props = defineProps<Props>()
const { latestState } = useLatestStateStore()

const selectedCategories = useCookie<SlotVizCategories[]>(COOKIE_KEY.SLOT_VIZ_SELECTED_CATEGORIES, { default: () => ['attestation', 'proposal', 'slashing', 'sync'] })

const icons:{ component: Component, value: SlotVizCategories }[] = [
  {
    component: IconSlotBlockProposal,
    value: 'proposal'
  }, {
    component: IconSlotAttestation,
    value: 'attestation'
  }, {
    component: IconSlotSync,
    value: 'sync'
  }, {
    component: IconSlotSlashing,
    value: 'slashing'
  }
]

const mostRecentScheduledSlotId = computed(() => {
  if (!props.data?.length) {
    return
  }
  let id = -1

  for (let i = 0; i < props.data.length; i++) {
    const row = props.data[i]
    if (!row.slots?.length) {
      continue
    }
    for (let j = row.slots.length - 1; j >= 0; j--) {
      if (row.slots[j].status === 'scheduled') {
        id = row.slots[j].slot
      } else {
        return id
      }
    }
  }
})

const currentSlotId = computed(() => {
  return Math.max(mostRecentScheduledSlotId.value ?? 0, latestState.value?.current_slot ?? 0)
})

</script>
<template>
  <div id="slot-viz" class="content">
    <div class="rows">
      <div class="row" />
      <div v-for="row in props.data" :key="row.epoch" class="row">
        <div class="epoch">
          <BcFormatNumber :text="row.state === 'head' ? $t('slotViz.head') : formatNumber(row.epoch) " />
        </div>
      </div>
    </div>
    <div class="rows">
      <div class="row">
        <BcToggleMultiBar v-model="selectedCategories" :icons="icons" />
      </div>
      <div v-for="row in props.data" :key="row.epoch" class="row">
        <SlotVizTile v-for="slot in row.slots" :key="slot.slot" :data="slot" :selected-categories="selectedCategories" :current-slot-id="currentSlotId" />
      </div>
    </div>
  </div>
</template>
<style lang="scss" scoped>
@use '~/assets/css/main.scss';
@use '~/assets/css/fonts.scss';

.content {
  @include main.container;
  display: flex;
  gap: var(--padding);
  overflow-x: auto;
  overflow-y: hidden;
  min-height: 180px;
  padding: var(--padding-large) var(--padding-large) var(--padding-large) 9px;

  .epoch {
    @include fonts.small_text_bold;
  }

  .rows {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: var(--padding-large);

    .row {
      display: flex;
      align-items: center;
      justify-content: flex-start;
      height: 30px;
      gap: var(--padding);
      &:first-child{
        height: 46px;
      }
    }
  }
}
</style>
