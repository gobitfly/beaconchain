<script setup lang="ts">
import type { DataTableSortEvent } from 'primevue/datatable'
import type { VDBSummaryTableRow } from '~/types/api/validator_dashboard'
import type { Cursor, TableQueryParams } from '~/types/datatable'
import { useValidatorDashboardOverviewStore } from '~/stores/dashboard/useValidatorDashboardOverviewStore'
import { DAHSHBOARDS_ALL_GROUPS_ID } from '~/types/dashboard'
import { getGroupLabel } from '~/utils/dashboard/group'

const { dashboardKey, isPublic } = useDashboardKey()

const cursor = ref<Cursor>()
const pageSize = ref<number>(10)
const { t: $t } = useI18n()
const showInDevelopment = Boolean(useRuntimeConfig().public.showInDevelopment)

const { summary, query: lastQuery, isLoading, getSummary } = useValidatorDashboardSummaryStore()
const { value: query, temp: tempQuery, bounce: setQuery } = useDebounceValue<TableQueryParams | undefined>(undefined, 500)

const { overview, hasValidators } = useValidatorDashboardOverviewStore()
const { groups } = useValidatorDashboardGroups()

const { width } = useWindowSize()
const colsVisible = computed(() => {
  return {
    validator: width.value >= 1400,
    efficiency_plus: width.value >= 1180
  }
})

const loadData = (q?: TableQueryParams) => {
  if (!q) {
    q = query.value ? { ...query.value } : { limit: pageSize.value, sort: 'group_id:desc' }
  }
  setQuery(q, true, true)
}

watch([dashboardKey, overview], () => {
  loadData()
}, { immediate: true })

watch(query, (q) => {
  if (q) {
    getSummary(dashboardKey.value, q)
  }
}, { immediate: true })

const groupNameLabel = (groupId?: number) => {
  return getGroupLabel($t, groupId, groups.value, 'Σ')
}

const onSort = (sort: DataTableSortEvent) => {
  loadData(setQuerySort(sort, lastQuery?.value))
}

const setCursor = (value: Cursor) => {
  cursor.value = value
  loadData(setQueryCursor(value, lastQuery?.value))
}

const setPageSize = (value: number) => {
  pageSize.value = value
  loadData(setQueryPageSize(value, lastQuery?.value))
}

const setSearch = (value?: string) => {
  loadData(setQuerySearch(value, lastQuery?.value))
}

const getRowClass = (row: VDBSummaryTableRow) => {
  if (row.group_id === DAHSHBOARDS_ALL_GROUPS_ID) {
    return 'total-row'
  }
}

const searchPlaceholder = computed(() => $t(isPublic.value ? 'dashboard.validator.summary.search_placeholder_public' : 'dashboard.validator.summary.search_placeholder'))

</script>
<template>
  <div>
    <BcTableControl
      :title="$t('dashboard.validator.summary.title')"
      :search-placeholder="searchPlaceholder"
      :chart-disabled="!showInDevelopment"
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
            :selected-sort="tempQuery?.sort"
            :loading="isLoading"
            @set-cursor="setCursor"
            @sort="onSort"
            @set-page-size="setPageSize"
          >
            <Column
              field="group_id"
              :sortable="showInDevelopment"
              body-class="group-id bold"
              header-class="group-id"
              :header="$t('dashboard.validator.col.group')"
            >
              <template #body="slotProps">
                {{ groupNameLabel(slotProps.data.group_id) }}
              </template>
            </Column>
            <Column
              field="efficiency_last_24h"
              :sortable="showInDevelopment"
              :header="$t('dashboard.validator.col.efficiency_last_24h')"
            >
              <template #body="slotProps">
                <BcFormatPercent :percent="slotProps.data.efficiency.last_24h" :color-break-point="80" />
              </template>
            </Column>
            <Column
              v-if="colsVisible.efficiency_plus"
              field="efficiency_last_7d"
              :sortable="showInDevelopment"
              :header="$t('dashboard.validator.col.efficiency_last_7d')"
            >
              <template #body="slotProps">
                <BcFormatPercent :percent="slotProps.data.efficiency.last_7d" :color-break-point="80" />
              </template>
            </Column>
            <Column
              v-if="colsVisible.efficiency_plus"
              field="efficiency_last_30d"
              :sortable="showInDevelopment"
              :header="$t('dashboard.validator.col.efficiency_last_30d')"
            >
              <template #body="slotProps">
                <BcFormatPercent :percent="slotProps.data.efficiency.last_30d" :color-break-point="80" />
              </template>
            </Column>
            <Column
              v-if="colsVisible.efficiency_plus"
              field="efficiency_all_time"
              :sortable="showInDevelopment"
              :header="$t('dashboard.validator.col.efficiency_all_time')"
            >
              <template #body="slotProps">
                <BcFormatPercent :percent="slotProps.data.efficiency.all_time" :color-break-point="80" />
              </template>
            </Column>
            <Column
              v-if="colsVisible.validator"
              class="validator_column"
              :sortable="showInDevelopment"
              :header="$t('dashboard.validator.col.validators')"
            >
              <template #body="slotProps">
                <DashboardTableValidators
                  :validators="slotProps.data.validators"
                  :group-id="slotProps.data.group_id"
                  :dashboard-key="dashboardKey"
                  context="group"
                />
              </template>
            </Column>
            <template #expansion="slotProps">
              <DashboardTableSummaryDetails :row="slotProps.data" />
            </template>
            <template #empty>
              <DashboardTableAddValidator v-if="!hasValidators" />
            </template>
          </BcTable>
        </ClientOnly>
      </template>
      <template #chart>
        <div class="chart-container">
          <DashboardChartSummaryChart v-if="showInDevelopment" />
        </div>
      </template>
    </BcTableControl>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/utils.scss";

:deep(.summary_table) {
  --col-width: 216px;

  >.p-datatable-wrapper {
    min-height: 529px;
  }

  .group-id {
    @include utils.truncate-text;
  }

  td:has(.validator_column) {
    @include utils.set-all-width(var(--col-width));
  }

  td,
  th {
    &:not(.expander):not(:last-child) {
      @include utils.set-all-width(var(--col-width));
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
