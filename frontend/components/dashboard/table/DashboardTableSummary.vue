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

const { width, isMobile } = useWindowSize()
const colsVisible = computed(() => {
  return {
    validator: width.value >= 1400,
    efficiency_plus: width.value >= 1180
  }
})

const currentOffset = ref<number>(0)

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
  if (isMobile.value) {
    return group.name
  }
  return `${group.name} (${$t('common.id')}: ${groupId})`
}

const setOffset = (value: number) => {
  currentOffset.value = value
}

const allExpanded = computed(() => {
  return !!data.value?.every(item => expandedRows.value[item.group_id])
})

const toggleAll = () => {
  const wasExpanded = allExpanded.value
  const rows = { ...expandedRows.value }
  data.value?.forEach((item) => {
    if (wasExpanded) {
      delete rows[item.group_id]
    } else {
      rows[item.group_id] = item
    }
  })
  expandedRows.value = rows
}

</script>
<template>
  <DataTable
    v-model:expandedRows="expandedRows"
    lazy
    :total-records="1000"
    :page-link-size="10"
    :rows="5"
    :value="data"
    data-key="group_id"
    class="summary_table"
  >
    <Column expander class="expander">
      <template #header>
        <IconChevron class="toggle" :direction="allExpanded ? 'bottom' : 'right'" @click="toggleAll" />
      </template>
      <template #rowtogglericon="slotProps">
        <IconChevron class="toggle" :direction="slotProps.rowExpanded ? 'bottom' : 'right'" />
      </template>
    </Column>
    <Column field="group" body-class="bold" :sortable="true" :header="$t('dashboard.validator.summary.col.group')">
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
      <DashboardTableSummaryDetails class="details" :row="slotProps.data" :dashboard-id="props.dashboardId" />
    </template>
    <template #footer>
      <BcTableOffsetPager :page-size="5" :total-count="999" :current-offset="currentOffset" @set-offset="setOffset" />
    </template>
  </DataTable>
</template>

<style lang="scss" scoped>
:deep(.expander) {
  width: 32px;
}

.toggle {
  cursor: pointer;
}

.summary_table {

  .details {
    margin-left: 21px;
  }

  :deep(td:not(.expander)):not(:last-child),
  :deep(th:not(.expander)):not(:last-child) {
    width: 220px;
    max-width: 220px;
    min-width: 220px;
  }
}
</style>
