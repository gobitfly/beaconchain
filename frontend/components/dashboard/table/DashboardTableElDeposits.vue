<script setup lang="ts">
import type { DataTableSortEvent } from 'primevue/datatable'
import type { VDBExecutionDepositsTableRow } from '~/types/api/validator_dashboard'
import type { Cursor, TableQueryParams } from '~/types/datatable'
import { useValidatorDashboardOverviewStore } from '~/stores/dashboard/useValidatorDashboardOverviewStore'
import { getGroupLabel } from '~/utils/dashboard/group'
import { useValidatorDashboardElDepositsStore } from '~/stores/dashboard/useValidatorDashboardElDepositsStore'

const { dashboardKey } = useDashboardKey()

const cursor = ref<Cursor>()
const pageSize = ref<number>(5)
const { t: $t } = useI18n()

const { deposits, query: lastQuery, getDeposits, getTotalAmount, totalAmount, isLoadingDeposits, isLoadingTotal } = useValidatorDashboardElDepositsStore()
const { value: query, bounce: setQuery } = useDebounceValue<TableQueryParams | undefined>(undefined, 500)

const { overview } = useValidatorDashboardOverviewStore()

const { width } = useWindowSize()
const colsVisible = computed(() => {
  return {
    duty: width.value > 1180,
    clRewards: width.value >= 860,
    elRewards: width.value >= 740,
    age: width.value >= 620
  }
})

const loadData = (query?: TableQueryParams) => {
  if (!query) {
    query = { limit: pageSize.value }
  }
  setQuery(query, true, true)
}

watch(dashboardKey, (key) => {
  loadData()
  getTotalAmount(key)
}, { immediate: true })

watch(query, async (q) => {
  if (q) {
    await getDeposits(dashboardKey.value, q)
  }
}, { immediate: true })

const tableData = computed(() => {
  if (!deposits.value?.data?.length) {
    return
  }
  return {
    paging: deposits.value.paging,
    data: [
      {
        amount: totalAmount.value
      },
      ...deposits.value.data
    ]
  }
})

