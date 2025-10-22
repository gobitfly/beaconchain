<script setup lang="ts">
import type { DataTableSortEvent } from 'primevue/datatable'
import { useStorage } from '@vueuse/core'
import type { VDBSummaryTableRow } from '~/types/api/validator_dashboard'
import type {
  Cursor, TableQueryParams,
} from '~/types/datatable'
import { DAHSHBOARDS_ALL_GROUPS_ID } from '~/types/dashboard'
import { getGroupLabel } from '~/utils/dashboard/group'
import {
  type SummaryChartFilter,
  type SummaryTableVisibility,
  type SummaryTimeFrame,
  SummaryTimeFrames,
} from '~/types/dashboard/summary'

type ShowAbsoluteValuesStorage = {
  [dashboardId: string]: boolean,
}

const {
  dashboardKey,
  isGuestDashboard,
  isSharedDashboard,
} = useDashboardKey()
const {
  getSummary,
  isLoading,
  query: lastQuery,
  summary,
} = useValidatorDashboardSummaryStore()
const {
  bounce: setQuery,
  temp: tempQuery,
  value: query,
} = useDebounceValue<TableQueryParams | undefined>(undefined, 500)
const validatorDashboardOverviewStore = useValidatorDashboardOverviewStore()
const {
  hasValidators,
  isLargeDashboard,
  overview,
} = storeToRefs(validatorDashboardOverviewStore)
const { groups } = useValidatorDashboardGroups()
const { width } = useWindowSize()
const storageDashboardKey = computed(() => {
  return dashboardKey.value || 'guest-dashboard'
})

const cursor = ref<Cursor>()
const pageSize = ref<number>(10)
const { t: $t } = useTranslation()
const chartFilter = ref<SummaryChartFilter>({
  aggregation: 'hourly',
  efficiency: 'all',
  groupIds: [],
})
const selectedTimeFrame = ref<SummaryTimeFrame>('last_24h')
const showAbsoluteValuesPersisted = useStorage<ShowAbsoluteValuesStorage>('bc-dashboard-table-summary-show-absolute-values', {})

const timeFrames = computed(() =>
  SummaryTimeFrames.map(t => ({
    id: t,
    name: $t(`time_frames.${t}`),
  })),
)

const colsVisible = computed<SummaryTableVisibility>(() => {
  return {
    attestations: width.value >= 1015,
    efficiency: width.value >= 730,
    proposals: width.value >= 1194,
    reward: width.value >= 933,
    validatorsSortable: width.value >= 571,
  }
})
const searchPlaceholder = computed(() =>
  $t(
    isGuestDashboard.value && (groups.value?.length ?? 0) <= 1
      ? 'dashboard.validator.summary.search_placeholder_public'
      : 'dashboard.validator.summary.search_placeholder',
  ),
)
const loadData = (q?: TableQueryParams) => {
  if (!q) {
    q = query.value
      ? { ...query.value }
      : {
          limit: pageSize.value,
          sort: 'efficiency:desc',
        }
  }
  setQuery(q, true, true)
}
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

onMounted(() => {
  if (!(storageDashboardKey.value in showAbsoluteValuesPersisted.value)) {
    showAbsoluteValuesPersisted.value[storageDashboardKey.value] = !isSharedDashboard.value || !isLargeDashboard.value
  }
})

watch(() => overview.value, () => {
  if (!(storageDashboardKey.value in showAbsoluteValuesPersisted.value)) {
    showAbsoluteValuesPersisted.value[storageDashboardKey.value] = !isSharedDashboard.value || !isLargeDashboard.value
  }
})
watch(
  [
    dashboardKey,
    overview,
  ],
  () => {
    loadData()
  },
  { immediate: true },
)
watch(
  [
    query,
    selectedTimeFrame,
  ],
  ([
    q,
    timeFrame,
  ]) => {
    if (q) {
      getSummary(dashboardKey.value, timeFrame, q)
    }
  },
  { immediate: true },
)
</script>

