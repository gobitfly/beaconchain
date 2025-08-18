<script setup lang="ts">
// import type { DataTableSortEvent } from 'primevue/datatable'
// import { useStorage } from '@vueuse/core'
import type { VDBSummaryTableRow } from '~/types/api/validator_dashboard'
import type {
  Cursor,
  //  TableQueryParams,
} from '~/types/datatable'
import { DAHSHBOARDS_ALL_GROUPS_ID } from '~/types/dashboard'
import { getGroupLabel } from '~/utils/dashboard/group'
import type {
  SummaryChartFilter,
  SummaryTableVisibility,
  // type SummaryTimeFrame,
  // SummaryTimeFrames,
} from '~/types/dashboard/summary'

// type ShowAbsoluteValuesStorage = {
//   [dashboardId: string]: boolean,
// }

const {
  hasValidators,
  key,
  variant,
} = useDashboard()
// const {
// getSummary,
// isLoading,
// query: lastQuery,
// summary,
// } = useValidatorDashboardSummaryStore()
// const {
//   bounce: setQuery,
//   temp: tempQuery,
//   value: query,
// } = useDebounceValue<TableQueryParams | undefined>(undefined, 500)
// const validatorDashboardOverviewStore = useValidatorDashboardOverviewStore()
// const {
// hasValidators,
// isLargeDashboard,
// overview,
// } = storeToRefs(validatorDashboardOverviewStore)
const { groups } = useValidatorDashboardGroups()
const { width } = useWindowSize()
// const storageDashboardKey = computed(() => {
//   if (variant.value === 'guest-dashboard') {
//     return 'guest-dashboard'
//   }
//   return key.value
// })

const cursor = ref<Cursor>()
const pageSize = ref<number>(10)
const { t: $t } = useTranslation()
const chartFilter = ref<SummaryChartFilter>({
  aggregation: 'hourly',
  efficiency: 'all',
  groupIds: [],
})

const timeFrames: { id: Query['period'], name: string }[] = [
  {
    id: 'last_1h',
    name: $t('time_frames.last_1h'),
  },
  {
    id: 'last_24h',
    name: $t('time_frames.last_24h'),
  },
  {
    id: 'last_7d',
    name: $t('time_frames.last_7d'),
  },
  {
    id: 'last_30d',
    name: $t('time_frames.last_30d'),
  },
  {
    id: 'all_time',
    name: $t('time_frames.all_time'),
  },
]

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
    variant.value === 'guest-dashboard' && (groups.value?.length ?? 0) <= 1
      ? 'dashboard.validator.summary.search_placeholder_public'
      : 'dashboard.validator.summary.search_placeholder',
  ),
)

const groupNameLabel = (groupId?: number) => {
  return getGroupLabel($t, groupId, groups.value, 'Σ')
}

const getRowClass = (row: VDBSummaryTableRow) => {
  if (row.group_id === DAHSHBOARDS_ALL_GROUPS_ID) {
    return 'total-row'
  }
}

const emit = defineEmits<{
  (e: 'add-validator'): void,
}>()

const id = computed(() => {
  if (variant.value === 'guest-dashboard') return 'guest-dashboard'
  if (variant.value === 'shared-dashboard') return 'shared-dashboard'
  return key.value as string
})

const summaryTableNumberFormat = useBcCookie<Record<string, 'absolute' | 'relative'>>('bc-summary-table-number-format', {
  default() {
    return {
      [id.value]: 'absolute',
    }
  },
})
const summaryTabView = useBcCookie<'chart' | 'table'>('bc-summary-tab-view', {
  default() {
    return 'table'
  },
})
const query = useDefaultQuery({ period: 'last_24h' })

// const abortController = new AbortController()
const {
  data,
  status,
} = useApi(() => `/api/bff/validator-dashboards/${key.value}/summary`, {
  immediate: !!key.value,
  query,
})
</script>

<template>
  <div>
    <BcTableControl
      v-model:search="query.search"
      v-model:tab-view="summaryTabView"
      :search-placeholder
    >
      <template #header-center="{ isTable }">
        <h1 class="summary_title">
          {{ $t("dashboard.validator.summary.title") }}
        </h1>
        <BcDropdown
          v-if="isTable"
          v-model="query.period"
          :options="timeFrames"
          option-value="id"
          option-label="name"
          class="small"
          :placeholder="$t('dashboard.group.selection.placeholder')"
        />
        <LazyDashboardChartSummaryFilter
          v-else
          v-model="chartFilter"
        />
      </template>
      <template #value-format>
        <BcToggleIcon
          v-model="summaryTableNumberFormat[id]"
          true-value="absolute"
          false-value="relative"
        >
          <template #trueIcon>
            <BcIcon
              size="sm"
              name="hashtag"
            />
          </template>
          <template #falseIcon>
            <BcIcon
              size="sm"
              name="percent"
            />
          </template>
        </BcToggleIcon>
      </template>
      <template #table>
        <ClientOnly fallback-tag="span">
          <BcTable
            :query
            :data
            data-key="group_id"
            :expandable="true"
            class="summary_table"
            :cursor
            :page-size
            :row-class="getRowClass"
            :is-loading="status === 'pending'"
            :hide-pager="true"
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
                <DashboardTableSummaryStatus
                  :class="slotProps.data.className"
                  :status="slotProps.data.status"
                />
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
              <template #body="slotProps">
                <DashboardTableSummaryValidators
                  :validators="slotProps.data.validators"
                  :is-absolute="summaryTableNumberFormat[id] === 'absolute'"
                  :row="slotProps.data"
                  :group-id="slotProps.data.group_id"
                  :dashboard-key="key"
                  :time-frame="query.period"
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
                  :time-frame="query.period"
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
                  :is-absolute="summaryTableNumberFormat[id] === 'absolute'"
                  :time-frame="query.period"
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
                  :is-absolute="summaryTableNumberFormat[id] === 'absolute'"
                  :time-frame="query.period"
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
                  :is-absolute="summaryTableNumberFormat[id] === 'absolute'"
                  :time-frame="query.period"
                  :row="slotProps.data"
                />
              </template>
            </Column>
            <template #expansion="slotProps">
              <DashboardTableSummaryDetails
                :table-visibility="colsVisible"
                :row="slotProps.data"
                :time-frame="query.period"
                :is-absolute="summaryTableNumberFormat[id] === 'absolute'"
              />
            </template>
            <template #empty>
              <LazyDashboardTableAddValidator
                v-if="!hasValidators"
                @add-validator="emit('add-validator')"
              />
            </template>
          </BcTable>
          <!-- </ClientOnly> -->
        </clientonly>
      </template>
      <template #chart="{ isTable }">
        <div class="chart-container">
          <LazyDashboardChartSummary
            v-if="!isTable"
            :filter="chartFilter"
          />
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
