<script setup lang="ts">
import type { DataTableSortEvent } from 'primevue/datatable'
import type { VDBRewardsTableRow } from '~/types/api/validator_dashboard'
import type {
  Cursor,
  TableQueryParams,
} from '~/types/datatable'
import {
  DAHSHBOARDS_ALL_GROUPS_ID,
  DAHSHBOARDS_NEXT_EPOCH_ID,
  type TableProps,
} from '~/types/dashboard'
import { totalElCl } from '~/utils/bigMath'
import { getGroupLabel } from '~/utils/dashboard/group'
import { formatRewardValueOption } from '~/utils/dashboard/table'

const {
  isGuestDashboard,
} = useDashboardKey()

const cursor = ref<Cursor>()
const pageSize = ref<number>(10)
const { t: $t } = useTranslation()

const props = defineProps<TableProps<VDBRewardsTableRow>>()
const emit = defineEmits<{
  (e: 'update', query: TableQueryParams): void,
}>()
const emitUpdate = (query: TableQueryParams) => {
  emit('update', query)
}

const {
  groups, hasValidators,
} = storeToRefs(useValidatorDashboardStore())

const { width } = useWindowSize()
const colsVisible = computed(() => {
  return {
    age: width.value >= 660,
    clRewards: width.value >= 900,
    duty: width.value > 1180,
    elRewards: width.value >= 780,
  }
})

const groupNameLabel = (groupId?: number) => {
  return getGroupLabel($t, groupId, groups.value, 'Σ')
}

const onSort = (sort: DataTableSortEvent) => {
  emitUpdate(setQuerySort(sort, props.query))
}

const setCursor = (value: Cursor) => {
  cursor.value = value
  emitUpdate(setQueryCursor(value, props.query))
}

const setPageSize = (value: number) => {
  pageSize.value = value
  emitUpdate(setQueryPageSize(value, props.query))
}

const setSearch = (value?: string) => {
  emitUpdate(setQuerySearch(value, props.query))
}

const getRowClass = (row: VDBRewardsTableRow) => {
  if (row.group_id === DAHSHBOARDS_ALL_GROUPS_ID) {
    return 'total-row'
  }
  else if (row.group_id === DAHSHBOARDS_NEXT_EPOCH_ID) {
    return 'future-row'
  }
}

const isRowExpandable = (row: VDBRewardsTableRow) => {
  return row.group_id !== DAHSHBOARDS_NEXT_EPOCH_ID
}
</script>

