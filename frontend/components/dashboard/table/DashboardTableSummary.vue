<script setup lang="ts">
import { type VDBSummaryTableRow, type VDBSummaryTableResponse } from '~/types/dashboard/summary'

interface Props {
  dashboardId: number
}
const props = defineProps<Props>()

const { t: $t } = useI18n()

const store = useValidatorDashboardSummaryStore()
const { getSummary } = store
const { summaryMap, queryMap } = storeToRefs(store)

const { overview } = storeToRefs(useValidatorDashboardOverview())

const expandedRows = ref<VDBSummaryTableRow[]>([])

const loadData = () => {
  getSummary(props.dashboardId, queryMap.value[props.dashboardId])
}

watch(() => props.dashboardId, () => {
  loadData()
}, { immediate: true })

const summary = computed<VDBSummaryTableResponse | undefined>(() => {
  return summaryMap.value?.[props.dashboardId]
})

const data = computed<VDBSummaryTableRow[]>(() => {
  return summary.value?.data || []
})

const mapGroup = (groupId?: number) => {
  if (groupId === undefined || groupId < 0) {
    return $t('dashboard.validator.summary.total_group_name')
  }
  const group = overview.value?.groups?.find(g => g.id === groupId)
  if (!group) {
    return groupId
  }
  return `${group.name} (${$t('common.id')}: ${groupId})`
}

</script>
<template>
  <DataTable
    v-model:expandedRows="expandedRows"
    lazy
    paginator
    :total-records="1000"
    :page-link-size="10"
    :rows="5"
    :value="data"
    data-key="group_id"
  >
    <Column expander class="expander">
      <template #rowtogglericon="slotProps">
        <IconChevron class="toggle" :expanded="slotProps.rowExpanded" />
      </template>
    </Column>
    <Column field="group" :sortable="true" :header="$t('dashboard.validator.summary.col.group')">
      <template #body="slotProps">
        {{ mapGroup(slotProps.data.group_id) }}
      </template>
    </Column>
    <Column field="group" :sortable="true" :header="$t('dashboard.validator.summary.col.efficiency_24h')">
      <template #body="slotProps">
        <BcFormatPercent :percent="slotProps.data.efficiency_24h" :color-break-point="80" />
      </template>
    </Column>
    <Column field="group" :sortable="true" :header="$t('dashboard.validator.summary.col.efficiency_7d')">
      <template #body="slotProps">
        <BcFormatPercent :percent="slotProps.data.efficiency_7d" :color-break-point="80" />
      </template>
    </Column>
    <Column field="group" :sortable="true" :header="$t('dashboard.validator.summary.col.efficiency_31d')">
      <template #body="slotProps">
        <BcFormatPercent :percent="slotProps.data.efficiency_31d" :color-break-point="80" />
      </template>
    </Column>
    <Column field="group" :sortable="true" :header="$t('dashboard.validator.summary.col.efficiency_all')">
      <template #body="slotProps">
        <BcFormatPercent :percent="slotProps.data.efficiency_all" :color-break-point="80" />
      </template>
    </Column>
    <Column field="group" :sortable="true" :header="$t('dashboard.validator.summary.col.validators')">
      <template #body="slotProps">
        <DashboardTableValidators :validators="slotProps.data.validators" :group-id="slotProps.data.group_id" context="group" />
      </template>
    </Column>
    <template #expansion="slotProps">
      <DashboardTableSummaryDetails class="details" :row="slotProps.data" :dashboard-id="props.dashboardId" />
    </template>
  </DataTable>
</template>

<style lang="scss" scoped>
:deep(.expander){
  width: 32px;
}

.details{
  margin-left: 64px;
}
</style>
