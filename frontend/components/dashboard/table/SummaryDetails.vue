<script setup lang="ts">
import type { VDBSummaryTableRow } from '~/types/api/validator_dashboard'
import { type SummaryDetailsEfficiencyCombinedProp, type SummaryRow, type SummaryTableVisibility, type SummaryTimeFrame } from '~/types/dashboard/summary'

interface Props {
  row: VDBSummaryTableRow
  timeFrame: SummaryTimeFrame
  tableVisibilty: SummaryTableVisibility
}
const props = defineProps<Props>()

const { dashboardKey } = useDashboardKey()

const { t: $t } = useI18n()
const { width } = useWindowSize()
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

  if()


  const props: SummaryDetailsEfficiencyCombinedProp[] = ['efficiency', 'attestation_total', 'attestations_head', 'attestations_source', 'attestations_target', 'attestation_efficiency', 'attestation_avg_incl_dist', 'sync', 'validators_sync', 'proposals', 'validators_proposal', 'slashings', 'validators_slashings', 'apr', 'luck']

  props.forEach((prop) => {

  })

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
  <div v-if="summary" class="table-container">
    <!--BcTable
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
    </BcTable-->
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
