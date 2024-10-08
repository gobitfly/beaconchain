<script setup lang="ts">
import type {
  VDBSlotVizDuty,
  VDBSlotVizSlot,
  VDBSlotVizStatus,
} from '~/types/api/slot_viz'
import { type SlotVizIcons } from '~/types/dashboard/slotViz'

type RowDuty = {
  duty_object?: number,
  dutySubLink?: string,
  dutySubText?: string,
  validator?: number,
}
type Row = {
  andMore?: number,
  change?: string,
  class?: string,
  count?: number,
  duties?: RowDuty[],
  dutyText?: string,
  icon: SlotVizIcons,
  validators?: number[],
}
interface Props {
  currentSlotId?: number,
  data: VDBSlotVizSlot,
  id: string,
}
const props = defineProps<Props>()
const { t: $t } = useTranslation()

const data = computed(() => {
  const slot = props.data
  const rows: Row[][] = []

  const status
    = slot.status === 'scheduled' && slot.slot < (props.currentSlotId ?? 0)
      ? 'scheduled_past'
      : slot.status

  const networkLabelPath = `slot_viz.tooltip.network.${status}`

  const hasDuties
    = !!slot?.proposal
    || !!slot?.slashing
    || !!slot?.attestations
    || !!slot?.sync
  let hasSuccessDuties = false
  let hasFailedDuties = false
  let maxCount = 0
  let hasScheduledDuty = false

  if (hasDuties) {
    if (slot.proposal) {
      const dutyText = $t(`slot_viz.tooltip.proposal.${slot.status}.main`)
      const dutySubText = $t(`slot_viz.tooltip.proposal.${slot.status}.sub`)
      let className = 'scheduled'
      switch (slot.status) {
        case 'proposed':
          className = 'success'
          hasSuccessDuties = true
          break
        case 'missed':
        case 'orphaned':
          className = 'failed'
          hasFailedDuties = true
          break
        case 'scheduled':
          hasScheduledDuty = true
          break
      }
      rows.push([ {
        class: className,
        count: 1,
        duties: [ {
          ...slot.proposal,
          dutySubLink:
                slot.status === 'proposed'
                  ? `/block/${slot.proposal.duty_object}`
                  : `/slot/${slot.proposal.duty_object}`,
          dutySubText,
        } ],
        dutyText,
        icon: 'proposal',
      } ])
    }

    if (slot.slashing?.failed) {
      const dutyText = $t('slot_viz.tooltip.slashing.failed.main')
      const dutySubText = $t('slot_viz.tooltip.slashing.failed.sub')
      rows.push([ {
        andMore: Math.max(
          0,
          slot.slashing.failed.total_count
          - slot.slashing.failed.slashings?.length,
        ),
        class: 'failed',
        count: slot.slashing.failed.total_count,
        duties: slot.slashing.failed.slashings?.map(slash => ({
          ...slash,
          dutySubLink: `/validator/${slash.duty_object}`,
          dutySubText,
        })),
        dutyText,
        icon: 'slashing',
      } ])
    }
    if (slot.slashing?.success) {
      hasSuccessDuties = true
      const dutyText = $t('slot_viz.tooltip.slashing.success.main')
      rows.push([ {
        class: 'success',
        count: slot.slashing.success.total_count,
        dutyText,
        icon: 'slashing',
      } ])
    }

    const addDuties = (
      type: SlotVizIcons,
      duty?: VDBSlotVizStatus<VDBSlotVizDuty>,
    ) => {
      if (!duty) {
        return
      }
      const subRows: Row[] = []
      rows.push(subRows)
      const dutyText = $t(`slot_viz.tooltip.${type}`)

      if (duty.scheduled) {
        hasScheduledDuty = true
        maxCount = Math.max(maxCount, duty.scheduled.total_count)
        subRows.push({
          andMore: Math.max(
            0,
            duty.scheduled.total_count - duty.scheduled.validators?.length,
          ),
          class: 'scheduled',
          count: duty.scheduled.total_count,
          dutyText,
          icon: type,
          validators: duty.scheduled.validators,
        })
      }
      if (duty.success) {
        hasSuccessDuties = true
        maxCount = Math.max(maxCount, duty.success.total_count)
        subRows.push({
          class: 'success',
          count: duty.success.total_count,
          dutyText,
          icon: type,
        })
      }
      if (duty.failed) {
        hasFailedDuties = true
        maxCount = Math.max(maxCount, duty.failed.total_count)
        subRows.push({
          andMore: Math.max(
            0,
            duty.failed.total_count - duty.failed.validators?.length,
          ),
          class: 'failed',
          count: duty.failed.total_count,
          dutyText,
          icon: type,
          validators: duty.failed.validators,
        })
      }
    }
    addDuties('attestation', slot.attestations)
    addDuties('sync', slot.sync)
  }

  const isScheduled
    = slot.status === 'scheduled'
    || (slot.status === 'proposed' && hasScheduledDuty)
  let stateLabel = ''
  if (isScheduled) {
    stateLabel = formatMultiPartSpan(
      $t,
      `slot_viz.tooltip.status.scheduled.${
        hasDuties ? 'has_duties' : 'no_duties'
      }`,
      [
        undefined,
        'scheduled',
        undefined,
      ],
    )
  }
  else if (hasFailedDuties && hasSuccessDuties) {
    stateLabel = formatMultiPartSpan($t, 'slot_viz.tooltip.status.duties_some', [
      undefined,
      'some',
      undefined,
    ])
  }
  else if (hasFailedDuties) {
    stateLabel = formatMultiPartSpan(
      $t,
      'slot_viz.tooltip.status.duties_failed',
      [
        undefined,
        'failed',
        undefined,
      ],
    )
  }
  else if (hasSuccessDuties) {
    stateLabel = formatMultiPartSpan(
      $t,
      'slot_viz.tooltip.status.duties_success',
      [
        undefined,
        'success',
        undefined,
      ],
    )
  }
  else {
    stateLabel = formatMultiPartSpan($t, 'slot_viz.tooltip.status.no_duties', [
      undefined,
      'scheduled',
      undefined,
    ])
  }

  return {
    hasDuties,
    minWidth: 1 + `${maxCount}`.length * 11 + 'px',
    networkLabelPath,
    rows,
    stateLabel,
  }
})
</script>

