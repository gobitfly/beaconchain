<script setup lang="ts">
import { faEye } from '@fortawesome/pro-solid-svg-icons'
import {
  faInfoCircle
} from '@fortawesome/pro-regular-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import type { SlotVizEpoch } from '~/types/api/slot_viz'
import { type SlotVizCategories } from '~/types/dashboard/slotViz'
import { formatNumber } from '~/utils/format'
import { IconSlotAttestation, IconSlotBlockProposal, IconSlotSlashing, IconSlotSync } from '#components'
import { COOKIE_KEY } from '~/types/cookie'
import { useNetworkStore } from '~/stores/useNetworkStore'
import type { MultiBarItem } from '~/types/multiBar'
import type { ChainInfoFields } from '~/types/network'

interface Props {
  data: SlotVizEpoch[],
  initiallyHideVisible?: boolean,
  timestamp?: number
  networkInfo?: ChainInfoFields
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

watch(() => props, () => {
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
    <div class="header-row">
      <BcTooltip class="info" :text="$t('slotViz.info_tootlip')" :dont-open-permanently="true">
        <BcLink to="https://kb.beaconcha.in/v2beta/slot-visualization#how-does-it-work" target="_blank" class="link">
          <FontAwesomeIcon :icon="faInfoCircle" />
        </BcLink>
      </BCTooltip>
      <div class="filter-row">
        <BcToggleMultiBar v-model="selectedCategories" :buttons="icons" />
      </div>
      <h1 class="network">
        {{ networkInfo?.name }}
      </h1>
      <div class="header-right">
        <slot name="header-right" />
      </div>
    </div>
    <div class="grid">
      <template v-for="row in props.data" :key="row.epoch">
        <div class="epoch">
          <BcFormatNumber :text="row.state === 'head' ? $t('slotViz.head') : formatNumber(row.epoch)" />
        </div>
        <div class="row">
          <SlotVizTile
            v-for="slot in row.slots"
            :key="slot.slot"
            :data="slot"
            :selected-categories="selectedCategories"
            :current-slot-id="currentSlotId"
          />
        </div>
      </template>
    </div>
  </div>
</template>
<style lang="scss" scoped>
@use '~/assets/css/main.scss';
@use '~/assets/css/fonts.scss';

.content {
  position: relative;
  @include main.container;

  .header-row {
    display: grid;
    justify-content: center;
    padding: var(--padding-large) var(--padding-large) var(--padding) 9px;
    gap: var(--padding);
    grid-template:
      [row1-start] "info filter-row network header-right"[row1-end] / max-content max-content 1fr max-content;

    @media (max-width: 800px) {
      column-gap: var(--padding-small);
      grid-template:
        [row1-start] "network network network"[row1-end] [row2-start] "info filter-row header-right"[row2-end] / max-content max-content 1fr;
    }

    @media (max-width: 490px) {
      padding-right: var(--padding);
    }

    .network {
      grid-area: network;
      flex-grow: 1;
      text-align: center;
      margin-top: auto;
      margin-bottom: auto;
      justify-self: stretch;
    }

    .info {
      grid-area: info;
      display: flex;
      justify-content: center;
      align-items: center;
      width: 49px;

      @media (max-width: 800px) {
        width: 20px;
      }
    }

    .header-right {
      grid-area: header-right;
      width: 196px;
      margin-top: auto;
      margin-bottom: auto;
      margin-left: auto;

      @media (max-width: 490px) {
        width: 104px;
      }
    }

    .filter-row {
      grid-area: filter-row;
      display: flex;
      align-items: center;
    }
  }

  .epoch {
    @include fonts.small_text_bold;
    margin-top: auto;
    margin-bottom: auto;
  }

  .grid {
    padding: 0 var(--padding-large) var(--padding-large) 9px;
    display: grid;
    gap: var(--padding);
    overflow-x: auto;
    overflow-y: hidden;
    min-height: 180px;
    grid-template-columns: 49px max-content;

    @media (max-width: 490px) {
      padding-right: var(--padding);
    }

    .row {
      display: flex;
      align-items: center;
      justify-content: flex-start;
      height: 30px;
      gap: var(--padding);
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
    left: -5px;
  }
}
</style>
