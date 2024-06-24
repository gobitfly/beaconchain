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

  if (!props.tableVisibility.proposals) {
    addToList(0, 'proposals')
  }

  // const props: SummaryDetailsEfficiencyCombinedProp[] = ['efficiency', 'attestation_total', 'attestations_head', 'attestations_source', 'attestations_target', 'attestation_efficiency', 'attestation_avg_incl_dist', 'sync', 'validators_sync', 'proposals', 'validators_proposal', 'slashings', 'validators_slashings', 'apr', 'luck']
  /*
  props.forEach((prop) => {

  }) */

  return list
})

const rowClass = (data:SummaryRow) => {
  if (!data.prop) {
    return 'bold' // headline without prop
  }
  const classNames: Partial<Record<SummaryDetailsEfficiencyCombinedProp, string>> = {
    efficiency: 'bold',
    attestation_total: 'bold',
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
    <div v-for="(list, index) in data" :key="index">
      <div v-for="(prop, pIndex) in list" :key="pIndex" :class="rowClass(prop)">
        <div class="">
          {{ prop.title }}
        </div>
        <DashboardTableSummaryValue
          v-if="prop.prop"
          :data="summary"
          :detail="summary"
          :property="prop.prop"
          :time-frame="timeFrame"
          :row="props.row"
        />
      </div>
    </div>
  </div>
  <div v-else>
    <BcLoadingSpinner class="spinner" :loading="true" alignment="center" />
  </div>
</template>

<style lang="scss" scoped>
.details-container {
  display: flex;
  flex-wrap: wrap;

  @media (max-width: 1180px) {
    flex-direction: column;
  }
}

.spinner{
  padding: var(--padding-large);
}
</style>