<template>
  <BcTooltip
    :target="props.id"
    layout="special"
    scroll-container="#slot-viz"
    :hover-delay="350"
  >
    <slot />
    <template #tooltip>
      <div class="with-duties">
        <div class="rows">
          <div class="row network">
            <i18n-t
              :keypath="data.networkLabelPath"
              tag="span"
            >
              <template #slot>
                <BcLink
                  :to="`/slot/${props.data.slot}`"
                  target="_blank"
                  class="link"
                >
                  <BcFormatNumber :value="props.data.slot" />
                </BcLink>
              </template>
            </i18n-t>
          </div>
          <!-- eslint-disable vue/no-v-html -->
          <div
            class="row state"
            v-html="data.stateLabel"
          />
          <!-- eslint-enable vue/no-v-html -->
        </div>
        <div
          v-for="(rows, index) in data.rows"
          :key="index"
          class="rows"
        >
          <div
            v-for="row in rows"
            :key="row.class"
            class="row"
          >
            <div
              class="count-icon"
              :class="row.class"
            >
              <span :style="{ minWidth: data.minWidth }">{{ row.count }}x</span>
              <SlotVizIcon
                :icon="row.icon"
                class="icon"
              />
            </div>
            <div class="value-col">
              <span :class="row.class">{{ row.dutyText }}</span>
              <div
                v-if="row.validators?.length"
                class="validators"
              >
                <span
                  v-for="(validator, vIndex) in row.validators"
                  :key="validator"
                >
                  <BcLink
                    :to="`/validator/${validator}`"
                    target="_blank"
                    class="link"
                  >
                    {{ validator }}
                  </BcLink>
                  <span v-if="vIndex < row.validators.length - 1 || row.andMore">,
                  </span>
                </span>
                <span v-if="row.andMore">
                  ...{{ $t("common.and_more", { count: row.andMore }) }}
                </span>
              </div>
              <div
                v-if="row.duties"
                class="duties"
              >
                <div
                  v-for="(duty, d_index) in row.duties"
                  :key="d_index"
                >
                  <BcLink
                    :to="`/validator/${duty.validator}`"
                    target="_blank"
                    class="link"
                  >
                    {{ duty.validator }}
                  </BcLink>
                  <span class="sub-text"> {{ duty.dutySubText }} </span>
                  <BcLink
                    v-if="duty.dutySubLink"
                    :to="duty.dutySubLink"
                    target="_blank"
                    class="link"
                  >
                    {{ duty.duty_object }}
                  </BcLink>
                </div>
                <div v-if="row.andMore">
                  ...{{ $t("common.and_more", { count: row.andMore }) }}
                </div>
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

    &:first-child {
      margin-left: calc(var(--padding) * -1);
      margin-right: calc(var(--padding) * -1);
      padding-left: var(--padding);
      padding-right: var(--padding);

      &:has(+ .rows) {
        border-bottom: 3px solid var(--container-border-color);
      }
    }

    &:not(:first-child):not(:nth-child(2)) {
      border-top: 1px solid var(--container-border-color);
    }

    &:nth-child(2) {
      border-width: 3px;
    }

    .row {
      display: flex;
      align-items: center;

      &.state {
        text-align: left;
      }

      &.network {
        text-wrap: nowrap;
        white-space: nowrap;
      }

      &:not(:first-child) {
        padding-top: var(--padding);
      }

      :deep(.some) {
        color: var(--yellow);
      }

      :deep(.scheduled) {
        color: var(--grey);
      }

      :deep(.success),
      &.success {
        color: var(--positive-color);
      }

      :deep(.failed),
      &.failed {
        color: var(--negative-color);
      }

      .duties > div,
      .validators {
        margin-top: var(--padding);
      }

      .count-icon {
        display: inline-flex;
        justify-content: flex-end;
        align-items: center;
        text-align: right;
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
        .validators {
          display: grid;
          grid-template-columns: 1fr 1fr 1fr;
          gap: var(--padding-tiny);
          // otherwise the text: `...and XX more` will mess up the first column
          > span:not(:has(> a)) {
            grid-column: span 2;
            text-align: left;
          }
        }
      }
    }
  }
}
</style>
