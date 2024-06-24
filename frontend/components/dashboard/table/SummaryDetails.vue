<script setup lang="ts">
import type { VDBSummaryTableRow } from '~/types/api/validator_dashboard'
import { type SummaryDetailsEfficiencyCombinedProp, type SummaryRow, type SummaryTableVisibility, type SummaryTimeFrame } from '~/types/dashboard/summary'

interface Props {
  row: VDBSummaryTableRow
  timeFrame: SummaryTimeFrame
  tableVisibility: SummaryTableVisibility
}
const props = defineProps<Props>()

const { dashboardKey } = useDashboardKey()

const { t: $t } = useI18n()
const { details: summary, getDetails } = useValidatorDashboardSummaryDetailsStore(dashboardKey.value, props.row.group_id)

watch(() => props.timeFrame, () => {
  getDetails(props.timeFrame)
}, { deep: true, immediate: true })

const data = computed<SummaryRow[][]>(() => {
  const list: SummaryRow[][] = [[], [], []]

  const addToList = (index: number, prop?: SummaryDetailsEfficiencyCombinedProp, titleKey?: string) => {
    const title = $t(`dashboard.validator.summary.row.${prop || titleKey}`)
    const row = { title, prop }
    list[index].push(row)
  }

  const addPropsTolist = (index: number, props: SummaryDetailsEfficiencyCombinedProp[]) => {
    props.forEach(p => addToList(index, p))
  }

  if (!props.tableVisibility.efficiency) {
    addToList(0, 'efficiency')
  }
  if (!props.tableVisibility.proposals) {
    addToList(0, 'proposals')
  }
  if (!props.tableVisibility.attestations) {
    addToList(0, 'attestations')
  } else {
    addToList(0, undefined, 'attestations')
  }
  addPropsTolist(0, ['attestations_head', 'attestations_source', 'attestations_target', 'attestation_efficiency', 'attestation_avg_incl_dist'])

  addPropsTolist(1, ['sync', 'validators_sync'])

  if (!props.tableVisibility.proposals) {
    addToList(1, 'proposals')
  } else {
    addToList(1, undefined, 'proposals')
  }
  addPropsTolist(1, ['validators_proposal', 'slashings', 'validators_slashings'])

  addPropsTolist(2, ['apr', 'luck', 'missed_rewards'])
  if (!props.tableVisibility.reward) {
    addToList(2, 'reward')
  }

  return list
})

const rowClass = (data: SummaryRow) => {
  if (!data.prop) {
    return 'bold' // headline without prop
  }
  const classNames: Partial<Record<SummaryDetailsEfficiencyCombinedProp, string>> = {
    efficiency: 'bold',
    attestations: 'bold',
    sync: 'bold spacing-top',
    proposals: 'bold spacing-top',
    slashings: 'bold spacing-top',
    apr: 'bold',
    luck: 'bold spacing-top',
    attestations_head: 'spacing-top'
  }
  return classNames[data.prop]
}

</script>
<template>
  <div v-if="summary" class="details-container">
    <div v-for="(list, index) in data" :key="index" class="group">
      <div v-for="(prop, pIndex) in list" :key="pIndex" :class="rowClass(prop)" class="row">
        <div class="label">
          {{ prop.title }}
        </div>
        <DashboardTableSummaryValue
          v-if="prop.prop"
          class="value"
          :data="summary"
          :property="prop.prop"
          :time-frame="timeFrame"
          :row-x="props.row"
          :row="{ ...props.row, reward: { el: '1239082134123213', cl: '02394' } }"
        />
      </div>
    </div>
  </div>
  <div v-else>
    <BcLoadingSpinner class="spinner" :loading="true" alignment="center" />
  </div>
</template>

<style lang="scss" scoped>
@use '~/assets/css/utils.scss';

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

    .spacing-top{
      margin-top: var(--padding-small);
    }

    @media (max-width: 1014px) {
      width: 344px;

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
      width: 100%;

      &:not(:first-child) {
        border-left: unset;
        margin-top: 0;
      }
    }

    .row {
      display: flex;
      gap: var(--padding);

      .label {
        width: 170px;
        @include utils.truncate-text;

        @media (max-width: 729px) {
          width: 151px;
        }

      }

      .value {
        flex-grow: 1;
      }
    }
  }
}

.spinner {
  padding: var(--padding-large);
}
</style>