<template>
  <div>
    <BcTableControl
      :title="$t('dashboard.validator.rewards.title')"
      :search-placeholder="
        $t(
          isGuestDashboard
            ? 'dashboard.validator.rewards.search_placeholder_public'
            : 'dashboard.validator.rewards.search_placeholder',
        )
      "
      @set-search="setSearch"
    >
      <template #table>
        <ClientOnly fallback-tag="span">
          <BcTable
            :data="addIdentifier({ data, paging }, 'epoch', 'group_id')"
            data-key="identifier"
            :expandable="true"
            class="rewards-table"
            :cursor
            :page-size
            :row-class="getRowClass"
            :add-spacer="colsVisible.age"
            :is-row-expandable
            :selected-sort="query.sort"
            :loading="isLoading"
            @set-cursor="setCursor"
            @sort="onSort"
            @set-page-size="setPageSize"
          >
            <Column
              field="epoch"
              :sortable="true"
              body-class="epoch"
              header-class="epoch"
              :header="$t('common.epoch')"
            >
              <template #body="slotProps">
                <BcLink
                  :to="`/epoch/${slotProps.data.epoch}`"
                  class="link"
                  target="_blank"
                >
                  <BcFormatNumber :value="slotProps.data.epoch" />
                </BcLink>
              </template>
            </Column>
            <Column
              v-if="colsVisible.age"
              field="age"
              body-class="age-field"
            >
              <template #header>
                <BcTableAgeHeader />
              </template>
              <template #body="slotProps">
                <BcFormatTimePassed :value="slotProps.data.epoch" />
              </template>
            </Column>
            <Column
              v-if="colsVisible.duty"
              field="duty"
              body-class="duty"
              header-class="duty"
              :header="$t('dashboard.validator.col.duty')"
            >
              <template #body="slotProps">
                <DashboardTableValueDuty
                  :is-number-visible="slotProps.data.group_id !== DAHSHBOARDS_NEXT_EPOCH_ID"
                  :duty="slotProps.data.duty"
                />
              </template>
            </Column>
            <Column
              field="group_id"
              body-class="group-id"
              header-class="group-id"
              :header="$t('dashboard.validator.col.group')"
            >
              <template #body="slotProps">
                <span>
                  {{ groupNameLabel(slotProps.data.group_id) }}
                </span>
              </template>
            </Column>
            <Column
              field="reward"
              body-class="reward"
              header-class="reward"
              :header="$t('dashboard.validator.summary.row.reward')"
            >
              <template #body="slotProps">
                <div
                  v-if="slotProps.data.group_id === DAHSHBOARDS_NEXT_EPOCH_ID"
                >
                  -
                </div>
                <BcFormatValue
                  v-else
                  :value="totalElCl(slotProps.data.reward)"
                  :use-colors="true"
                  :options="formatRewardValueOption"
                />
              </template>
            </Column>
            <Column
              v-if="colsVisible.elRewards"
              field="reward_el"
              body-class="reward"
              header-class="reward"
              :header="$t('dashboard.validator.col.el_rewards')"
            >
              <template #body="slotProps">
                <div
                  v-if="slotProps.data.group_id === DAHSHBOARDS_NEXT_EPOCH_ID"
                >
                  -
                </div>
                <BcFormatValue
                  v-else
                  :value="slotProps.data.reward?.el"
                  :use-colors="true"
                  :options="formatRewardValueOption"
                />
              </template>
            </Column>
            <Column
              v-if="colsVisible.clRewards"
              field="reward_cl"
              body-class="reward"
              header-class="reward"
              :header="$t('dashboard.validator.col.cl_rewards')"
            >
              <template #body="slotProps">
                <div
                  v-if="slotProps.data.group_id === DAHSHBOARDS_NEXT_EPOCH_ID"
                >
                  -
                </div>
                <BcFormatValue
                  v-else
                  :value="slotProps.data.reward?.cl"
                  :use-colors="true"
                  :options="formatRewardValueOption"
                />
              </template>
            </Column>
            <template #expansion="slotProps">
              <DashboardTableRewardsDetails
                :row="slotProps.data"
                :group-name="groupNameLabel(slotProps.data.group_id)"
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
          <DashboardChartRewards />
        </div>
      </template>
    </BcTableControl>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/utils.scss";

:deep(.rewards-table) {
  > .p-datatable-wrapper {
    min-height: 577px;
  }

  .epoch {
    @include utils.set-all-width(84px);
  }

  .group-id {
    @include utils.set-all-width(120px);
    @include utils.truncate-text;

    @media (max-width: 450px) {
      @include utils.set-all-width(80px);
    }
  }

  .reward {
    @include utils.set-all-width(154px);
  }

  .age-field {
    white-space: nowrap;
  }
  tr > td.age-field {
    padding: 0 7px;
    @include utils.set-all-width(205px);
  }

  tr:not(.p-datatable-row-expansion) {
    @media (max-width: 1300px) {
      .duty {
        @include utils.set-all-width(300px);
      }
    }
  }

  tr:has(+ .total-row) {
    td {
      border-bottom-color: var(--primary-color);
    }
  }

  .future-row {
    td {
      > div,
      > span {
        opacity: 0.5;
      }
    }
  }
}

.chart-container {
  width: 100%;
  height: 625px;
}
</style>
