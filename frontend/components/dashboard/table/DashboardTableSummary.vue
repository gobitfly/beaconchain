<script setup lang="ts">
import type { DataTableSortEvent } from 'primevue/datatable'
import { type VDBSummaryTableResponse } from '~/types/dashboard/summary'
import type { Cursor } from '~/types/datatable'

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

const loadData = () => {
  getSummary(props.dashboardId, queryMap.value[props.dashboardId])
}

watch(() => props.dashboardId, () => {
  loadData()
}, { immediate: true })

const summary = computed<VDBSummaryTableResponse | undefined>(() => {
  return summaryMap.value?.[props.dashboardId]
})

const mapGroup = (groupId?: number) => {
  if (groupId === undefined || groupId < 0) {
    return $t('dashboard.validator.summary.total_group_name')
  }
  const group = overview.value?.groups?.find(g => g.id === groupId)
  if (!group) {
    return groupId
  }
  if (isMobile.value) {
    return group.name
  }
  return `${group.name} (${$t('common.id')}: ${groupId})`
}

const onSort = (sort:DataTableSortEvent) => {
  queryMap.value[props.dashboardId] = setQuerySort(sort, queryMap.value[props.dashboardId])
  loadData()
}

const setCursor = (value: Cursor) => {
  cursor.value = value
  queryMap.value[props.dashboardId] = setQueryCursor(value, queryMap.value[props.dashboardId])
  loadData()
}

const setPageSize = (value: number) => {
  pageSize.value = value
  queryMap.value[props.dashboardId] = setQueryPageSize(value, queryMap.value[props.dashboardId])
  loadData()
}

const setSearch = (value?: string) => {
  queryMap.value[props.dashboardId] = setQuerySearch(value, queryMap.value[props.dashboardId])
  loadData()
}

</script>
<template>
  <BcTable
    :data="summary"
    data-key="group_id"
    :expandable="true"
    class="summary_table"
    :cursor="cursor"
    :page-size="pageSize"
    @set-cursor="setCursor"
    @sort="onSort"
    @set-search="setSearch"
    @set-page-size="setPageSize"
  >
    <Column field="group_id" body-class="bold" :sortable="true" :header="$t('dashboard.validator.summary.col.group')">
      <template #body="slotProps">
        {{ mapGroup(slotProps.data.group_id) }}
      </template>
    </Column>
    <Column field="efficiency_24h" :sortable="true" :header="$t('dashboard.validator.summary.col.efficiency_24h')">
      <template #body="slotProps">
        <BcFormatPercent :percent="slotProps.data.efficiency_24h" :color-break-point="80" />
      </template>
    </Column>
    <Column
      v-if="colsVisible.efficiency_plus"
      field="efficiency_7d"
      :sortable="true"
      :header="$t('dashboard.validator.summary.col.efficiency_7d')"
    >
      <template #body="slotProps">
        <BcFormatPercent :percent="slotProps.data.efficiency_7d" :color-break-point="80" />
      </template>
    </Column>
    <Column
      v-if="colsVisible.efficiency_plus"
      field="efficiency_31d"
      :sortable="true"
      :header="$t('dashboard.validator.summary.col.efficiency_31d')"
    >
      <template #body="slotProps">
        <BcFormatPercent :percent="slotProps.data.efficiency_31d" :color-break-point="80" />
      </template>
    </Column>
    <Column
      v-if="colsVisible.efficiency_plus"
      field="efficiency_all"
      :sortable="true"
      :header="$t('dashboard.validator.summary.col.efficiency_all')"
    >
      <template #body="slotProps">
        <BcFormatPercent :percent="slotProps.data.efficiency_all" :color-break-point="80" />
      </template>
    </Column>
    <Column
      v-if="colsVisible.validator"
      field="validators"
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
</template>

<style lang="scss" scoped>
.summary_table {
  :deep(td:not(.expander)):not(:last-child),
  :deep(th:not(.expander)):not(:last-child) {
    width: 220px;
    max-width: 220px;
    min-width: 220px;
  }
}
</style>
