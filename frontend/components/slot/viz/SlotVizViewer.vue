<script setup lang="ts">
import type { SlotVizEpoch } from '~/types/api/slot_viz'
import { type SlotVizCategories } from '~/types/dashboard/slotViz'
import { formatNumber } from '~/utils/format'
import { IconSlotAttestation, IconSlotBlockProposal, IconSlotSlashing, IconSlotSync } from '#components'
interface Props {
  data: SlotVizEpoch[]
}
const props = defineProps<Props>()

const selectedCategoris = shallowRef<SlotVizCategories[]>(['attestation', 'proposal', 'slashing', 'sync'])

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

</script>
<template>
  <div id="slot-viz" class="content">
    <div class="rows">
      <div class="row" />
      <div v-for="row in props.data" :key="row.epoch" class="row">
        <div class="epoch">
          {{ row.state === 'head' ? $t('slotViz.head') : formatNumber(row.epoch) }}
        </div>
      </div>
    </div>
    <div class="rows">
      <div class="row">
        <BcToggleMultiBar v-model="selectedCategoris" :icons="icons" />
      </div>
      <div v-for="row in props.data" :key="row.epoch" class="row">
        <SlotVizTile v-for="slot in row.slots" :key="slot.slot" :data="slot" :selected-categoris="selectedCategoris" />
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
