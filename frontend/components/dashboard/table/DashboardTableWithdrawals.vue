<script setup lang="ts">
import type { DataTableSortEvent } from 'primevue/datatable'
import type { VDBWithdrawalsTableRow } from '~/types/api/validator_dashboard'
import type { Cursor, TableQueryParams } from '~/types/datatable'
import { useValidatorDashboardOverviewStore } from '~/stores/dashboard/useValidatorDashboardOverviewStore'
import { useValidatorDashboardWithdrawalsStore } from '~/stores/dashboard/useValidatorDashboardWithdrawalsStore'
import { BcFormatHash } from '#components'
import { getGroupLabel } from '~/utils/dashboard/group'
import { DAHSHBOARDS_ALL_GROUPS_ID } from '~/types/dashboard/index'

const { dashboardKey } = useDashboardKey()

const cursor = ref<Cursor>()
const pageSize = ref<number>(5)
const { t: $t } = useI18n()

const { withdrawals, query: lastQuery, getWithdrawals, totalWithdrawals, getTotalWithdrawals } = useValidatorDashboardWithdrawalsStore()
const { value: query, bounce: setQuery } = useDebounceValue<TableQueryParams | undefined>(undefined, 500)

const { overview } = useValidatorDashboardOverviewStore()

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
    query = { limit: pageSize.value }
  }
  setQuery(query, true, true)
}

watch(dashboardKey, () => {
  loadData()
  getTotalWithdrawals(dashboardKey.value)
}, { immediate: true })

watch(query, (q) => {
  if (q) {
    getWithdrawals(dashboardKey.value, q)
  }
}, { immediate: true })

watch(withdrawals, () => {
  // keep total withdrawals sticky at the top
  if (withdrawals.value?.data && totalWithdrawals.value !== undefined) {
    if (withdrawals.value.data[0].group_id === DAHSHBOARDS_ALL_GROUPS_ID) {
      return
    }

    withdrawals.value.data.unshift({
      epoch: 0,
      slot: 0,
      group_id: DAHSHBOARDS_ALL_GROUPS_ID,
      recipient: { hash: '' },
      amount: totalWithdrawals.value.data.total_amount,
      index: 0
    })
  }
})

