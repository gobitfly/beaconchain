<script setup lang="ts">
import { SummaryDetails, type SummaryDetail, type SummaryDetailsEfficiencyCombinedProp, type SummaryRow, type VDBSummaryTableRow } from '~/types/dashboard/summary'

interface Props {
  dashboardId: number
  row: VDBSummaryTableRow
}
const props = defineProps<Props>()

const { t: $t } = useI18n()
const { width } = useWindowSize()
const store = useValidatorDashboardSummaryDetailsStore()
const { getKey, getDetails } = store
const { detailsMap } = storeToRefs(store)

const key = computed(() => getKey(props.dashboardId, props.row.group_id))

watch(key, () => {
  getDetails(props.dashboardId, props.row.group_id)
}, { immediate: true })

const isWideEnough = computed(() => width.value >= 1400)
const summary = computed(() => detailsMap.value[key.value])

const data = computed<SummaryRow[][]>(() => {
  const tableCount = isWideEnough.value ? 1 : 4
  const list:SummaryRow[][] = [...Array.from({ length: tableCount }).map(() => [])]
  const addToList = (detail: SummaryDetail, tableIndex:number, prop: SummaryDetailsEfficiencyCombinedProp) => {
    let row:SummaryRow | undefined
    if (tableIndex && isWideEnough.value) {
      row = list[0].find(row => row.prop === prop)
    }
    if (!row) {
      let title = $t(`dashboard.validator.summary.row.${prop}`)
      if (prop === 'efficiency_total') {
        title = `${title} ${$t(`statistics.${detail.split('_')[1]}`)}`
      }
      row = { title, prop, details: [], className: '' }
      list[tableIndex].push(row)
    }
    row?.details.push(detail)
  }

  const props: SummaryDetailsEfficiencyCombinedProp[] = ['efficiency_total', 'attestation_total', 'attestation_head', 'attestation_source', 'attestation_target', 'attestation_efficiency', 'attestation_avg_incl_dist', 'sync', 'validators_sync', 'proposals', 'validators_proposal', 'slashings', 'validators_slashings', 'apr', 'luck']
  SummaryDetails.forEach((detail, index) => {
    props.forEach((prop, propIndex) => {
      if (!isWideEnough.value || propIndex) {
        addToList(detail, index, prop)
      }
    })
  })

  return list
})

</script>
<template>
  <div v-if="summary">
    <DataTable v-for="(table, index) in data" :key="index" :value="table">
      <Column field="title" />
      <Column field="col_1">
        <template #body="slotProps">
          <DashboardTableSummaryValue :class="slotProps.data.class" :data="summary" :detail="slotProps.data.details[0]" :property="slotProps.data.prop" :row="props.row" />
        </template>
      </Column>
      <Column v-if="isWideEnough" field="col_2">
        <template #body="slotProps">
          <DashboardTableSummaryValue :class="slotProps.data.class" :data="summary" :detail="slotProps.data.details[1]" :property="slotProps.data.prop" :row="props.row" />
        </template>
      </Column>
      <Column v-if="isWideEnough" field="col_3">
        <template #body="slotProps">
          <DashboardTableSummaryValue :class="slotProps.data.class" :data="summary" :detail="slotProps.data.details[2]" :property="slotProps.data.prop" :row="props.row" />
        </template>
      </Column>
    </DataTable>
  </div>
  <div v-else>
    ... todo loading ...
  </div>
</template>

<style lang="scss" scoped>
</style>
