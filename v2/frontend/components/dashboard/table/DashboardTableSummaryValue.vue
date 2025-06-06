<script setup lang="ts">
import {
  type DashboardValidatorContext,
  type SummaryDetailsEfficiencyCombinedProp,
  type SummaryDetailsEfficiencyProp,
  SummaryDetailsEfficiencyProps,
  type SummaryDetailsEfficiencyValidatorProp,
  SummaryDetailsEfficiencyValidatorProps,
  type SummaryTimeFrame,
} from '~/types/dashboard/summary'
import { getGroupLabel } from '~/utils/dashboard/group'
import type {
  VDBGroupSummaryColumnItem,
  VDBGroupSummaryData,
  VDBSummaryTableRow,
} from '~/types/api/validator_dashboard'
import type { StatusCount } from '~/types/api/common'
import { DashboardValidatorSubsetModal } from '#components'

interface Props {
  absolute?: boolean,
  data?: VDBGroupSummaryData,
  inDetailView?: boolean,
  property: SummaryDetailsEfficiencyCombinedProp,
  row: VDBSummaryTableRow,
  timeFrame: SummaryTimeFrame,
}
const props = defineProps<Props>()

const { t: $t } = useTranslation()
const { dashboardKey } = useDashboardKey()
const dialog = useDialog()
const { groups } = useValidatorDashboardGroups()

const data = computed(() => {
  const col = props.data
  const row = props.row
  if (row && props.property === 'attestations') {
    return { efficiency: { status_count: row.attestations } }
  }
  else if (row && props.property === 'proposals') {
    return {
      context: !props.inDetailView ? 'proposal' : undefined,
      efficiency: { status_count: row.proposals },
    }
  }
  else if (
    col
    && SummaryDetailsEfficiencyProps.includes(
      props.property as SummaryDetailsEfficiencyProp,
    )
  ) {
    let tooltip: undefined | { text: string, title: string }
    if (props.property === 'sync') {
      tooltip = {
        text: $t('dashboard.validator.tooltip.sync_participation.text'),
        title: $t('dashboard.validator.tooltip.sync_participation.title'),
      }
    }

    const prop = col[props.property as SummaryDetailsEfficiencyProp]

    return {
      efficiency: {
        status_count:
          (prop as VDBGroupSummaryColumnItem).status_count
          || (prop as StatusCount),
        sync_count: props.property === 'sync' ? col.sync_count : undefined,
      },
      tooltip,
    }
  }
  else if (
    col
    && SummaryDetailsEfficiencyValidatorProps.includes(
      props.property as SummaryDetailsEfficiencyValidatorProp,
    )
  ) {
    let validators: number[] = []
    let validatorCount: number = 0
    let context: DashboardValidatorContext = 'attestation'
    if (props.property === 'validators_proposal') {
      validators = col.proposal_validators ?? []
      validatorCount = col.proposal_validator_count ?? 0
      context = 'proposal'
    }
    else if (props.property === 'validators_sync') {
      validators = col.sync.validators ?? []
      validatorCount = col.sync.validator_count ?? 0
      context = 'sync'
    }
    else if (props.property === 'validators_slashings') {
      validators = col.slashings?.validators ?? []
      validatorCount = col.slashings.validator_count ?? 0
      context = 'slashings'
    }
    return {
      context,
      validatorCount,
      validators,
    }
  }
  else if (col && props.property === 'attestation_efficiency') {
    const tooltip: undefined | { text: string,
      title: string, }
      = {
        text: $t('dashboard.validator.tooltip.attestation_efficiency.text'),
        title: $t('dashboard.validator.tooltip.attestation_efficiency.title'),
      }
    return {
      attestationEfficiency: col.attestation_efficiency,
      tooltip,
    }
  }
  else if (col && props.property === 'sync_efficiency') {
    return {
      syncEfficiency: col.sync_efficiency,
    }
  }
  else if (col && props.property === 'proposal_efficiency') {
    return {
      proposalEfficiency: col.proposal_efficiency,
    }
  }
  else if (row && col && props.property === 'apr') {
    return {
      apr: {
        apr: col.apr,
        income: row.reward,
        total: col.apr.cl + col.apr.el,
      },
    }
  }
  else if (col && props.property === 'luck') {
    return { luck: col.luck }
  }
  else if (row && props.property === 'efficiency') {
    return {
      efficiencyTotal: {
        compare: row.average_network_efficiency,
        value: row.efficiency,
      },
    }
  }
  else if (col && props.property === 'attestation_avg_incl_dist') {
    return {
      simple: {
        value: formatNumber(col.attestation_avg_incl_dist, {
          maximumFractionDigits: 2,
          minimumFractionDigits: 2,
        }),
      },
    }
  }
  else if (row && props.property === 'reward') {
    return { reward: row.reward }
  }
  else if (col && props.property === 'missed_rewards') {
    return { missedRewards: col.missed_rewards }
  }
  return undefined
})

