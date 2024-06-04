<script setup lang="ts">
import type { DataTableSortEvent } from 'primevue/datatable'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faInfoCircle } from '@fortawesome/pro-regular-svg-icons'
import type { VDBWithdrawalsTableRow } from '~/types/api/validator_dashboard'
import type { Cursor, TableQueryParams } from '~/types/datatable'
import { useValidatorDashboardWithdrawalsStore } from '~/stores/dashboard/useValidatorDashboardWithdrawalsStore'
import { BcFormatHash } from '#components'
import { getGroupLabel } from '~/utils/dashboard/group'

type ExtendedVDBWithdrawalsTableRow = VDBWithdrawalsTableRow & {identifier: string}

const { dashboardKey } = useDashboardKey()

const cursor = ref<Cursor>()
const pageSize = ref<number>(10)
const { t: $t } = useI18n()

const { latestState } = useLatestStateStore()
const { slotToEpoch } = useNetwork()
const { withdrawals, query: lastQuery, getWithdrawals, totalAmount, getTotalAmount, isLoadingWithdrawals, isLoadingTotal } = useValidatorDashboardWithdrawalsStore()
const { value: query, temp: tempQuery, bounce: setQuery } = useDebounceValue<TableQueryParams | undefined>(undefined, 500)
const totalIdentifier = 'total'

const { groups } = useValidatorDashboardGroups()

const { width } = useWindowSize()
const colsVisible = computed(() => {
  return {
    group: width.value > 995,
    slot: width.value > 875,
    epoch: width.value > 805,
    recipient: width.value > 695,
    amount: width.value > 500
  }
})

const loadData = (query?: TableQueryParams) => {
  if (!query) {
    query = { limit: pageSize.value, sort: 'slot:desc' }
  }
  setQuery(query, true, true)
}

watch(dashboardKey, () => {
  loadData()
  getTotalAmount(dashboardKey.value)
}, { immediate: true })

watch(query, (q) => {
  if (q) {
    getWithdrawals(dashboardKey.value, q)
  }
}, { immediate: true })

const tableData = computed(() => {
  if (!withdrawals.value?.data?.length) {
    return
  }

  return {
    paging: withdrawals.value.paging,
    data: [
      {
        amount: totalAmount.value,
        identifier: totalIdentifier
      },
      ...withdrawals.value.data.map(w => ({
        ...w,
        identifier: `${w.slot}-${w.index}`
      }))
    ]
  }
})

