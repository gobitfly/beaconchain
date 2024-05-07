<script setup lang="ts">
import type { VDBSummaryTableRow } from '~/types/api/validator_dashboard'
import { type SummaryDetailsEfficiencyCombinedProp, type SummaryRow } from '~/types/dashboard/summary'
import { TimeFrames, type TimeFrame } from '~/types/value'

interface Props {
  row: VDBSummaryTableRow
}
const props = defineProps<Props>()

const { dashboardKey } = useDashboardKey()

const { t: $t } = useI18n()
const { width } = useWindowSize()
const { details: summary } = useValidatorDashboardSummaryDetailsStore(dashboardKey.value, props.row.group_id)

const isWideEnough = computed(() => width.value >= 1400)

const data = computed<SummaryRow[][]>(() => {
  const tableCount = isWideEnough.value ? 1 : 4
  const list: SummaryRow[][] = [...Array.from({ length: tableCount }).map(() => [])]

  const addToList = (detail: TimeFrame, tableIndex: number, prop: SummaryDetailsEfficiencyCombinedProp) => {
    let row: SummaryRow | undefined
    if (tableIndex && isWideEnough.value) {
      row = list[0].find(row => row.prop === prop)
    }
    if (!row) {
      let title = $t(`dashboard.validator.summary.row.${prop}`)
      if (prop === 'efficiency_all_time') {
        title = `${title} (${$t(`statistics.${detail}`)})`
      }
      row = { title, prop, details: [] }
      list[tableIndex].push(row)
    }
    row?.details.push(detail)
  }

  const props: SummaryDetailsEfficiencyCombinedProp[] = ['efficiency_all_time', 'attestation_total', 'attestations_head', 'attestations_source', 'attestations_target', 'attestation_efficiency', 'attestation_avg_incl_dist', 'sync', 'validators_sync', 'proposals', 'validators_proposal', 'slashed', 'validators_slashings', 'apr', 'luck']
  TimeFrames.forEach((detail, index) => {
    props.forEach((prop, propIndex) => {
      if (!isWideEnough.value || propIndex) {
        addToList(detail, index, prop)
      }
    })
  })

  return list
})

const rowClass = (data:SummaryRow) => {
  const classNames: Partial<Record<SummaryDetailsEfficiencyCombinedProp, string>> = {
    efficiency_all_time: 'bold',
    attestation_total: 'bold',
    sync: 'bold spacing-top',
    proposals: 'bold spacing-top',
    slashed: 'bold spacing-top',
    apr: 'bold',
    luck: 'bold spacing-top',
    attestations_head: 'spacing-top'
  }
  return classNames[data.prop]
}

</script>
<template>
  <div v-if="summary" class="table-container">
    <BcTable
      v-for="(table, index) in data"
      :key="index"
      :row-class="rowClass"
      class="no-header bc-compact-table summary-details-table"
      :class="{ small: !isWideEnough }"
      :value="table"
      :add-spacer="true"
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
      <template v-for="(num, i) in 4" :key="i">
        <Column v-if="!i || isWideEnough" :field="`col_${num}`">
          <template #body="slotProps">
            <DashboardTableSummaryValue
              :class="slotProps.data.className"
              :data="summary"
              :detail="slotProps.data.details[i]"
              :property="slotProps.data.prop"
              :row="props.row"
            />
          </template>
        </Column>
      </template>
    </BcTable>
  </div>
  <div v-else>
    <BcLoadingSpinner class="spinner" :loading="true" alignment="center" />
  </div>
</template>

<style lang="scss" scoped>
.table-container {
  display: flex;
  flex-wrap: wrap;

  @media (max-width: 1180px) {
    flex-direction: column;
  }
}

.spinner{
  padding: var(--padding-large);
}

:deep(.summary-details-table) {
  width: 100%;

  &.small {
    width: 50%;

    @media (min-width: 1181px) {
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
      width: unset;

      .p-datatable-tbody {
        >tr {
          &:first-child {
            >td {
              border-width: 1px 0 0 0;
            }
          }
        }
      }
    }
  }

  .p-datatable-wrapper>.p-datatable-table>.p-datatable-tbody>tr>td {
    font-size: var(--small_text_font_size);
  }
}
</style>