const groupNameLabel = (groupId?: number) => {
  return getGroupLabel($t, groupId, overview.value?.groups)
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

const getRowClass = (row: VDBExecutionDepositsTableRow) => {
  if (row.index === undefined) {
    return 'total-row'
  }
}

const isRowExpandable = (row: VDBExecutionDepositsTableRow) => {
  return row.index !== undefined
}

</script>
<template>
  <div>
    <BcTableControl :title="$t('dashboard.validator.el_deposits.title')">
      <template #table>
        <ClientOnly fallback-tag="span">
          <BcTable
            :data="tableData"
            data-key="index"
            :expandable="true"
            class="el_deposits_table"
            :cursor="cursor"
            :page-size="pageSize"
            :row-class="getRowClass"
            :add-spacer="true"
            :is-row-expandable="isRowExpandable"
            :loading="isLoadingDeposits"
            @set-cursor="setCursor"
            @sort="onSort"
            @set-page-size="setPageSize"
          >
            <Column field="public_key" header-class="public_key" :header="$t('dashboard.validator.col.public_key')">
              <template #body="slotProps">
                <BcFormatHash
                  v-if="slotProps.data.index !== undefined"
                  :hash="slotProps.data.public_key"
                  type="public_key"
                />
                <span v-else>{{ $t('table.all_time_total') }}</span>
              </template>
            </Column>
            <Column field="index" header-class="index" :header="$t('common.index')">
              <template #body="slotProps">
                <NuxtLink :to="`/validator/${slotProps.data.index}`" target="_blank" class="link" :no-prefetch="true">
                  <BcFormatNumber :value="slotProps.data.index" />
                </NuxtLink>
              </template>
            </Column>
            <Column
              field="group_id"
              body-class="group-id"
              header-class="group-id"
              :header="$t('dashboard.validator.col.group')"
            >
              <template #body="slotProps">
                <span v-if="slotProps.data.index !== undefined">
                  {{ groupNameLabel(slotProps.data.group_id) }}
                </span>
              </template>
            </Column>
            <Column field="block" body-class="block" header-class="block" :header="$t('common.block')">
              <template #body="slotProps">
                <NuxtLink :to="`/block/${slotProps.data.block}`" target="_blank" class="link" :no-prefetch="true">
                  <BcFormatNumber v-if="slotProps.data.index !== undefined" :value="slotProps.data.block" />
                </NuxtLink>
              </template>
            </Column>
            <Column v-if="colsVisible.age" field="age" body-class="age" header-class="age">
              <template #header>
                <BcTableAgeHeader />
              </template>
              <template #body="slotProps">
                <BcFormatTimePassed
                  v-if="slotProps.data.index !== undefined"
                  class="time-passed"
                  :value="slotProps.data.timestamp"
                  type="go-timestamp"
                />
              </template>
            </Column>
            <Column field="from" header-class="from" :header="$t('table.from')">
              <template #body="slotProps">
                <BcFormatHash
                  v-if="slotProps.data.index !== undefined"
                  :hash="slotProps.data.from.hash"
                  :ens="slotProps.data.from.ens"
                  type="address"
                />
              </template>
            </Column>
            <Column field="depositor" header-class="depositor" :header="$t('dashboard.validator.col.depositor')">
              <template #body="slotProps">
                <BcFormatHash
                  v-if="slotProps.data.index !== undefined"
                  :hash="slotProps.data.depositor.hash"
                  :ens="slotProps.data.depositor.ens"
                  type="address"
                />
              </template>
            </Column>
            <Column field="tx_hash" header-class="tx_hash" :header="$t('block.col.tx_hash')">
              <template #body="slotProps">
                <BcFormatHash v-if="slotProps.data.index !== undefined" :hash="slotProps.data.tx_hash" type="tx" />
              </template>
            </Column>
            <Column
              field="withdrawal_credentials"
              header-class="withdrawal_credentials"
              :header="$t('dashboard.validator.col.withdrawal_credentials')"
            >
              <template #body="slotProps">
                <BcFormatHash
                  v-if="slotProps.data.index !== undefined"
                  :hash="slotProps.data.withdrawal_credentials"
                  type="withdrawal_credentials"
                />
              </template>
            </Column>
            <Column field="amount" body-class="amount" header-class="amount" :header="$t('table.amount')">
              <template #body="slotProps">
                <div v-if="slotProps.data.index === undefined && isLoadingTotal">
                  <BcLoadingSpinner :loading="true" size="small" />
                </div>
                <BcFormatValue v-else :value="slotProps.data.amount" :options="{ fixedDecimalCount: 0 }" />
              </template>
            </Column>
            <Column field="valid" header-class="valid" body-class="valid" :header="$t('table.valid')">
              <template #body="slotProps">
                <BcTableValidTag v-if="slotProps.data.index !== undefined" :valid="slotProps.data.valid" />
              </template>
            </Column>
            <template #expansion="slotProps">
              <div class="expansion">
                <div class="row">
                  <div class="label">
                    {{ $t('dashboard.validator.col.public_key') }}
                  </div>
                  <BcFormatHash
                    :hash="slotProps.data.public_key"
                    type="public_key"
                  />
                </div>
                <div class="row">
                  <div class="label">
                    {{ $t('dashboard.validator.col.group') }}
                  </div>
                  <div>
                    {{ groupNameLabel(slotProps.data.group_id) }}
                  </div>
                </div>
                <div class="row">
                  <div class="label">
                    {{ $t('common.block') }}
                  </div>
                  <NuxtLink :to="`/block/${slotProps.data.block}`" target="_blank" class="link" :no-prefetch="true">
                    <BcFormatNumber :value="slotProps.data.block" />
                  </NuxtLink>
                </div>
                <div class="row">
                  <div class="label">
                    {{ $t('table.from') }}
                  </div>
                  <BcFormatHash
                    :hash="slotProps.data.from.hash"
                    :ens="slotProps.data.from.ens"
                    type="address"
                  />
                </div>
                <div class="row">
                  <div class="label">
                    {{ $t('dashboard.validator.col.depositor') }}
                  </div>
                  <BcFormatHash
                    :hash="slotProps.data.depositor.hash"
                    :ens="slotProps.data.depositor.ens"
                    type="address"
                  />
                </div>
                <div class="row">
                  <div class="label">
                    {{ $t('block.col.tx_hash') }}
                  </div><BcFormatHash :hash="slotProps.data.tx_hash" type="tx" />
                </div>
                <div class="row">
                  <div class="label">
                    {{ $t('dashboard.validator.col.withdrawal_credentials') }}
                  </div>
                  <BcFormatHash
                    :hash="slotProps.data.withdrawal_credentials"
                    type="withdrawal_credentials"
                  />
                </div>
                <div class="row">
                  <div class="label">
                    {{ $t('table.valid') }}
                  </div>
                  <div>
                    <BcTableValidTag :valid="slotProps.data.valid" />
                  </div>
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
@use "~/assets/css/utils.scss";

:deep(.el_deposits_table) {
  --col-width: 154px;

  .epoch {
    @include utils.set-all-width(80px);
  }

  .group_id,
  .reward {
    @include utils.set-all-width(120px);
  }

  .time-passed {
    white-space: nowrap;
  }

  tr:not(.p-datatable-row-expansion) {
    @media (max-width: 1300px) {
      .duty {
        @include utils.set-all-width(300px);
      }
    }
  }

  .total-row {
    td {
      font-weight: var(--standard_text_medium_font_weight);
      border-bottom-color: var(--primary-color);
    }
  }
}

.expansion {
  color: var(--container-color);
  background-color: var(--container-background);
  display: flex;
  flex-direction: column;
  gap: var(--padding);
  padding: var(--padding);

  .row {
    display: flex;
    gap: var(--padding);

    .label {
      width: 164px;
      font-weight: var(--standard_text_bold_font_weight);
    }
  }
}
</style>
