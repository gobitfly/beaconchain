<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import {
  faInfoCircle
} from '@fortawesome/pro-regular-svg-icons'
import type { DataTableSortEvent } from 'primevue/datatable'
import type { VDBSummaryTableRow } from '~/types/api/validator_dashboard'
import type { Cursor, TableQueryParams } from '~/types/datatable'
import { useValidatorDashboardOverviewStore } from '~/stores/dashboard/useValidatorDashboardOverviewStore'
import { DAHSHBOARDS_ALL_GROUPS_ID } from '~/types/dashboard'
import { getGroupLabel } from '~/utils/dashboard/group'
import { SummaryTimeFrames, type SummaryChartFilter, type SummaryTableVisibility, type SummaryTimeFrame } from '~/types/dashboard/summary'

const { dashboardKey, isPublic } = useDashboardKey()

const cursor = ref<Cursor>()
const pageSize = ref<number>(10)
const { t: $t } = useI18n()
const showInDevelopment = Boolean(useRuntimeConfig().public.showInDevelopment)
const chartFilter = ref<SummaryChartFilter>({ aggregation: 'hourly', efficiency: 'all', groupIds: [] })

const { summary, query: lastQuery, isLoading, getSummary } = useValidatorDashboardSummaryStore()
const { value: query, temp: tempQuery, bounce: setQuery } = useDebounceValue<TableQueryParams | undefined>(undefined, 500)

const showAbsoluteValues = ref<boolean | null>(null)

const { overview, hasValidators, validatorCount } = useValidatorDashboardOverviewStore()
const { groups } = useValidatorDashboardGroups()

const timeFrames = computed(() => SummaryTimeFrames.filter(t => showInDevelopment || t !== 'last_1h').map(t => ({ name: $t(`time_frames.${t}`), id: t })))
const selectedTimeFrame = ref<SummaryTimeFrame>('last_24h')

const { width } = useWindowSize()
const colsVisible = computed<SummaryTableVisibility>(() => {
  return {
    proposals: width.value >= 1194,
    attestations: width.value >= 1015,
    reward: width.value >= 933,
    efficiency: width.value >= 730,
    validatorsSortable: width.value >= 571
  }
})
const loadData = (q?: TableQueryParams) => {
  if (!q) {
    q = query.value ? { ...query.value } : { limit: pageSize.value, sort: 'efficiency:desc' }
  }
  setQuery(q, true, true)
}

watch(validatorCount, (count) => {
  if (count !== undefined && showAbsoluteValues.value === null) {
    showAbsoluteValues.value = count < 100_000
  }
}, { immediate: true })

watch([dashboardKey, overview], () => {
  loadData()
}, { immediate: true })

watch([query, selectedTimeFrame], ([q, timeFrame]) => {
  if (q) {
    getSummary(dashboardKey.value, timeFrame, q)
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

const searchPlaceholder = computed(() => $t(isPublic.value && (groups.value?.length ?? 0) <= 1 ? 'dashboard.validator.summary.search_placeholder_public' : 'dashboard.validator.summary.search_placeholder'))

</script>
<template>
  <div>
    <BcTableControl
      v-model:="showAbsoluteValues"
      :search-placeholder="searchPlaceholder"
      @set-search="setSearch"
    >
      <template #header-center="{tableIsShown}">
        <h1 class="summary_title">
          {{ $t('dashboard.validator.summary.title') }}
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
        <DashboardChartSummaryChartFilter v-else v-model="chartFilter" />
      </template>
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
            :selected-sort="tempQuery?.sort"
            :loading="isLoading"
            :hide-pager="true"
            @set-cursor="setCursor"
            @sort="onSort"
            @set-page-size="setPageSize"
          >
            <Column
              field="group_id"
              :sortable="true"
              body-class="group-id-column bold"
              header-class="group-id-column"
              :header="$t('dashboard.validator.col.group')"
            >
              <template #body="slotProps">
                {{ groupNameLabel(slotProps.data.group_id) }}
              </template>
            </Column>
            <Column
              field="status"
              header-class="status-column"
              body-class="status-column"
              :header="$t('dashboard.validator.col.status')"
            >
              <template #body="slotProps">
                <DashboardTableSummaryStatus :class="slotProps.data.className" :status="slotProps.data.status" />
              </template>
            </Column>
            <Column
              field="validators"
              body-class="validator-column"
              header-class="validator-column"
              :sortable="colsVisible.validatorsSortable"
            >
              <template #header>
                <div class="validators-header">
                  <div>{{ $t('dashboard.validator.col.validators') }}</div>
                  <div class="sub-header">
                    {{ $t('common.live') }}
                  </div>
                  <BcTooltip
                    class="info"
                    tooltip-class="summary-info-tooltip"
                    :text="$t('dashboard.validator.summary.tooltip.live')"
                    @click.stop.prevent="() => { }"
                  >
                    <FontAwesomeIcon :icon="faInfoCircle" />
                  </BcTooltip>
                </div>
              </template>
              <template #body="slotProps">
                <DashboardTableSummaryValidators
                  :absolute="showAbsoluteValues ?? true"
                  :row="slotProps.data"
                  :group-id="slotProps.data.group_id"
                  :dashboard-key="dashboardKey"
                  :time-frame="selectedTimeFrame"
                  context="group"
                />
              </template>
            </Column>
            <Column
              v-if="colsVisible.efficiency"
              field="efficiency"
              :sortable="true"
              body-class="efficiency-column"
              :header="$t('dashboard.validator.col.efficiency')"
            >
              <template #body="slotProps">
                <DashboardTableSummaryValue
                  :class="slotProps.data.className"
                  property="efficiency"
                  :time-frame="selectedTimeFrame"
                  :row="slotProps.data"
                />
              </template>
            </Column>
            <Column
              v-if="colsVisible.attestations"
              field="attestations"
              :sortable="true"
              :header="$t('dashboard.validator.summary.row.attestations')"
            >
              <template #body="slotProps">
                <DashboardTableSummaryValue
                  :class="slotProps.data.className"
                  property="attestations"
                  :absolute="showAbsoluteValues ?? true"
                  :time-frame="selectedTimeFrame"
                  :row="slotProps.data"
                />
              </template>
            </Column>
            <Column
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
                  :absolute="showAbsoluteValues ?? true"
                  :time-frame="selectedTimeFrame"
                  :row="slotProps.data"
                />
              </template>
            </Column>
            <Column
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
                  :absolute="showAbsoluteValues ?? true"
                  :time-frame="selectedTimeFrame"
                  :row="slotProps.data"
                />
              </template>
            </Column>
            <template #expansion="slotProps">
              <DashboardTableSummaryDetails
                :table-visibility="colsVisible"
                :row="slotProps.data"
                :time-frame="selectedTimeFrame"
                :absolute="showAbsoluteValues ?? true"
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
          <DashboardChartSummaryChart :filter="chartFilter" />
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
      top: 8px;
      right: -50px;
    }
  }
}

:global(.summary-info-tooltip .bc-tooltip) {
  width: 120px;
}

:deep(.summary_table) {

  >.p-datatable-wrapper {
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
