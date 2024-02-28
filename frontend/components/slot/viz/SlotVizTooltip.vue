<script setup lang="ts">
import type { VDBSlotVizActiveDuty, VDBSlotVizPassiveDuty, VDBSlotVizSlot } from '~/types/api/slot_viz'
import { type SlotVizIcons } from '~/types/dashboard/slotViz'
import { type TooltipLayout } from '~/types/layouts'
import { formatNumber } from '~/utils/format'
type Row = { count?: number; icon: SlotVizIcons; class?: string; change?: string; dutyText?: string, validator?: number; dutySubText?: string; dutySubLink?: string, duty_object?: number}
interface Props {
  id: string
  data: VDBSlotVizSlot
}
const props = defineProps<Props>()
const { t: $t } = useI18n()

const data = computed(() => {
  const slot = props.data
  const rows: Row[][] = []

  const hasDuties = !!slot?.proposal || !!slot?.slashing?.length || !!slot?.attestations || !!slot?.sync
  const tooltipLayout: TooltipLayout = hasDuties ? 'dark' : 'default'
  if (hasDuties) {
    const addActiveDuty = (type: SlotVizIcons, duty: VDBSlotVizActiveDuty) => {
      const subRows: Row[] = []
      rows.push(subRows)
      const dutyText = $t(`slotViz.tooltip.${type}.${duty.status}.main`)
      const dutySubText = $t(`slotViz.tooltip.${type}.${duty.status}.sub`)
      let dutySubLink = ''
      if (type === 'proposal') {
        if (duty.status === 'success') {
          dutySubLink = `/block/${duty.duty_object}`
        } else {
          dutySubLink = `/slot/${duty.duty_object}`
        }
      } else if (type === 'slashing') {
        dutySubLink = `/validator/${duty.duty_object}`
      }

      subRows.push({ class: duty.status, icon: type, dutyText, count: 1, dutySubText, validator: duty.validator, dutySubLink, duty_object: duty.duty_object })
    }

    slot.proposal && addActiveDuty('proposal', slot.proposal)
    slot.slashing?.forEach(duty => addActiveDuty('slashing', duty))

    const addPassiveDuty = (type: SlotVizIcons, duty?: VDBSlotVizPassiveDuty) => {
      if (duty) {
        const subRows: Row[] = []
        rows.push(subRows)
        const dutyText = $t(`slotViz.tooltip.${type}`)
        if (duty.pending_count) {
          subRows.push({ class: 'scheduled', icon: type, count: duty.pending_count, dutyText })
        }
        if (duty.success_count) {
          subRows.push({ class: 'success', icon: type, count: duty.success_count, dutyText })
        }
        if (duty.failed_count) {
          subRows.push({ class: 'failed', icon: type, count: duty.failed_count, dutyText })
        }
      }
    }
    addPassiveDuty('attestation', slot.attestations)
    addPassiveDuty('sync', slot.sync)
  }
  const stateLabel = $t(`slot_state.${slot.status}`)
  const slotLabel = `${stateLabel} ${$t('common.slot')} ${formatNumber(slot.slot)}`

  return {
    slotLabel,
    tooltipLayout,
    rows,
    hasDuties
  }
})
</script>
<template>
  <BcTooltip :target="props.id" :layout="data.tooltipLayout">
    <slot />
    <template v-if="data.hasDuties" #tooltip>
      <div class="with-duties">
        <div v-for="(rows, index) in data.rows" :key="index" class="rows">
          <div v-for="row in rows" :key="row.class" class="row" :class="row.class">
            <div class="count-icon">
              <span>{{ row.count }}x</span>
              <SlotVizIcon :icon="row.icon" class="icon" />
            </div>
            <div class="value-col">
              {{ row.dutyText }}
              <div v-if="row.validator">
                <NuxtLink :to="`/validator/${row.validator}`" target="_blank" class="link">
                  {{ row.validator }}
                </NuxtLink>
                <span class="sub-text"> {{ row.dutySubText }} </span>
                <NuxtLink v-if="row.dutySubLink" :to="row.dutySubLink" target="_blank" class="link">
                  {{ row.duty_object }}
                </NuxtLink>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>
  </BcTooltip>
</template>

<style lang="scss" scoped>
.with-duties {
  font-size: var(--paragraph_4_font_size);
  font-family: var(--roboto-family);

  .rows {
    padding-bottom: var(--padding);
    padding-top: var(--padding);

    &:not(:first-child) {
      border-top: 1px solid var(--container-border-color);
    }

    .row {
      color: var(--light-grey);
      display: flex;
      align-items: center;

      &:not(:first-child) {
        padding-top: var(--padding);
      }

      &.success {
        color: var(--green);
      }

      &.failed {
        color: var(--flashy-red);
      }

      .count-icon{
        display: inline-flex;
        width: 90px;
        justify-content: end;
        align-items: center;
      }
      .sub-text {
        color: var(--light-grey);
        padding: 0 3px;
      }

      .icon {
        margin-left: 6px;
        margin-right: 20px;
      }

      .value-col {
        display: flex;
        flex-direction: column;
        align-items: flex-start;
        word-wrap: nowrap;
        white-space: nowrap;
      }
    }
  }
}
</style>