const groupNameLabel = (groupId?: number) => {
  return getGroupLabel($t, groupId, groups.value, '')
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

const getRowClass = (row: ExtendedVDBWithdrawalsTableRow) => {
  if (row.identifier === totalIdentifier) {
    return 'total-row'
  }

  if (isRowInFuture(row) || row.is_missing_estimate) {
    return 'gray-out'
  }
}

const getExpansionValueClass = (row: ExtendedVDBWithdrawalsTableRow) => {
  if (isRowInFuture(row)) {
    return 'gray-out'
  }
}

const isRowExpandable = (row: ExtendedVDBWithdrawalsTableRow) => {
  return row.identifier !== totalIdentifier
}

const isRowInFuture = (row: ExtendedVDBWithdrawalsTableRow) => {
  if (latestState?.value) {
    return row.epoch > slotToEpoch(latestState.value.current_slot)
  }

  return false
}

</script>
<template>
  <div>
    <BcTableControl
      :title="$t('dashboard.validator.withdrawals.title')"
      :search-placeholder="$t('dashboard.validator.withdrawals.search_placeholder')"
      @set-search="setSearch"
    >
      <template #table>
        <ClientOnly fallback-tag="span">
          <BcTable
            :data="tableData"
            data-key="identifier"
            :expandable="!colsVisible.group"
            class="withdrawal-table"
            :cursor="cursor"
            :page-size="pageSize"
            :row-class="getRowClass"
            :add-spacer="true"
            :is-row-expandable="isRowExpandable"
            :loading="isLoadingWithdrawals"
            :selected-sort="tempQuery?.sort"
            @set-cursor="setCursor"
            @sort="onSort"
            @set-page-size="setPageSize"
          >
            <Column
              field="index"
              :sortable="true"
              :header="$t('dashboard.validator.col.index')"
              body-class="index"
              header-class="index"
            >
              <template #body="slotProps">
                <div v-if="slotProps.data.is_missing_estimate" class="value-with-tooltip-container">
                  {{ $t('dashboard.validator.withdrawals.pending') }}
                  <BcTooltip>
                    <FontAwesomeIcon :icon="faInfoCircle" />
                    <template #tooltip>
                      {{ $t('dashboard.validator.withdrawals.pending_tooltip') }}
                    </template>
                  </BcTooltip>
                </div>
                <BcLink
                  v-else-if="slotProps.data.identifier !== totalIdentifier"
                  :to="`/validator/${slotProps.data.index}`"
                  target="_blank"
                  class="link"
                >
                  <BcFormatNumber :value="slotProps.data.index" default="-" />
                </BcLink>
                <div v-else class="all-time-total">
                  Σ
                </div>
              </template>
            </Column>
            <Column
              v-if="colsVisible.group"
              field="group_id"
              body-class="group-id"
              header-class="group-id"
              :header="$t('dashboard.validator.col.group')"
            >
              <template #body="slotProps">
                <span v-if="slotProps.data.identifier !== totalIdentifier && !slotProps.data.is_missing_estimate">
                  {{ groupNameLabel(slotProps.data.group_id) }}
                </span>
              </template>
            </Column>
            <Column
              v-if="colsVisible.epoch"
              field="epoch"
              :sortable="true"
              :header="$t('common.epoch')"
            >
              <template #body="slotProps">
                <BcLink
                  v-if="slotProps.data.identifier !== totalIdentifier && !slotProps.data.is_missing_estimate"
                  :to="`/epoch/${slotProps.data.epoch}`"
                  target="_blank"
                  class="link"
                >
                  <BcFormatNumber :value="slotProps.data.epoch" default="-" />
                </BcLink>
              </template>
            </Column>
            <Column
              v-if="colsVisible.slot"
              field="slot"
              :sortable="true"
              :header="$t('common.slot')"
            >
              <template #body="slotProps">
                <BcLink
                  v-if="slotProps.data.identifier !== totalIdentifier && !slotProps.data.is_missing_estimate"
                  :to="`/slot/${slotProps.data.slot}`"
                  target="_blank"
                  class="link"
                >
                  <BcFormatNumber :value="slotProps.data.slot" default="-" />
                </BcLink>
              </template>
            </Column>
            <Column field="age">
              <template #header>
                <BcTableAgeHeader />
              </template>
              <template #body="slotProps">
                <BcFormatTimePassed
                  v-if="slotProps.data.identifier !== totalIdentifier && !slotProps.data.is_missing_estimate"
                  type="slot"
                  class="time-passed"
                  :value="slotProps.data.slot"
                />
              </template>
            </Column>
            <Column
              v-if="colsVisible.recipient"
              field="recipient"
              :sortable="true"
              :header="$t('dashboard.validator.col.recipient')"
            >
              <template #body="slotProps">
                <div v-if="slotProps.data.identifier !== totalIdentifier && !slotProps.data.is_missing_estimate">
                  <BcFormatHash
                    v-if="slotProps.data.recipient?.hash"
                    type="address"
                    class="recipient"
                    :hash="slotProps.data.recipient?.hash"
                    :ens="slotProps.data.recipient?.ens"
                  />
                  <span v-else>-</span>
                </div>
              </template>
            </Column>
            <Column
              v-if="colsVisible.amount"
              field="amount"
              :sortable="true"
              body-class="amount"
              header-class="amount"
              :header="$t('dashboard.validator.col.amount')"
            >
              <template #body="slotProps">
                <div v-if="slotProps.data.identifier === totalIdentifier && isLoadingTotal">
                  <BcLoadingSpinner :loading="true" size="small" />
                </div>
                <div v-else-if="!slotProps.data.is_missing_estimate" class="value-with-tooltip-container">
                  <BcFormatValue
                    :value="slotProps.data.amount"
                    :class="{'all-time-total':slotProps.data.identifier === totalIdentifier}"
                  />
                  <BcTooltip v-if="isRowInFuture(slotProps.data)">
                    <FontAwesomeIcon :icon="faInfoCircle" />
                    <template #tooltip>
                      {{ $t('dashboard.validator.withdrawals.future_tooltip') }}
                    </template>
                  </BcTooltip>
                </div>
              </template>
            </Column>
            <template #expansion="slotProps">
              <div class="expansion">
                <div class="row">
                  <div class="label">
                    {{ $t('dashboard.validator.col.group') }}:
                  </div>
                  <div :class="getExpansionValueClass(slotProps.data)">
                    {{ groupNameLabel(slotProps.data.group_id) }}
                  </div>
                </div>
                <div class="row">
                  <div class="label">
                    {{ $t('common.epoch') }}:
                  </div>
                  <BcLink :to="`/epoch/${slotProps.data.epoch}`" target="_blank" class="link" :class="getExpansionValueClass(slotProps.data)">
                    <BcFormatNumber :value="slotProps.data.epoch" default="-" />
                  </BcLink>
                </div>
                <div class="row">
                  <div class="label">
                    {{ $t('common.slot') }}:
                  </div>
                  <BcLink :to="`/slot/${slotProps.data.slot}`" target="_blank" class="link" :class="getExpansionValueClass(slotProps.data)">
                    <BcFormatNumber :value="slotProps.data.slot" default="-" />
                  </BcLink>
                </div>
                <div class="row">
                  <div class="label">
                    {{ $t('dashboard.validator.col.recipient') }}:
                  </div>
                  <BcFormatHash
                    v-if="slotProps.data.recipient?.hash"
                    type="address"
                    class="recipient"
                    :class="getExpansionValueClass(slotProps.data)"
                    :hash="slotProps.data.recipient?.hash"
                    :ens="slotProps.data.recipient?.ens"
                  />
                  <span v-else>-</span>
                </div>
                <div class="row">
                  <div class="label">
                    {{ $t('dashboard.validator.col.amount') }}:
                  </div>
                  <BcFormatValue :value="slotProps.data.amount" :class="getExpansionValueClass(slotProps.data)" />
                </div>
              </div>
            </template>
            <template #empty>
              <DashboardTableAddValidator />
            </template>
          </BcTable>
        </ClientOnly>
      </template>
    </BcTableControl>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/fonts.scss";
@use "~/assets/css/utils.scss";

:deep(.withdrawal-table) {
  >.p-datatable-wrapper {
    min-height: 577px;
  }
  .index .all-time-total {
    @include fonts.standard_text;
    font-weight: var(--standard_text_medium_font_weight);
  }

  .group-id {
    @include utils.truncate-text;
  }

  .time-passed {
    white-space: nowrap;
  }

  .amount .all-time-total {
    @include fonts.standard_text;
    font-weight: var(--standard_text_medium_font_weight);
  }

  .value-with-tooltip-container {
    display: flex;
    gap: var(--padding);
  }

  tr.total-row > td {
    border-bottom-color: var(--primary-color);
  }

  .gray-out > td {
    >a,
    >div,
    >span {
      opacity: 0.5;
    }
  }
}

.expansion {
  display: flex;
  flex-direction: column;
  gap: var(--padding-small);
  padding: var(--padding) 41px;
  background: var(--table-header-background);
  font-size: var(--small_text_font_size);

  .row {
    display: flex;
    align-items: center;

    .label {
      width: 90px;
      font-weight: var(--small_text_bold_font_weight);
      flex-shrink: 0;
    }

    .gray-out {
      opacity: 0.5;
    }
  }
}
</style>
