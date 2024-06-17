<script setup lang="ts">
import { faEye } from '@fortawesome/pro-solid-svg-icons'
import type { SlotVizEpoch } from '~/types/api/slot_viz'
import { type SlotVizCategories } from '~/types/dashboard/slotViz'
import { formatNumber } from '~/utils/format'
import { IconSlotAttestation, IconSlotBlockProposal, IconSlotSlashing, IconSlotSync } from '#components'
import { COOKIE_KEY } from '~/types/cookie'
import { useNetworkStore } from '~/stores/useNetworkStore'
import type { MultiBarItem } from '~/types/multiBar'

interface Props {
  data: SlotVizEpoch[],
  initiallyHideVisible?: boolean,
  timestamp?: number
}
const props = defineProps<Props>()

const { tsToSlot } = useNetworkStore()
const { t: $t } = useI18n()

const selectedCategories = useCookie<SlotVizCategories[]>(COOKIE_KEY.SLOT_VIZ_SELECTED_CATEGORIES, { default: () => ['attestation', 'proposal', 'slashing', 'sync', 'visible', 'initial'] })

const icons: MultiBarItem[] = [
  {
    component: IconSlotBlockProposal,
    value: 'proposal',
    tooltip: $t('slotViz.filter.proposal')
  }, {
    component: IconSlotAttestation,
    value: 'attestation',
    tooltip: $t('slotViz.filter.attestation')
  }, {
    component: IconSlotSync,
    value: 'sync',
    tooltip: $t('slotViz.filter.sync')
  }, {
    component: IconSlotSlashing,
    value: 'slashing',
    tooltip: $t('slotViz.filter.slashing')
  }, {
    icon: faEye,
    value: 'visible',
    className: 'visible-icon',
    tooltip: $t('slotViz.filter.visible')
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
  // in case of some backend issues Inan want's us to tick in the future ... so let's tick
  return Math.max(mostRecentScheduledSlotId.value ?? 0, tsToSlot((props.timestamp ?? 0) / 1000) - 1)
})

watch(props, () => {
  if (props.initiallyHideVisible !== undefined) {
    const initialIndex = selectedCategories.value.indexOf('initial')
    if (initialIndex < 0) {
      return
    }
    const categories = selectedCategories.value
    if (props.initiallyHideVisible) {
      categories.splice(initialIndex, 1)
      const visibleIndex = categories.indexOf('visible')
      if (visibleIndex >= 0) {
        categories.splice(visibleIndex, 1)
      }
    } else {
      categories.splice(initialIndex, 1, 'visible')
    }

    selectedCategories.value = categories
  }
}, { immediate: true })

</script>
<template>
  <div id="slot-viz" class="content">
    <div class="rows">
      <div class="row" />
      <div v-for="row in props.data" :key="row.epoch" class="row">
        <div class="epoch">
          <BcFormatNumber :text="row.state === 'head' ? $t('slotViz.head') : formatNumber(row.epoch)" />
        </div>
      </div>
    </div>
    <div class="rows">
      <div class="row">
        <BcToggleMultiBar v-model="selectedCategories" :icons="icons" />
      </div>
      <div v-for="row in props.data" :key="row.epoch" class="row">
        <SlotVizTile
          v-for="slot in row.slots"
          :key="slot.slot"
          :data="slot"
          :selected-categories="selectedCategories"
          :current-slot-id="currentSlotId"
        />
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

      &:first-child {
        height: 46px;
      }
    }
  }

  :deep(.visible-icon) {
    margin-left: 4px;
    overflow: visible;
    position: relative;
  }

  :deep(.visible-icon):before {
    content: ' ';
    background-color: var(--container-border-color);
    height: 100%;
    width: 1px;
    position: absolute;
    left: -8px;
  }
}
</style>
