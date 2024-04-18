<script setup lang="ts">
import type { ValidatorHistoryDuties } from '~/types/api/common'

interface Props {
  data?: ValidatorHistoryDuties,
  compact?: boolean
}
const props = defineProps<Props>()

const { t: $t } = useI18n()
const { slotsPerEpoch } = useNetwork()

const mapped = computed(() => {
  const mapSuccess = (status?: 'success' | 'partial' | 'failed' | 'orphaned') => {
    const success = status === 'success'

    let className = ''
    switch (status) {
      case 'success':
        className = 'positive'
        break
      case 'partial':
        className = 'partial'
        break
      case 'failed':
        className = 'negative'
        break
      case 'orphaned':
        className = 'orphaned'
        break
    }
    return {
      success,
      className,
      status,
      tooltip: ''
    }
  }
  const head = mapSuccess(props?.data?.attestation_head?.status)
  const source = mapSuccess(props?.data?.attestation_source?.status)
  const target = mapSuccess(props?.data?.attestation_target?.status)
  const proposal = mapSuccess(props?.data?.proposal?.status)
  if (proposal.status) {
    proposal.tooltip = $t(`validator.duty.proposal_${proposal.status}`)
  }
  const slashing = mapSuccess(props?.data?.slashing?.status)
  if (slashing.status) {
    slashing.tooltip = $t(`validator.duty.slashing_${slashing.status}`)
  }
  const sync = mapSuccess(props?.data?.sync?.status)
  if (sync.status && props?.data?.sync_count !== undefined) {
    const success = props.data.sync_count
    const failed = slotsPerEpoch - success
    sync.tooltip = `${success} / ${failed}`
  }
  const totalClassName = !(head.status && source.status && target.status) ? '' : (head.success || source.success || target.success) ? 'positive' : 'negative'
  const totalTooltipTitle = totalClassName ? totalClassName === 'positive' ? $t('validator.duty.attestation_included') : $t('validator.duty.attestation_missed') : ''
  return {
    total: {
      className: totalClassName,
      tooltipTitle: totalTooltipTitle
    },
    head,
    source,
    target,
    proposal,
    slashing,
    sync
  }
})

</script>
<template>
  <div class="duty-status-container">
    <BcTooltip :fit-content="true">
      <template v-if="mapped.total.tooltipTitle" #tooltip>
        <div class="tooltip">
          <b>
            {{ mapped.total.tooltipTitle }}
          </b>
          <div class="head">
            <b>{{ $t('validator.duty.head') }}:</b> {{ $t(`common.${mapped.head.success}`) }}
          </div>
          <div><b>{{ $t('validator.duty.source') }}:</b> {{ $t(`common.${mapped.source.success}`) }}</div>
          <div><b>{{ $t('validator.duty.target') }}:</b> {{ $t(`common.${mapped.target.success}`) }}</div>
        </div>
      </template>
      <div class="attestations group" :class="mapped.total.className">
        <SlotVizIcon :class="mapped.head.className" icon="head_attestation" />
        <SlotVizIcon :class="mapped.source.className" icon="source_attestation" />
        <SlotVizIcon :class="mapped.target.className" icon="target_attestation" />
      </div>
    </BcTooltip>
    <div v-if="!compact" class="group">
      <BcTooltip :text="mapped.proposal.tooltip" :fit-content="true">
        <SlotVizIcon :class="mapped.proposal.className" icon="proposal" />
      </BcTooltip>
      <BcTooltip :text="mapped.slashing.tooltip" :fit-content="true">
        <SlotVizIcon :class="mapped.slashing.className" icon="slashing" />
      </BcTooltip>
      <BcTooltip :text="mapped.sync.tooltip" :fit-content="true">
        <SlotVizIcon :class="mapped.sync.className" icon="sync" />
      </BcTooltip>
    </div>
  </div>
</template>
<style lang="scss" scoped>
.tooltip{
  text-align: left;
  .head{
    margin-top: var(--padding);
  }
}
.duty-status-container {
  background-color: var(--subcontainer-background);
  border-radius: var(--border-radius);
  display: inline-flex;
  flex-wrap: nowrap;
  color: var(--text-color-disabled);
  height: 20px;

  .group {
    display: flex;
    flex-wrap: nowrap;
    border-radius: var(--border-radius);
    border: solid 1px transparent;

    svg{
      margin: 3px 4px;
      height: 12px;
      width: auto;
    }

    &.attestations {
      border-color: var(--text-color-disabled);

      &.negative {
        border-color: var(--negative-color);
      }
      &.positive{
        border-color: var(--positive-color);

      }
    }
  }
}
</style>
