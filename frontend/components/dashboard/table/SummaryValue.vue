<script setup lang="ts">
import { union } from 'lodash-es'
import {
  SummaryDetailsEfficiencyProps,
  SummaryDetailsEfficiencyValidatorProps,
  type SummaryDetail,
  type VDBGroupSummaryResponse,
  type SummaryDetailsEfficiencyValidatorProp,
  type SummaryDetailsEfficiencyProp,
  type SummaryDetailsEfficiencyCombinedProp,
  type DashboardValidatorContext,
  type VDBSummaryTableRow
} from '~/types/dashboard/summary'

interface Props {
  property: SummaryDetailsEfficiencyCombinedProp,
  detail: SummaryDetail,
  data: VDBGroupSummaryResponse,
  row: VDBSummaryTableRow
}
const props = defineProps<Props>()

const { tm: $tm } = useI18n()

const data = computed(() => {
  const col = props.data?.[props.detail]
  if (!col) {
    return null
  }
  if (props.property === 'attestation_total') {
    return {
      efficiency: {
        count: {
          success: col.attestation_head.count.success + col.attestation_source.count.success + col.attestation_target.count.success,
          failed: col.attestation_head.count.failed + col.attestation_source.count.failed + col.attestation_target.count.failed
        }
      }
    }
  } else if (SummaryDetailsEfficiencyProps.includes(props.property as SummaryDetailsEfficiencyProp)) {
    const tooltip: {title: string, text: string} | undefined = $tm(`dashboard.validator.tooltip.${props.property}`)
    return {
      efficiency: col[props.property as SummaryDetailsEfficiencyProp],
      tooltip
    }
  } else if (SummaryDetailsEfficiencyValidatorProps.includes(props.property as SummaryDetailsEfficiencyValidatorProp)) {
    let validators: number[] = []
    let context: DashboardValidatorContext = 'attestation'
    if (props.property === 'validators_attestation') {
      validators = union(col.attestation_head.validators, col.attestation_source.validators, col.attestation_target.validators)
    } else if (props.property === 'validators_proposal') {
      validators = col.proposals.validators ?? []
      context = 'proposal'
    } else if (props.property === 'validators_sync') {
      validators = col.sync.validators ?? []
      context = 'sync'
    } else if (props.property === 'validators_slashings') {
      validators = col.slashings.validators ?? []
      context = 'slashings'
    }
    return {
      validators,
      context
    }
  } else if (props.property === 'attestation_efficiency') {
    const tooltip: {title: string, text: string} | undefined = $tm('dashboard.validator.tooltip.attestation_efficiency')
    return {
      attestationEfficiency: col.attestation_efficiency,
      tooltip
    }
  } else if (props.property === 'apr') {
    return {
      apr: {
        apr: col.apr,
        total: col.apr.cl + col.apr.el,
        income: col.income
      }
    }
  } else if (props.property === 'luck') {
    return {
      luck: {
        sync: col.sync_luck,
        proposal: col.proposal_luck
      }
    }
  } else if (props.property === 'efficiency_total') {
    let efficiencyTotal = props.row.efficiency_24h
    switch (props.detail) {
      case 'details_31d':
        efficiencyTotal = props.row.efficiency_31d
        break
      case 'details_7d':
        efficiencyTotal = props.row.efficiency_7d
        break
      case 'details_all':
        efficiencyTotal = props.row.efficiency_all
        break
    }
    return {
      efficiencyTotal: {
        total: efficiencyTotal
      }
    }
  } else if (props.property === 'attestation_avg_incl_dist') {
    return {
      simple: {
        value: col.attestation_avg_incl_dist
      }
    }
  }
})