const groupName = computed(() => {
  return getGroupLabel(
    $t,
    props.row.group_id,
    groups.value,
    $t('common.total'),
  )
})

const openValidatorModal = () => {
  dialog.open(DashboardValidatorSubsetModal, {
    data: {
      context: data.value?.context,
      dashboardKey: dashboardKey.value,
      groupId: props.row.group_id,
      groupName: groupName.value,
      summary: {
        data: props.data,
        row: props.row,
      },
      timeFrame: props.timeFrame,
    },
  })
}
</script>

<template>
  <DashboardTableSummaryMissedRewards
    v-if="data?.missedRewards"
    :missed-rewards="data.missedRewards"
  />
  <div
    v-else-if="data?.syncEfficiency !== undefined"
    class="info_row"
  >
    <BaseFormatPercent
      :value="data?.syncEfficiency"
      :color="data?.syncEfficiency >= 0.8 ? 'green' : 'red'"
    />
  </div>
  <div
    v-else-if="data?.proposalEfficiency !== undefined"
    class="info_row"
  >
    <BaseFormatPercent
      :value="data?.proposalEfficiency"
      :color="data?.proposalEfficiency >= 0.8 ? 'green' : 'red'"
    />
  </div>
  <DashboardTableSummaryReward
    v-else-if="data?.reward"
    :reward="data.reward"
  />
  <div
    v-else-if="data?.efficiency"
    class="info_row"
  >
    <DashboardTableEfficiency
      :absolute
      :success="data.efficiency.status_count.success"
      :failed="data.efficiency.status_count.failed"
    >
      <template
        v-if="data.efficiency.sync_count"
        #tooltip
      >
        <div>
          <div class="row">
            <b>{{ $t("dashboard.validator.summary.row.sync_committee") }}: </b>
            <DashboardTableEfficiency
              :absolute="true"
              :is-tooltip="true"
              :success="data.efficiency.status_count.success"
              :failed="data.efficiency.status_count.failed"
            />
            (
            <DashboardTableEfficiency
              :absolute="false"
              :is-tooltip="true"
              :success="data.efficiency.status_count.success"
              :failed="data.efficiency.status_count.failed"
            />
            )
          </div>
          <div class="row next_chapter">
            <b>{{ $t("common.current") }}: </b>
            <span>{{ data.efficiency.sync_count.current_validators }}
              {{
                $t("dashboard.validator.summary.tooltip.amount_of_validators")
              }}</span>
          </div>
          <div class="row">
            <b>{{ $t("common.upcoming") }}: </b>
            <span>{{ data.efficiency.sync_count.upcoming_validators }}
              {{
                $t("dashboard.validator.summary.tooltip.amount_of_validators")
              }}</span>
          </div>
          <div class="row">
            <b>{{ $t("common.past") }}: </b>
            <span>{{ data.efficiency.sync_count.past_periods }}
              {{
                $t("dashboard.validator.summary.tooltip.amount_of_rounds")
              }}</span>
          </div>
        </div>
      </template>
    </DashboardTableEfficiency>
    <BcTooltip
      position="top"
      tooltip-class="dashboard-table-summary-value__tooltip"
      :text="data.tooltip?.text"
      :title="data.tooltip?.title"
    >
      <BcIcon
        v-if="data.tooltip?.title"
        name="circle-info"
      />
    </BcTooltip>
    <BcButtonIcon
      v-if="
        data?.context
          && (data.efficiency.status_count.success || data.efficiency.status_count.failed)
      "
      screenreader-text="dashboard.validator.rewards.open_validator_details"
      class="link popout"
      name="arrow-upright-from-square"
      @click="openValidatorModal"
    />
  </div>
  <DashboardTableValidators
    v-else-if="data?.validators"
    :validators="data.validators"
    :validator-count="data?.validatorCount"
    :time-frame="props.timeFrame"
    :context="data.context"
    :dashboard-key
    :group-id="props.row.group_id"
    :data="props.data"
    :row="props.row"
  />
  <div
    v-else-if="data?.attestationEfficiency !== undefined"
    class="info_row"
  >
    <BaseFormatPercent
      :value="data?.attestationEfficiency"
      :color="data?.attestationEfficiency >= 0.8 ? 'green' : 'red'"
    />
    <BcTooltip
      position="top"
      :text="data.tooltip?.text"
      :title="data.tooltip?.title"
    >
      <BcIcon name="circle-info" />
    </BcTooltip>
  </div>
  <div
    v-else-if="data?.apr"
    class="info_row"
  >
    <BaseFormatPercent
      :value="data.apr.total"
    />
    <BcTooltip position="top">
      <BcIcon name="circle-info" />
      <template #tooltip>
        <div class="row">
          <b>{{ $t("common.execution_layer") }}:</b>
          <BcFormatAmount
            class="space_before"
            :value="data.apr.income.el"
            source-currency="elCurrency"
          /> (
          <BaseFormatPercent
            :value="data.apr.apr.el"
          />
          )
        </div>
        <div class="row">
          <b>{{ $t("common.consensus_layer") }}:</b>
          <BcFormatAmount
            class="space_before"
            :value="data.apr.income.cl"
          /> (<BaseFormatPercent
            :value="data.apr.apr.cl"
          />)
        </div>
      </template>
    </BcTooltip>
  </div>
  <div
    v-else-if="data?.luck"
    class="info_row"
  >
    <span class="info_row">
      <span class="no-wrap info_row-group">
        <BcIcon name="cube" />
        <BaseFormatPercent
          class="space_before"
          :value="data.luck.proposal.percent"
          :maximum-fraction-digits="0"
        />
      </span>
      <span class="no-wrap info_row-group">
        <BcIcon name="sync" />
        <BaseFormatPercent
          class="space_before"
          :value="data.luck.sync.percent"
          :maximum-fraction-digits="0"
        />
      </span>
    </span>
    <BcTooltip position="top">
      <BcIcon name="circle-info" />
      <template #tooltip>
        <div class="row">
          <b>
            {{ $t("dashboard.validator.tooltip.block_proposal") }}
          </b>
        </div>
        <div class="row">
          <b> {{ $t("common.luck") }}: </b>
          <BaseFormatPercent :value="data.luck.proposal.percent" />
        </div>
        <div class="row">
          <b> {{ $t("common.average") }}: </b>
          {{
            $t("common.every_x", {
              duration: formatTimeDuration(
                data.luck.proposal.average_interval_seconds,
                $t,
              ),
            })
          }}
        </div>
        <br>
        <div class="row next_chapter">
          <b class="part">
            {{ $t("dashboard.validator.tooltip.sync_committee") }}
          </b>
        </div>
        <div class="row">
          <b> {{ $t("common.luck") }}: </b>
          <BaseFormatPercent :value="data.luck.sync.percent" />
        </div>
        <div class="row">
          <b> {{ $t("common.average") }}: </b>
          {{
            $t("common.every_x", {
              duration: formatTimeDuration(data.luck.sync.average_interval_seconds, $t),
            })
          }}
        </div>
      </template>
    </BcTooltip>
  </div>

  <BcFormatPercent
    v-else-if="data?.efficiencyTotal"
    :percent="data.efficiencyTotal.value * 100"
    :compare-percent="data.efficiencyTotal.compare * 100"
    :color-break-point="80"
  >
    <template #leading-tooltip="{ compare }">
      <span class="efficiency-total-tooltip">
        {{
          $t(`dashboard.validator.summary.tooltip.${compare}`, {
            name: groupName,
            average: formatPercent(row.average_network_efficiency, {
              maximumFractionDigits: 2,
            }),
          })
        }}
      </span>
    </template>
  </BcFormatPercent>
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
  gap: var(--padding-small);
}
.info_row-group {
  display: flex;
  width: fit-content;
  gap: var(--padding-small);
  justify-content: space-between;
  align-items: center;
}

.efficiency-total-tooltip {
  width: 155px;
}

:global(.dashboard-table-summary-value__tooltip) {
  white-space: pre-line;
}
</style>
