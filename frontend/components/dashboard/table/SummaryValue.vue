<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import {
  faInfoCircle
} from '@fortawesome/pro-regular-svg-icons'
import {
  faCube,
  faSync
} from '@fortawesome/pro-solid-svg-icons'
import {
  SummaryDetailsEfficiencyProps,
  SummaryDetailsEfficiencyValidatorProps,
  type SummaryDetailsEfficiencyValidatorProp,
  type SummaryDetailsEfficiencyProp,
  type SummaryDetailsEfficiencyCombinedProp,
  type DashboardValidatorContext,
  type SummaryTimeFrame
} from '~/types/dashboard/summary'
import type { VDBGroupSummaryColumnItem, VDBGroupSummaryData, VDBSummaryTableRow } from '~/types/api/validator_dashboard'
import type { StatusCount } from '~/types/api/common'

interface Props {
  property: SummaryDetailsEfficiencyCombinedProp,
  timeFrame: SummaryTimeFrame,
  data?: VDBGroupSummaryData,
  row: VDBSummaryTableRow,
  absolute?: boolean,
}
const props = defineProps<Props>()

const { tm: $tm } = useI18n()
const { dashboardKey } = useDashboardKey()

const data = computed(() => {
  const col = props.data
  const row = props.row
  if (row && props.property === 'attestation_total') {
    return {
      efficiency: {
        status_count: row.attestations
      }
    }
  } else if (row && props.property === 'proposals') {
    const tooltip: { title: string, text: string } | undefined = $tm(`dashboard.validator.tooltip.${props.property}`)
    return {
      efficiency: {
        status_count: row.proposals
      },
      tooltip
    }
  } else if (col && SummaryDetailsEfficiencyProps.includes(props.property as SummaryDetailsEfficiencyProp)) {
    const tooltip: { title: string, text: string } | undefined = $tm(`dashboard.validator.tooltip.${props.property}`)
    const prop = col[props.property as SummaryDetailsEfficiencyProp]

    return {
      efficiency: {
        status_count: (prop as VDBGroupSummaryColumnItem).status_count || prop as StatusCount
      },
      tooltip
    }
  } else if (col && SummaryDetailsEfficiencyValidatorProps.includes(props.property as SummaryDetailsEfficiencyValidatorProp)) {
    let validators: number[] = []
    let context: DashboardValidatorContext = 'attestation'
    if (props.property === 'validators_proposal') {
      validators = col.proposal_validators
      context = 'proposal'
    } else if (props.property === 'validators_sync') {
      validators = col.sync.validators ?? []
      context = 'sync'
    } else if (props.property === 'validators_slashings') {
      validators = col.slashings?.validators ?? []
      context = 'slashings'
    }
    return {
      validators,
      context
    }
  } else if (col && props.property === 'attestation_efficiency') {
    const tooltip: { title: string, text: string } | undefined = $tm('dashboard.validator.tooltip.attestation_efficiency')
    return {
      attestationEfficiency: col.attestation_efficiency,
      tooltip
    }
  } else if (col && props.property === 'apr') {
    return {
      apr: {
        apr: col.apr,
        total: col.apr.cl + col.apr.el,
        income: col.income
      }
    }
  } else if (col && props.property === 'luck') {
    return {
      luck: col.luck
    }
  } else if (row && props.property === 'efficiency') {
    return {
      efficiencyTotal: {
        value: row.efficiency,
        compare: row.average_network_efficiency
      }
    }
  } else if (col && props.property === 'attestation_avg_incl_dist') {
    return {
      simple: {
        value: trim(col.attestation_avg_incl_dist, 2, 2)
      }
    }
  } else if (row && props.property === 'reward') {
    return {
      reward: row.reward
    }
  }
})

</script>

<template>
  <DashboardTableSummaryReward v-if="data?.reward" :reward="data.reward" />
  <div v-else-if="data?.efficiency" class="info_row">
    <DashboardTableEfficiency
      :absolute="absolute"
      :success="data.efficiency.status_count.success"
      :failed="data.efficiency.status_count.failed"
    />
    <BcTooltip position="top" :text="data.tooltip?.text" :title="data.tooltip?.title">
      <FontAwesomeIcon v-if="data.tooltip?.title" class="link" :icon="faInfoCircle" />
    </BcTooltip>
  </div>
  <DashboardTableValidators
    v-else-if="data?.validators"
    :validators="data.validators"
    :time-frame="props.timeFrame"
    :context="data.context"
    :dashboard-key="dashboardKey"
    :group-id="props.row.group_id"
  />
  <div v-else-if="data?.attestationEfficiency !== undefined" class="info_row">
    <BcFormatPercent :percent="data?.attestationEfficiency" :color-break-point="80" />
    <BcTooltip position="top" :text="data.tooltip?.text" :title="data.tooltip?.title">
      <FontAwesomeIcon class="link" :icon="faInfoCircle" />
    </BcTooltip>
  </div>
  <div v-else-if="data?.apr" class="info_row">
    <BcFormatPercent :percent="data.apr.total" />
    <BcTooltip position="top">
      <FontAwesomeIcon class="link" :icon="faInfoCircle" />
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
  </div>
  <div v-else-if="data?.luck" class="info_row">
    <span>
      <span class="no-wrap">
        <FontAwesomeIcon :icon="faCube" />
        <BcFormatPercent class="space_before" :percent="data.luck.proposal.percent" :precision="0" />
      </span>
      <span> | </span>
      <span class="no-wrap">
        <FontAwesomeIcon :icon="faSync" />
        <BcFormatPercent class="space_before" :percent="data.luck.sync.percent" :precision="0" />
      </span>
    </span>
    <BcTooltip position="top">
      <FontAwesomeIcon class="link" :icon="faInfoCircle" />
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
            {{ $t('common.average') }}:
          </b>
          {{ $t('common.every_x', { duration: formatNanoSecondDuration(data.luck.proposal.average, $t)}) }}
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
          <BcFormatPercent :percent="data.luck.sync.percent" />
        </div>
        <div class="row">
          <b>
            {{ $t('common.average') }}:
          </b>
          {{ $t('common.every_x', { duration: formatNanoSecondDuration(data.luck.sync.average, $t)}) }}
        </div>
      </template>
    </BcTooltip>
  </div>

  <BcFormatPercent v-else-if="data?.efficiencyTotal" :percent="data.efficiencyTotal.value" :compare-percent="data.efficiencyTotal.compare" :color-break-point="80" />
  <span v-else-if="data?.simple">
    {{ data.simple?.value }}
  </span>
</template>

<style lang="scss" scoped>
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

.info_row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