const groupNameLabel = (groupId?: number) => {
  return getGroupLabel($t, groupId, overview.value?.groups, '')
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

const getRowClass = (row: VDBWithdrawalsTableRow) => {
  if (row.group_id === DAHSHBOARDS_ALL_GROUPS_ID) {
    return 'total-row'
  }
  // TODO: Future withdrawals
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
            :data="withdrawals"
            data-key="epoch"
            :expandable="!colsVisible.group"
            class="withdrawal-table"
            :cursor="cursor"
            :page-size="pageSize"
            :row-class="getRowClass"
            :add-spacer="true"
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
                <div v-if="slotProps.data.group_id === DAHSHBOARDS_ALL_GROUPS_ID" class="all-time-total">
                  {{ $t('dashboard.validator.withdrawals.all_time_total') }}
                </div>
                <NuxtLink
                  v-else
                  :to="`/validator/${slotProps.data.index}`"
                  target="_blank"
                  class="link"
                  :no-prefetch="true"
                >
                  <BcFormatNumber :value="slotProps.data.index" default="-" />
                </NuxtLink>
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
                <span>
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
                <NuxtLink
                  v-if="slotProps.data.group_id !== DAHSHBOARDS_ALL_GROUPS_ID"
                  :to="`/epoch/${slotProps.data.epoch}`"
                  target="_blank"
                  class="link"
                  :no-prefetch="true"
                >
                  <BcFormatNumber :value="slotProps.data.epoch" default="-" />
                </NuxtLink>
              </template>
            </Column>
            <Column
              v-if="colsVisible.slot"
              field="slot"
              :sortable="true"
              :header="$t('common.slot')"
            >
              <template #body="slotProps">
                <NuxtLink
                  v-if="slotProps.data.group_id !== DAHSHBOARDS_ALL_GROUPS_ID"
                  :to="`/slot/${slotProps.data.slot}`"
                  target="_blank"
                  class="link"
                  :no-prefetch="true"
                >
                  <BcFormatNumber :value="slotProps.data.slot" default="-" />
                </NuxtLink>
              </template>
            </Column>
            <Column
              field="age"
              body-class="age"
              header-class="age"
            >
              <template #header>
                <BcTableAgeHeader />
              </template>
              <template #body="slotProps">
                <BcFormatTimePassed
                  v-if="slotProps.data.group_id !== DAHSHBOARDS_ALL_GROUPS_ID"
                  type="slot"
                  class="time-passed"
                  :value="slotProps.data.epoch"
                />
              </template>
            </Column>
            <Column
              v-if="colsVisible.recipient"
              field="recipient"
              header-class="recipient"
              body-class="recipient"
              :sortable="true"
              :header="$t('dashboard.validator.col.recipient')"
            >
              <template #body="slotProps">
                <div v-if="slotProps.data.group_id !== DAHSHBOARDS_ALL_GROUPS_ID">
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
                <BcFormatValue
                  :value="slotProps.data.amount"
                  :class="{'all-time-total':slotProps.data.group_id === DAHSHBOARDS_ALL_GROUPS_ID}"
                />
              </template>
            </Column>
            <template #expansion="slotProps">
              <div class="expansion">
                <div v-if="!colsVisible.group" class="row">
                  <div class="label">
                    {{ $t('dashboard.validator.col.group') }}:
                  </div>
                  {{ groupNameLabel(slotProps.data.group_id) }}
                </div>
                <div v-if="!colsVisible.group" class="row">
                  <div class="label">
                    {{ $t('common.epoch') }}:
                  </div>
                  <NuxtLink :to="`/epoch/${slotProps.data.epoch}`" target="_blank" class="link" :no-prefetch="true">
                    <BcFormatNumber :value="slotProps.data.epoch" default="-" />
                  </NuxtLink>
                </div>
                <div v-if="!colsVisible.group" class="row">
                  <div class="label">
                    {{ $t('common.slot') }}:
                  </div>
                  <NuxtLink :to="`/slot/${slotProps.data.slot}`" target="_blank" class="link" :no-prefetch="true">
                    <BcFormatNumber :value="slotProps.data.slot" default="-" />
                  </NuxtLink>
                </div>
                <div v-if="!colsVisible.group" class="row">
                  <div class="label">
                    {{ $t('dashboard.validator.col.recipient') }}:
                  </div>
                  <BcFormatHash
                    v-if="slotProps.data.recipient?.hash"
                    type="address"
                    class="recipient"
                    :hash="slotProps.data.recipient?.hash"
                    :ens="slotProps.data.recipient?.ens"
                  />
                  <span v-else>-</span>
                </div>
                <div v-if="!colsVisible.group" class="row">
                  <div class="label">
                    {{ $t('dashboard.validator.col.amount') }}:
                  </div>
                  <BcFormatValue :value="slotProps.data.amount" />
                </div>
              </div>
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
  .index {
    @include utils.set-all-width(140px);

    .all-time-total {
      @include fonts.standard_text;
      font-weight: var(--standard_text_medium_font_weight);
    }
  }

  .group-id {
    @include utils.set-all-width(160px);
    @include utils.truncate-text;
  }

  .age {
    @include utils.set-all-width(195px);
  }

  .recipient {
    @include utils.set-all-width(180px);
  }

  .time-passed {
    white-space: nowrap;
  }

  .amount {
    .all-time-total {
      @include fonts.standard_text;
      font-weight: var(--standard_text_medium_font_weight);
    }
  }

  tr.total-row > td {
    border-bottom-color: var(--primary-color);
  }

  // TODO: Tooltip in future row for amount
  .future-row > td {
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

    .value {
      text-wrap: wrap;
      word-break: break-all;
    }
  }
}
</style>
