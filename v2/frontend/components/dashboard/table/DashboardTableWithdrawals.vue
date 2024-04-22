<script setup lang="ts">
import type { DataTableSortEvent } from 'primevue/datatable'
import type { VDBBlocksTableRow } from '~/types/api/validator_dashboard'
import type { Cursor, TableQueryParams } from '~/types/datatable'
import { useValidatorDashboardOverviewStore } from '~/stores/dashboard/useValidatorDashboardOverviewStore'
import { useValidatorDashboardWithdrawalsStore } from '~/stores/dashboard/useValidatorDashboardWithdrawalsStore'
import { BcFormatHash } from '#components'
import { getGroupLabel } from '~/utils/dashboard/group'

const { dashboardKey } = useDashboardKey()

const cursor = ref<Cursor>()
const pageSize = ref<number>(5)
const { t: $t } = useI18n()

const { withdrawals, query: lastQuery, getWithdrawals } = useValidatorDashboardWithdrawalsStore()
const { value: query, bounce: setQuery } = useDebounceValue<TableQueryParams | undefined>(undefined, 500)

const { overview } = useValidatorDashboardOverviewStore()

const { width } = useWindowSize()
const colsVisible = computed(() => { // TODO: finetune
  return {
    group: width.value > 1050,
    slot: width.value > 850,
    epoch: width.value > 650,
    recipient: width.value > 450
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
}, { immediate: true })

watch(query, (q) => {
  if (q) {
    getWithdrawals(dashboardKey.value, q)
  }
}, { immediate: true })

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

const setSearch = (value?: string) => {
  loadData(setQuerySearch(value, lastQuery.value))
}

const getRowClass = (row: VDBBlocksTableRow) => {
// TODO (scheduled and total)
  if (row.status === 'scheduled') {
    return 'future-row'
  }
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
            :expandable="true"
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
                <NuxtLink
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
              field="epoch"
              :sortable="true"
              :header="$t('common.epoch')"
              body-class="epoch"
              header-class="epoch"
            >
              <template #body="slotProps">
                <NuxtLink
                  :to="`/epoch/${slotProps.data.epoch}`"
                  target="_blank"
                  class="link"
                  :no-prefetch="true"
                >
                  <BcFormatNumber :value="slotProps.data.epoch" default="-" />
                </NuxtLink>
              </template>
            </Column>
            <Column v-if="colsVisible.slot" field="slot" :sortable="true" :header="$t('common.slot')">
              <template #body="slotProps">
                <NuxtLink :to="`/slot/${slotProps.data.slot}`" target="_blank" class="link" :no-prefetch="true">
                  <BcFormatNumber :value="slotProps.data.slot" default="-" />
                </NuxtLink>
              </template>
            </Column>
            <Column field="age" body-class="age" header-class="age">
              <template #header>
                <BcTableAgeHeader />
              </template>
              <template #body="slotProps">
                <!-- TODO: Use slot here and in blocks table, requires new formating function for BcFormatTimePassed -->
                <BcFormatTimePassed class="time-passed" :value="slotProps.data.epoch" />
              </template>
            </Column>
            <Column
              v-if="colsVisible.recipient"
              field="recipient"
              header-class="recipient"
              :sortable="true"
              :header="$t('dashboard.validator.col.recipient')"
            >
              <template #body="slotProps">
                <BcFormatHash
                  v-if="slotProps.data.recipient?.hash"
                  type="address"
                  class="recipient"
                  :hash="slotProps.data.recipient?.hash"
                  :ens="slotProps.data.recipient?.ens"
                />
                <span v-else>-</span>
              </template>
            </Column>
            <Column
              field="amount"
              :sortable="true"
              body-class="amount"
              header-class="amount"
              :header="$t('dashboard.validator.col.amount')"
            >
              <template #body="slotProps">
                <!--TODO: Using BlockTableRewardItem is not perfect-->
                <BlockTableRewardItem :reward="slotProps.data.amount" />
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
                    {{ $t('common.slot') }}:
                  </div>
                  <NuxtLink :to="`/slot/${slotProps.data.slot}`" target="_blank" class="link" :no-prefetch="true">
                    <BcFormatNumber :value="slotProps.data.slot" default="-" />
                  </NuxtLink>
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

:deep(.withdrawal-table) {

  // TODO: Finetune

  .index {
    @include utils.set-all-width(110px);
  }

  .group-id {
    @include utils.set-all-width(120px);
    @include utils.truncate-text;
  }

  .epoch {
    // TODO
  }

  @media (max-width: 399px) {
    .group-id {
      @include utils.set-all-width(80px);
      @include utils.truncate-text;
    }
  }

  // TODO: Check
  .status-mobile {
    @include utils.set-all-width(40px);
    @include utils.truncate-text;
  }

  .time-passed {
    white-space: nowrap;
  }

  .reward,
  .recipient {
    .p-column-title {
      @include utils.truncate-text;
    }
  }

  .future-row {
    td {

      >div,
      >span {
        opacity: 0.5;
      }
    }
  }
}

.recipient {
  @include utils.set-all-width(120px);
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
