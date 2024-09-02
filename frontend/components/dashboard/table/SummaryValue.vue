<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'
import {
  faArrowUpRightFromSquare,
  faCube,
  faSync,
} from '@fortawesome/pro-solid-svg-icons'
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

const {
  t: $t, tm: $tm,
} = useTranslation()
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
    const tooltip: { text: string,
      title: string, } | undefined = $tm(
      `dashboard.validator.tooltip.${props.property}`,
      )
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
    let context: DashboardValidatorContext = 'attestation'
    if (props.property === 'validators_proposal') {
      validators = col.proposal_validators
      context = 'proposal'
    }
    else if (props.property === 'validators_sync') {
      validators = col.sync.validators ?? []
      context = 'sync'
    }
    else if (props.property === 'validators_slashings') {
      validators = col.slashings?.validators ?? []
      context = 'slashings'
    }
    return {
      context,
      validators,
    }
  }
  else if (col && props.property === 'attestation_efficiency') {
    const tooltip: { text: string,
      title: string, } | undefined = $tm(
        'dashboard.validator.tooltip.attestation_efficiency',
      )
    return {
      attestationEfficiency: col.attestation_efficiency,
      tooltip,
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
    return { simple: { value: trim(col.attestation_avg_incl_dist, 2, 2) } }
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
      :text="data.tooltip?.text"
      :title="data.tooltip?.title"
    >
      <FontAwesomeIcon
        v-if="data.tooltip?.title"
        :icon="faInfoCircle"
      />
    </BcTooltip>
    <FontAwesomeIcon
      v-if="data?.context"
      class="link popout"
      :icon="faArrowUpRightFromSquare"
      @click="openValidatorModal"
    />
  </div>
  <DashboardTableValidators
    v-else-if="data?.validators"
    :validators="data.validators"
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
    <BcFormatPercent
      :percent="data?.attestationEfficiency"
      :color-break-point="80"
    />
    <BcTooltip
      position="top"
      :text="data.tooltip?.text"
      :title="data.tooltip?.title"
    >
      <FontAwesomeIcon :icon="faInfoCircle" />
    </BcTooltip>
  </div>
  <div
    v-else-if="data?.apr"
    class="info_row"
  >
    <BcFormatPercent :percent="data.apr.total" />
    <BcTooltip position="top">
      <FontAwesomeIcon :icon="faInfoCircle" />
      <template #tooltip>
        <div class="row">
          <b>{{ $t("common.execution_layer") }}:</b>
          <BcFormatValue
            class="space_before"
            :value="data.apr.income.el"
          /> (
          <BcFormatPercent :percent="data.apr.apr.el" />)
        </div>
        <div class="row">
          <b>{{ $t("common.consensus_layer") }}:</b>
          <BcFormatValue
            class="space_before"
            :value="data.apr.income.cl"
          /> (
          <BcFormatPercent :percent="data.apr.apr.cl" />)
        </div>
      </template>
    </BcTooltip>
  </div>
  <div
    v-else-if="data?.luck"
    class="info_row"
  >
    <span>
      <span class="no-wrap">
        <FontAwesomeIcon :icon="faCube" />
        <BcFormatPercent
          class="space_before"
          :percent="data.luck.proposal.percent"
          :precision="0"
        />
      </span>
      <span> | </span>
      <span class="no-wrap">
        <FontAwesomeIcon :icon="faSync" />
        <BcFormatPercent
          class="space_before"
          :percent="data.luck.sync.percent"
          :precision="0"
        />
      </span>
    </span>
    <BcTooltip position="top">
      <FontAwesomeIcon :icon="faInfoCircle" />
      <template #tooltip>
        <div class="row">
          <b>
            {{ $t("dashboard.validator.tooltip.block_proposal") }}
          </b>
        </div>
        <div class="row">
          <b> {{ $t("common.luck") }}: </b>
          <BcFormatPercent :percent="data.luck.proposal.percent" />
        </div>
        <div class="row">
          <b> {{ $t("common.average") }}: </b>
          {{
            $t("common.every_x", {
              duration: formatNanoSecondDuration(
                data.luck.proposal.average,
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
          <BcFormatPercent :percent="data.luck.sync.percent" />
        </div>
        <div class="row">
          <b> {{ $t("common.average") }}: </b>
          {{
            $t("common.every_x", {
              duration: formatNanoSecondDuration(data.luck.sync.average, $t),
            })
          }}
        </div>
      </template>
    </BcTooltip>
  </div>

  <BcFormatPercent
    v-else-if="data?.efficiencyTotal"
    :percent="data.efficiencyTotal.value"
    :compare-percent="data.efficiencyTotal.compare"
    :color-break-point="80"
  >
    <template #leading-tooltip="{ compare }">
      <span class="efficiency-total-tooltip">
        {{
          $t(`dashboard.validator.summary.tooltip.${compare}`, {
            name: groupName,
            average: formatPercent(row.average_network_efficiency),
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
}

.efficiency-total-tooltip {
  width: 155px;
}
</style>
