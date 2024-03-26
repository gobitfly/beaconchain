<script setup lang="ts">
import type { DataTableSortEvent } from 'primevue/datatable'
import type { VDBRewardsTableRow } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import type { Cursor, TableQueryParams } from '~/types/datatable'
import { useValidatorDashboardOverviewStore } from '~/stores/dashboard/useValidatorDashboardOverviewStore'
import { DAHSHBOARDS_ALL_GROUPS_ID } from '~/types/dashboard'
import type { BcFormatNumber } from '#build/components'

interface Props {
  dashboardKey: DashboardKey
}
const props = defineProps<Props>()

const cursor = ref<Cursor>()
const pageSize = ref<number>(5)
const { t: $t } = useI18n()

const { rewards, query: lastQuery, getRewards } = useValidatorDashboardRewardsStore(props.dashboardKey)
const { value: query, bounce: setQuery } = useDebounceValue<TableQueryParams | undefined>(undefined, 500)

const { overview } = storeToRefs(useValidatorDashboardOverviewStore())

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
    getRewards(q)
  }
}, { immediate: true })

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
  loadData(setQuerySort(sort, lastQuery.value))
}

const setCursor = (value: Cursor) => {
  cursor.value = value
  loadData(setQueryCursor(value, lastQuery.value))
}

const setPageSize = (value: number) => {
  pageSize.value = value
  loadData(setQueryPageSize(value, lastQuery.value))
}

const setSearch = (value?: string) => {
  loadData(setQuerySearch(value, lastQuery.value))
}

const getRowClass = (row: VDBRewardsTableRow) => {
  // TODO: get info from backend on how to identify the total group
  if (row.group_id === DAHSHBOARDS_ALL_GROUPS_ID) {
    return 'total-row'
  } else if (row.group_id === -2) {
    return 'future-row'
  }
}

const isRowExpandable = (row: VDBRewardsTableRow) => {
  // TODO: get info from backend on how to identify the future group [which is not expandable]
  return row.group_id === -2
}

</script>
<template>
  <div>
    <BcTableControl
      :title="$t('dashboard.validator.summary.title')"
      :search-placeholder="$t('dashboard.validator.summary.search_placeholder')"
      :is-row-expandable="isRowExpandable"
      @set-search="setSearch"
    >
      <template #table>
        <ClientOnly fallback-tag="span">
          <BcTable
            :data="rewards"
            data-key="epoch"
            :expandable="true"
            class="rewards_table"
            :cursor="cursor"
            :page-size="pageSize"
            :row-class="getRowClass"
            :add-spacer="true"
            @set-cursor="setCursor"
            @sort="onSort"
            @set-page-size="setPageSize"
          >
            <Column
              field="epoch"
              :sortable="true"
              body-class="bold"
              :header="$t('dashboard.validator.col.epoch')"
            >
              <template #body="slotProps">
                <BcFormatNumber :value="slotProps.data.epoch" />
              </template>
            </Column>
            <Column
              field="group_id"
              body-class="bold"
              :header="$t('dashboard.validator.col.group')"
            >
              <template #body="slotProps">
                {{ groupNameLabel(slotProps.data.group_id) }}<span class="discreet">{{
                  groupIdLabel(slotProps.data.group_id) }}</span>
              </template>
            </Column>
            <template #expansion="slotProps">
              Here can be your details {{ slotProps.data.group_id }}
            </template>
          </BcTable>
        </ClientOnly>
      </template>
      <template #chart>
        <div class="chart-container">
          <!--TODO: chart-->
        </div>
      </template>
    </BcTableControl>
  </div>
</template>

<style lang="scss" scoped>
:deep(.rewards_table) {
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
