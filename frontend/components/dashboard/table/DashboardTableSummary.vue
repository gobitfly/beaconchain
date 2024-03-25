<script setup lang="ts">
import type { DataTableSortEvent } from 'primevue/datatable'
import SummaryChart from '../chart/SummaryChart.vue'
import type { InternalGetValidatorDashboardSummaryResponse, VDBSummaryTableRow } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import type { Cursor, TableQueryParams } from '~/types/datatable'
import { useValidatorDashboardOverviewStore } from '~/stores/dashboard/useValidatorDashboardOverviewStore'
import { DAHSHBOARDS_ALL_GROUPS_ID } from '~/types/dashboard'

interface Props {
  dashboardKey: DashboardKey
}
const props = defineProps<Props>()

const cursor = ref<Cursor>()
const pageSize = ref<number>(5)
const { t: $t } = useI18n()

const store = useValidatorDashboardSummaryStore()
const { getSummary } = store
const { summaryMap, queryMap } = storeToRefs(store)
const { value: query, bounce: setQuery } = useDebounceValue<TableQueryParams | undefined>(undefined, 500)

const { validatorDashboardOverview } = useValidatorDashboardOverviewStore()

const { width, isMobile } = useWindowSize()
const colsVisible = computed(() => {
  return {
    validator: width.value >= 1400,
    efficiency_plus: width.value >= 1180
  }
})

const loadData = (query?: TableQueryParams) => {
  if (!query) {
    query = { limit: pageSize.value }
  }
  setQuery(query, true, true)
}

watch(() => props.dashboardKey, () => {
  loadData()
}, { immediate: true })

watch(query, (q) => {
  if (q) {
    getSummary(props.dashboardKey, q)
  }
}, { immediate: true })

const summary = computed<InternalGetValidatorDashboardSummaryResponse | undefined>(() => {
  return summaryMap.value?.[props.dashboardKey]
})

const groupNameLabel = (groupId?: number) => {
  if (groupId === undefined || groupId < 0) {
    return `${$t('dashboard.validator.summary.total_group_name')}`
  }
  const group = validatorDashboardOverview.value?.groups?.find(g => g.id === groupId)
  if (!group) {
    return `${groupId}` // fallback if we could not match the group name
  }
  return `${group.name}`
}

const groupIdLabel = (groupId?: number) => {
  if (groupId === undefined || groupId < 0) {
    return
  }
  const group = validatorDashboardOverview.value?.groups?.find(g => g.id === groupId)
  if (group && isMobile.value) {
    return
  }
  return ` (ID: ${groupId})`
}

const onSort = (sort: DataTableSortEvent) => {
  loadData(setQuerySort(sort, queryMap.value[props.dashboardKey]))
}

const setCursor = (value: Cursor) => {
  cursor.value = value
  loadData(setQueryCursor(value, queryMap.value[props.dashboardKey]))
}

const setPageSize = (value: number) => {
  pageSize.value = value
  loadData(setQueryPageSize(value, queryMap.value[props.dashboardKey]))
}

const setSearch = (value?: string) => {
  loadData(setQuerySearch(value, queryMap.value[props.dashboardKey]))
}

const getRowClass = (row: VDBSummaryTableRow) => {
  if (row.group_id === DAHSHBOARDS_ALL_GROUPS_ID) {
    return 'total-row'
  }
}

</script>
<template>
  <div>
    <BcTableControl
      :title="$t('dashboard.validator.summary.title')"
      :search-placeholder="$t('dashboard.validator.summary.search_placeholder')"
      @set-search="setSearch"
    >
      <template #table>
        <ClientOnly fallback-tag="span">
          <BcTable
            :data="summary"
            data-key="group_id"
            :expandable="true"
            class="summary_table"
            :cursor="cursor"
            :page-size="pageSize"
            :row-class="getRowClass"
            :add-spacer="true"
            @set-cursor="setCursor"
            @sort="onSort"
            @set-page-size="setPageSize"
          >
            <Column
              field="group_id"
              :sortable="true"
              body-class="bold"
              :header="$t('dashboard.validator.col.group')"
            >
              <template #body="slotProps">
                {{ groupNameLabel(slotProps.data.group_id) }}<span class="discreet">{{
                  groupIdLabel(slotProps.data.group_id) }}</span>
              </template>
            </Column>
            <Column
              field="efficiency_last_24h"
              :sortable="true"
              :header="$t('dashboard.validator.col.efficiency_last_24h')"
            >
              <template #body="slotProps">
                <BcFormatPercent :percent="slotProps.data.efficiency.last_24h" :color-break-point="80" />
              </template>
            </Column>
            <Column
              v-if="colsVisible.efficiency_plus"
              field="efficiency_last_7d"
              :sortable="true"
              :header="$t('dashboard.validator.col.efficiency_last_7d')"
            >
              <template #body="slotProps">
                <BcFormatPercent :percent="slotProps.data.efficiency.last_7d" :color-break-point="80" />
              </template>
            </Column>
            <Column
              v-if="colsVisible.efficiency_plus"
              field="efficiency_last_30d"
              :sortable="true"
              :header="$t('dashboard.validator.col.efficiency_last_30d')"
            >
              <template #body="slotProps">
                <BcFormatPercent :percent="slotProps.data.efficiency.last_30d" :color-break-point="80" />
              </template>
            </Column>
            <Column
              v-if="colsVisible.efficiency_plus"
              field="efficiency_all_time"
              :sortable="true"
              :header="$t('dashboard.validator.col.efficiency_all_time')"
            >
              <template #body="slotProps">
                <BcFormatPercent :percent="slotProps.data.efficiency.all_time" :color-break-point="80" />
              </template>
            </Column>
            <Column
              v-if="colsVisible.validator"
              class="validator_column"
              :sortable="true"
              :header="$t('dashboard.validator.col.validators')"
            >
              <template #body="slotProps">
                <DashboardTableValidators
                  :validators="slotProps.data.validators"
                  :group-id="slotProps.data.group_id"
                  context="group"
                />
              </template>
            </Column>
            <template #expansion="slotProps">
              <DashboardTableSummaryDetails :row="slotProps.data" :dashboard-key="props.dashboardKey" />
            </template>
          </BcTable>
        </ClientOnly>
      </template>
      <template #chart>
        <div class="chart-container">
          <SummaryChart :dashboard-key="props.dashboardKey" />
        </div>
      </template>
    </BcTableControl>
  </div>
</template>

<style lang="scss" scoped>
:deep(.summary_table) {
  --col-width: 216px;

  td:has(.validator_column) {
    width: var(--col-width);
    max-width: var(--col-width);
    min-width: var(--col-width);
  }

  td,
  th {
    &:not(.expander):not(:last-child) {
      width: var(--col-width);
      max-width: var(--col-width);
      min-width: var(--col-width);

    }
  }

  @media (max-width: 600px) {
    --col-width: 140px;
  }

  .total-row {
    &:not(:has(.bottom)) {
      td {
        border-bottom-color: var(--primary-color);
      }
    }
  }

  .total-row+.p-datatable-row-expansion {
    td {
      border-bottom-color: var(--primary-color);
    }
  }
}

.chart-container {
  width: 100%;
  height: 625px;
}
</style>
