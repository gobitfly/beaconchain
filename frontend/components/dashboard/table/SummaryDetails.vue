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
  const list: SummaryRow[][] = [...Array.from({ length: tableCount }).map(() => [])]
  const bold: Partial<Record<SummaryDetailsEfficiencyCombinedProp, boolean>> = {
    efficiency_total: true,
    attestation_total: true,
    sync: true,
    proposals: true,
    slashings: true,
    apr: true,
    luck: true
  }
  const addToList = (detail: SummaryDetail, tableIndex: number, prop: SummaryDetailsEfficiencyCombinedProp) => {
    let row: SummaryRow | undefined
    if (tableIndex && isWideEnough.value) {
      row = list[0].find(row => row.prop === prop)
    }
    if (!row) {
      let title = $t(`dashboard.validator.summary.row.${prop}`)
      if (prop === 'efficiency_total') {
        title = `${title} ${$t(`statistics.${detail.split('_')[1]}`)}`
      }
      row = { title, prop, details: [], className: bold[prop] ? 'bold' : '' }
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
  <div v-if="summary" class="table-container">
    <DataTable
      v-for="(table, index) in data"
      :key="index"
      class="no-header bc-compact-table summary-details-table"
      :class="{small:!isWideEnough}"
      :value="table"
    >
      <Column field="expansion-spacer" class="expansion-spacer">
        <template #body>
          <span />
        </template>
      </Column>
      <Column field="title">
        <template #body="slotProps">
          <span :class="slotProps.data.className">
            {{ slotProps.data.title }}
          </span>
        </template>
      </Column>
      <Column field="col_1">
        <template #body="slotProps">
          <DashboardTableSummaryValue
            :class="slotProps.data.className"
            :data="summary"
            :detail="slotProps.data.details[0]"
            :property="slotProps.data.prop"
            :row="props.row"
          />
        </template>
      </Column>
      <Column v-if="isWideEnough" field="col_2">
        <template #body="slotProps">
          <DashboardTableSummaryValue
            :class="slotProps.data.className"
            :data="summary"
            :detail="slotProps.data.details[1]"
            :property="slotProps.data.prop"
            :row="props.row"
          />
        </template>
      </Column>
      <Column v-if="isWideEnough" field="col_3">
        <template #body="slotProps">
          <DashboardTableSummaryValue
            :class="slotProps.data.className"
            :data="summary"
            :detail="slotProps.data.details[2]"
            :property="slotProps.data.prop"
            :row="props.row"
          />
        </template>
      </Column>
      <Column v-if="isWideEnough" field="col_4">
        <template #body="slotProps">
          <DashboardTableSummaryValue
            :class="slotProps.data.className"
            :data="summary"
            :detail="slotProps.data.details[3]"
            :property="slotProps.data.prop"
            :row="props.row"
          />
        </template>
      </Column>
      <Column field="space_filler">
        <template #body>
          <span /> <!--used to fill up the empty space so that the last column does not strech endlessly -->
        </template>
      </Column>
    </DataTable>
  </div>
  <div v-else>
    ... TODO: loading ...
  </div>
</template>

<style lang="scss" scoped>
.table-container {
  display: flex;
  flex-wrap: wrap;
}

:deep(.summary-details-table) {
  width: 100%;

  &.small {
    width: 50%;

    &:nth-child(even) {
      .p-datatable-tbody {
        >tr {
          >td:first-child {
            border-width: 0 0 0 1px;
          }
        }
      }
    }
  }

  @media (max-width: 1180px) {
    &:not(:first-child) {
      .p-datatable-tbody {
        >tr:first-child {
          >td {
            border-width: 1px 0 0 0;
          }
        }
      }
    }
  }
}
</style>