<template>
  <div>
    <BcTableControl
      v-model:="showAbsoluteValuesPersisted[storageDashboardKey]"
      :search-placeholder
      @set-search="setSearch"
    >
      <template #header-center="{ tableIsShown }">
        <h1 class="summary_title">
          {{ $t("dashboard.validator.summary.title") }}
        </h1>
        <BcDropdown
          v-if="tableIsShown"
          v-model="selectedTimeFrame"
          :options="timeFrames"
          option-value="id"
          option-label="name"
          class="small"
          :placeholder="$t('dashboard.group.selection.placeholder')"
        />
        <DashboardChartSummaryFilter
          v-else
          v-model="chartFilter"
        />
      </template>
      <template #table>
        <ClientOnly fallback-tag="span">
          <BcTable
            :data="summary"
            data-key="group_id"
            :expandable="true"
            class="summary_table"
            :cursor
            :page-size
            :row-class="getRowClass"
            :selected-sort="tempQuery?.sort"
            :is-loading
            :hide-pager="true"
            @set-cursor="setCursor"
            @sort="onSort"
            @set-page-size="setPageSize"
          >
            <PvColumn
              field="group_id"
              :sortable="true"
              body-class="group-id-column bold"
              header-class="group-id-column"
              :header="$t('dashboard.validator.col.group')"
            >
              <template #body="slotProps">
                {{ groupNameLabel(slotProps.data.group_id) }}
              </template>
            </PvColumn>
            <PvColumn
              field="status"
              header-class="status-column"
              body-class="status-column"
              :header="$t('dashboard.validator.col.status')"
            >
              <template #body="slotProps">
                <DashboardTableSummaryStatus
                  :class="slotProps.data.className"
                  :status="slotProps.data.status"
                />
              </template>
            </PvColumn>
            <PvColumn
              field="validators"
              body-class="validator-column"
              header-class="validator-column"
              :sortable="colsVisible.validatorsSortable"
            >
              <template #header>
                <div class="validators-header">
                  <div>{{ $t("dashboard.validator.col.validators") }}</div>
                  <div class="sub-header">
                    {{ $t("common.live") }}
                  </div>
                  <BcTooltip
                    class="info"
                    tooltip-class="summary-info-tooltip"
                    :text="$t('dashboard.validator.summary.tooltip.live')"
                    @click.stop.prevent="() => {}"
                  >
                    <BcIcon name="circle-info" />
                  </BcTooltip>
                </div>
              </template>
              <template #body="{ data }">
                <DashboardTableSummaryValidators
                  :validators="data.validators"
                  :is-absolute="showAbsoluteValuesPersisted[storageDashboardKey] ?? false"
                  :row="data"
                  :group-id="data.group_id"
                  :dashboard-key
                  :time-frame="selectedTimeFrame"
                  context="group"
                />
              </template>
            </PvColumn>
            <PvColumn
              v-if="colsVisible.efficiency"
              field="efficiency"
              :sortable="true"
              body-class="efficiency-column"
            >
              <template #header>
                <div class="validators-header">
                  <div>{{ $t("dashboard.validator.col.beaconscore") }}</div>
                  <BcTooltip
                    class="info"
                    tooltip-class="summary-info-tooltip"
                    @click.stop
                  >
                    <template #tooltip>
                      <BcTranslation
                        keypath="dashboard.beaconscore.template"
                        linkpath="dashboard.beaconscore.link"
                        :to="externalLink.knowledgeBase.beaconScore"
                      />
                    </template>
                    <BcIcon name="circle-info" />
                  </BcTooltip>
                </div>
              </template>
              <template #body="slotProps">
                <DashboardTableSummaryValue
                  :class="slotProps.data.className"
                  property="efficiency"
                  :time-frame="selectedTimeFrame"
                  :row="slotProps.data"
                />
              </template>
            </PvColumn>
            <PvColumn
              v-if="colsVisible.attestations"
              field="attestations"
              :sortable="true"
              :header="$t('dashboard.validator.summary.row.attestations')"
            >
              <template #body="slotProps">
                <DashboardTableSummaryValue
                  :class="slotProps.data.className"
                  property="attestations"
                  :absolute="showAbsoluteValuesPersisted[storageDashboardKey] ?? true"
                  :time-frame="selectedTimeFrame"
                  :row="slotProps.data"
                />
              </template>
            </PvColumn>
            <PvColumn
              v-if="colsVisible.proposals"
              field="proposals"
              :sortable="true"
              :header="$t('dashboard.validator.summary.row.proposals')"
            >
              <template #body="slotProps">
                <DashboardTableSummaryValue
                  :class="slotProps.data.className"
                  property="proposals"
                  class="no-space-between-value"
                  :absolute="showAbsoluteValuesPersisted[storageDashboardKey] ?? true"
                  :time-frame="selectedTimeFrame"
                  :row="slotProps.data"
                />
              </template>
            </PvColumn>
            <PvColumn
              v-if="colsVisible.reward"
              field="reward"
              :sortable="true"
              :header="$t('dashboard.validator.col.rewards')"
            >
              <template #body="slotProps">
                <DashboardTableSummaryValue
                  :class="slotProps.data.className"
                  property="reward"
                  class="no-space-between-value"
                  :absolute="showAbsoluteValuesPersisted[storageDashboardKey] ?? true"
                  :time-frame="selectedTimeFrame"
                  :row="slotProps.data"
                />
              </template>
            </PvColumn>
            <template #expansion="slotProps">
              <DashboardTableSummaryDetails
                :table-visibility="colsVisible"
                :row="slotProps.data"
                :time-frame="selectedTimeFrame"
                :absolute="showAbsoluteValuesPersisted[storageDashboardKey] ?? true"
              />
            </template>
            <template #empty>
              <DashboardTableAddValidator v-if="!hasValidators" />
            </template>
          </BcTable>
        </ClientOnly>
      </template>
      <template #chart>
        <div class="chart-container">
          <DashboardChartSummary :filter="chartFilter" />
        </div>
      </template>
    </BcTableControl>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/utils.scss";

.summary_title {
  @media (max-width: 600px) {
    display: none;
  }
}

.sub-header {
  color: var(--text-color-disabled);
  font-size: var(--tiny_text_font_size);
}

.no-space-between-value {
  justify-content: unset;
  gap: var(--padding);
}

.validators-header {
  .info {
    position: absolute;
    top: 16px;
    right: var(--padding-large);

    svg {
      width: 14px;
      height: 14px;
    }
  }

  @media (min-width: 730px) {
    position: relative;

    .info {
      top: calc(50% - 9px);
      right: -50px;
    }
  }
}

:global(.summary-info-tooltip .bc-tooltip) {
  width: 120px;
}

:deep(.summary_table) {
  > .p-datatable-wrapper {
    min-height: 529px;
  }

  .group-id-column {
    @include utils.truncate-text;
    @include utils.set-all-width(200px);

    @media (max-width: 570px) {
      @include utils.set-all-width(80px);
    }
  }

  .status-column {
    @include utils.set-all-width(90px);
  }

  .status-column,
  .efficiency-column {
    padding: 7px !important;
    min-width: 170px;
  }

  .validator-column {
    @include utils.set-all-width(240px);
    padding: 3px 7px !important;

    @media (max-width: 570px) {
      @include utils.set-all-width(120px);
    }
  }

  .total-row {
    &:not(:has(.bottom)) {
      td {
        border-bottom-color: var(--primary-color);
      }
    }
  }

  .total-row + .p-datatable-row-expansion {
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
