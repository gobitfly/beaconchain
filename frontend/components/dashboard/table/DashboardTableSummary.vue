<script setup lang="ts">
import type { DataTableSortEvent } from 'primevue/datatable'
import type { InternalGetValidatorDashboardSummaryResponse, VDBSummaryTableRow } from '~/types/api/validator_dashboard'
import type { Cursor, TableQueryParams } from '~/types/datatable'

interface Props {
  dashboardId: number
}
const props = defineProps<Props>()

const cursor = ref<Cursor>()
const pageSize = ref<number>(5)
const { t: $t } = useI18n()

const store = useValidatorDashboardSummaryStore()
const { getSummary } = store
const { summaryMap, queryMap } = storeToRefs(store)

const { overview } = storeToRefs(useValidatorDashboardOverview())

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
  getSummary(props.dashboardId, query)
}

watch(() => props.dashboardId, () => {
  loadData()
}, { immediate: true })

const summary = computed<InternalGetValidatorDashboardSummaryResponse | undefined>(() => {
  return summaryMap.value?.[props.dashboardId]
})

const groupNameLabel = (groupId?: number) => {
  if (groupId === undefined || groupId < 0) {
    return `${$t('dashboard.validator.summary.total_group_name')}`
  }
  const group = overview.value?.groups?.find(g => g.id === groupId)
  if (!group) {
    return `${groupId}` // fallback if we could not match the group name
  }
  return `${group.name}`
}

const groupIdLabel = (groupId?: number) => {
  if (groupId === undefined || groupId < 0) {
    return
  }
  const group = overview.value?.groups?.find(g => g.id === groupId)
  if (group && isMobile.value) {
    return
  }
  return ` (ID: ${groupId})`
}

const onSort = (sort: DataTableSortEvent) => {
  loadData(setQuerySort(sort, queryMap.value[props.dashboardId]))
}

const setCursor = (value: Cursor) => {
  cursor.value = value
  loadData(setQueryCursor(value, queryMap.value[props.dashboardId]))
}

const setPageSize = (value: number) => {
  pageSize.value = value
  loadData(setQueryPageSize(value, queryMap.value[props.dashboardId]))
}

const setSearch = (value?: string) => {
  loadData(setQuerySearch(value, queryMap.value[props.dashboardId]))
}

const getRowClass = (row: VDBSummaryTableRow) => {
  if (row.group_id === -1) {
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
            @set-cursor="setCursor"
            @sort="onSort"
            @set-page-size="setPageSize"
          >
            <Column
              field="group_id"
              :sortable="true"
              body-class="bold"
              :header="$t('dashboard.validator.summary.col.group')"
            >
              <template #body="slotProps">
                {{ groupNameLabel(slotProps.data.group_id) }}<span class="discreet">{{ groupIdLabel(slotProps.data.group_id) }}</span>
              </template>
            </Column>
            <Column
              field="efficiency_day"
              :sortable="true"
              :header="$t('dashboard.validator.summary.col.efficiency_day')"
            >
              <template #body="slotProps">
                <BcFormatPercent :percent="slotProps.data.efficiency_day" :color-break-point="80" />
              </template>
            </Column>
            <Column
              v-if="colsVisible.efficiency_plus"
              field="efficiency_week"
              :sortable="true"
              :header="$t('dashboard.validator.summary.col.efficiency_week')"
            >
              <template #body="slotProps">
                <BcFormatPercent :percent="slotProps.data.efficiency_week" :color-break-point="80" />
              </template>
            </Column>
            <Column
              v-if="colsVisible.efficiency_plus"
              field="efficiency_month"
              :sortable="true"
              :header="$t('dashboard.validator.summary.col.efficiency_month')"
            >
              <template #body="slotProps">
                <BcFormatPercent :percent="slotProps.data.efficiency_month" :color-break-point="80" />
              </template>
            </Column>
            <Column
              v-if="colsVisible.efficiency_plus"
              field="efficiency_total"
              :sortable="true"
              :header="$t('dashboard.validator.summary.col.efficiency_total')"
            >
              <template #body="slotProps">
                <BcFormatPercent :percent="slotProps.data.efficiency_total" :color-break-point="80" />
              </template>
            </Column>
            <Column
              v-if="colsVisible.validator"
              class="validator_column"
              :sortable="true"
              :header="$t('dashboard.validator.summary.col.validators')"
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
              <DashboardTableSummaryDetails :row="slotProps.data" :dashboard-id="props.dashboardId" />
            </template>
          </BcTable>
        </ClientOnly>
      </template>
      <template #chart>
        TODO: Chart
      </template>
    </BcTableControl>
  </div>
</template>

<style lang="scss" scoped>
:deep(.summary_table) {
  --col-width: 220px;

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
    td {
      border-bottom-color: var(--primary-color);
    }
  }
}
</style>