// TODO: find better matching icon for fa-info-circle, once new fontawesome files are merged
</script>
<template>
  <BcTooltip v-if="data?.efficiency" position="top" :text="data.tooltip?.text" :title="data.tooltip?.title">
    <div class="info_row">
      <DashboardTableEfficiency
        :success="data.efficiency.count.success"
        :failed="data.efficiency.count.failed"
      />
      <i v-if="data.tooltip?.title" class="fas fa-info-circle link" />
    </div>
  </BcTooltip>
  <DashboardTableValidators
    v-else-if="data?.validators"
    :validators="data.validators"
    :context="data.context"
    :group-id="props.row.group_id"
  />
  <BcTooltip v-else-if="data?.attestationEfficiency" position="top" :text="data.tooltip?.text" :title="data.tooltip?.title">
    <div class="info_row">
      <BcFormatPercent :percent="data?.attestationEfficiency" :color-break-point="80" />
      <i class="fas fa-info-circle link" />
    </div>
  </BcTooltip>
  <BcTooltip v-else-if="data?.apr" position="top">
    <div class="info_row">
      <BcFormatPercent :percent="data.apr.total" />
      <i class="fas fa-info-circle link" />
    </div>
    <template #tooltip>
      <div class="row">
        <b>{{ $t('common.execution_layer') }}:</b>
        <BcFormatValue class="space_before" :value="data.apr.income.el" /> (
        <BcFormatPercent :percent="data.apr.apr.el" />)
      </div>
      <div class="row">
        <b>{{ $t('common.consensus_layer') }}:</b>
        <BcFormatValue class="space_before" :value="data.apr.income.cl" /> (
        <BcFormatPercent :percent="data.apr.apr.cl" />)
      </div>
    </template>
  </BcTooltip>
  <BcTooltip v-else-if="data?.luck" position="top">
    <div class="info_row">
      <span><i class="fas fa-cube" />
        <BcFormatPercent class="space_before" :percent="data.luck.proposal.percent" /> / <i class="fas fa-sync" />
        <BcFormatPercent class="space_before" :percent="data.luck.sync.percent" />
      </span>
      <i class="fas fa-info-circle link" />
    </div>
    <template #tooltip>
      <div class="row">
        <b>
          {{ $t('dashboard.validator.tooltip.block_proposal') }}
        </b>
      </div>
      <div class="row">
        <b>
          {{ $t('common.luck') }}:
        </b>
        <BcFormatPercent :percent="data.luck.proposal.percent" />
      </div>
      <div class="row">
        <b>
          {{ $t('common.expected') }}:
        </b>
        {{ $t('common.in_day', {}, data.luck.proposal.expected) }}
      </div>
      <div class="row">
        <b>
          {{ $t('common.average') }}:
        </b>
        {{ $t('common.every_day', {}, data.luck.proposal.average) }}
      </div>
      <br>
      <div class="row next_chapter">
        <b class="part">
          {{ $t('dashboard.validator.tooltip.sync_committee') }}
        </b>
      </div>
      <div class="row">
        <b>
          {{ $t('common.luck') }}:
        </b>
        <BcFormatPercent :percent="data.luck.proposal.percent" />
      </div>
      <div class="row">
        <b>
          {{ $t('common.expected') }}:
        </b>
        {{ $t('common.in_day', {}, data.luck.proposal.expected) }}
      </div>
      <div class="row">
        <b>
          {{ $t('common.average') }}:
        </b>
        {{ $t('common.every_day', {}, data.luck.proposal.average) }}
      </div>
    </template>
  </BcTooltip>
  <BcFormatPercent v-else-if="data?.efficiencyTotal" :percent="data.efficiencyTotal.total" :color-break-point="80" />
  <span v-else-if="data?.simple">
    {{ data.simple?.value }}
  </span>
</template>

<style>
.row {
  text-wrap: nowrap;
  min-width: 100%;
  text-align: left;

  &.next_chapter {
    margin-top: var(--padding);
  }
}

.space_before {
  &::before {
    content: " ";
  }
}
.info_row{
  display: flex;
  justify-content: space-between;
  .fa-info {
  height: 14px;
  width: 14px;

}
}
</style>
