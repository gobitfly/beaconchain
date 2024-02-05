<script setup lang="ts">
import { type SlotVizSlot, type SlotVizIcons } from '~/types/dashboard/slotViz'
import { type TooltipLayout } from '~/types/layouts'
import { formatNumber } from '~/utils/format'
type Row = { count: number; icon: SlotVizIcons; class: string; change?: string; validator?: number; }
interface Props {
  id: string
  data: SlotVizSlot
}
const props = defineProps<Props>()
const { t: $t } = useI18n()

const data = computed(() => {
  const slot = props.data
  const rows: Row[][] = []

  const hasDuties = !!slot.duties?.length
  const tooltipLayout: TooltipLayout = hasDuties ? 'dark' : 'default'
  if (hasDuties) {
    const addDuty = (type: SlotVizIcons) => {
      const duty = slot.duties?.find(s => s.type === type)
      if (duty) {
        const subRows: Row[] = []
        rows.push(subRows)
        if (duty.pendingCount) {
          subRows.push({ class: 'pending', icon: type, validator: duty.validator, count: duty.pendingCount })
        }
        if (duty.successCount) {
          subRows.push({ class: 'success', icon: type, validator: duty.validator, count: duty.successCount, change: duty.successEarning })
        }
        if (duty.failedCount) {
          subRows.push({ class: 'failed', icon: type, validator: duty.validator, count: duty.failedCount, change: duty.failedEarnings })
        }
      }
    }
    const types: SlotVizIcons[] = ['proposal', 'slashing', 'sync', 'attestation']
    types.forEach(type => addDuty(type))
  }
  const stateLabel = $t(`slot_state.${slot.state}`)
  const slotLabel = `${stateLabel} ${$t('common.slot')} ${formatNumber(slot.id)}`

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
    <template #tooltip>
      <div v-if="!data.hasDuties">
        {{ data.slotLabel }}
      </div>
      <div class="with-duties">
        <div v-for="(rows, index) in data.rows" :key="index" class="rows">
          <div v-for="row in rows" :key="row.class" class="row" :class="row.class">
            <span>{{ row.count }}x</span>
            <SlotVizIcon :icon="row.icon" class="icon" />
            <div class="value-col">
              <BcFormatValue v-if="row.change" :value="row.change" :options="{addPlus: true}" />
              <div v-if="row.validator">
                {{ $t('common.validator') }}
                <NuxtLink :to="`/validator/${ row.validator }`" class="link">
                  {{ row.validator }}
                </NuxtLink>
              </div>
              <div v-else-if="row.class === 'pending'">
                {{ $t('validator_state.pending') }}
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
