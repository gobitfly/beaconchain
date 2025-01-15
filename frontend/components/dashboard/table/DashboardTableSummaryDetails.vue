<script setup lang="ts">
import type { VDBSummaryTableRow } from '~/types/api/validator_dashboard'
import type {
  SummaryDetailsEfficiencyCombinedProp,
  SummaryRow,
  SummaryTableVisibility,
  SummaryTimeFrame,
} from '~/types/dashboard/summary'

interface Props {
  absolute: boolean,
  row: VDBSummaryTableRow,
  tableVisibility: SummaryTableVisibility,
  timeFrame: SummaryTimeFrame,
}
const props = defineProps<Props>()

const { dashboardKey } = useDashboardKey()

const { t: $t } = useTranslation()
const {
  details: summary, getDetails,
}
  = useValidatorDashboardSummaryDetailsStore(
    dashboardKey.value,
    props.row.group_id,
  )

watch(
  () => props.timeFrame,
  () => {
    getDetails(props.timeFrame)
  },
  {
    deep: true,
    immediate: true,
  },
)

type CombinedPropOrUndefined = SummaryDetailsEfficiencyCombinedProp | undefined

const summarySections = computed<SummaryRow[][]>(() => {
  const sections: SummaryRow[][] = [
    [],
    [],
    [],
  ]

  const addToSection = (
    index: number,
    summaryProperty?: SummaryDetailsEfficiencyCombinedProp,
  ) => {
    if (!summaryProperty) {
      return
    }
    const title = $t(`dashboard.validator.summary.row.${summaryProperty}`)
    const row = {
      summaryProperty,
      title,
    }
    sections[index].push(row)
  }

  const addSummaryPropertiesToSection = (
    sectionIndex: number,
    summaryProperties: CombinedPropOrUndefined[],
  ) => {
    summaryProperties.forEach(summaryProperty => addToSection(sectionIndex, summaryProperty))
  }

  const rewardCols: CombinedPropOrUndefined[]
  = [ (!props.tableVisibility.reward ? 'reward' : undefined) ]

  let addCols: CombinedPropOrUndefined[] = props.tableVisibility
    .attestations
    ? []
    : rewardCols
  addSummaryPropertiesToSection(0, [
    (!props.tableVisibility.efficiency ? 'efficiency' : undefined),
    ...addCols,
    'attestations',
    'attestations_source',
    'attestations_target',
    'attestations_head',
    'attestation_efficiency',
    'attestation_avg_incl_dist',
  ])

  addSummaryPropertiesToSection(1, [
    'sync',
    'validators_sync',
    'proposals',
    'validators_proposal',
    'slashings',
    'validators_slashings',
  ])

  addCols = !props.tableVisibility.attestations ? [] : rewardCols
  addSummaryPropertiesToSection(2, [
    'apr',
    'luck',
    'missed_rewards',
    ...addCols,
  ])

  return sections
})

const rowClass = (data: SummaryRow) => {
  if (!data.property) {
    return 'bold' // headline without prop
  }
  const classNames: Partial<
    Record<SummaryDetailsEfficiencyCombinedProp, string>
  > = {
    apr: props.tableVisibility.attestations ? '' : 'spacing-top',
    attestations: 'bold',
    attestations_head: 'spacing-top',
    efficiency: 'bold',
    luck: 'spacing-top',
    proposals: 'bold spacing-top',
    reward: 'bold',
    slashings: 'bold spacing-top',
    sync: props.tableVisibility.efficiency ? 'bold' : 'bold spacing-top',
  }
  return classNames[data.property]
}
</script>

<template>
  <div
    v-if="summary"
    class="details-container"
  >
    <div
      v-for="(summarySection, index) in summarySections"
      :key="index"
      class="group"
    >
      <div
        v-for="(summaryRow, rowIndex) in summarySection"
        :key="rowIndex"
        :class="rowClass(summaryRow)"
        class="row"
      >
        <div class="label">
          {{ summaryRow.title }}
        </div>
        <DashboardTableSummaryValue
          v-if="summaryRow.property"
          class="value"
          :data="summary"
          :absolute
          :property="summaryRow.property"
          :time-frame
          :row="props.row"
          :in-detail-view="true"
        />
      </div>
    </div>
  </div>
  <div v-else>
    <BcLoadingSpinner
      class="spinner"
      :loading="true"
      alignment="center"
    />
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/utils.scss";

.details-container {
  display: flex;
  flex-wrap: wrap;
  padding: 6px 0 0 var(--padding);
  color: var(--container-color);
  background-color: var(--container-background);

  font-size: var(--small_text_font_size);

  .bold {
    font-weight: var(--standard_text_bold_font_weight);
  }

  .group {
    display: flex;
    flex-direction: column;
    gap: 9px;
    padding: 6px var(--padding-large);
    margin: var(--padding) 0;
    width: 33%;

    &:not(:first-child) {
      border-left: var(--container-border);
    }

    .spacing-top {
      margin-top: var(--padding-small);
    }

    @media (max-width: 1014px) {
      width: 50%;

      &:last-child {
        border-top: var(--container-border);
        border-left: unset;
        margin-top: 0;

        @media (max-width: 729px) {
          border-top: unset;
        }
      }
    }

    @media (max-width: 729px) {
      width: 340px;

      &:not(:first-child) {
        border-left: unset;
        margin-top: 0;
      }
    }

    .row {
      display: flex;
      gap: var(--padding);

      .label {
        flex-shrink: 0;
        width: 170px;
        @include utils.truncate-text;

        @media (max-width: 729px) {
          width: 151px;
        }
      }

      .value {
        flex-grow: 1;
        overflow: hidden;
      }
    }
  }
}

.spinner {
  padding: var(--padding-large);
}
</style>
