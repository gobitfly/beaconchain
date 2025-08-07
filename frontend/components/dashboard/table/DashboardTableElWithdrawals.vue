<script setup lang="ts">
import type { DataTableSortEvent } from 'primevue/datatable'
import type {
  GetValidatorDashboardExecutionLayerWithdrawalsResponse,
  GetValidatorDashboardTotalExecutionWithdrawalsResponse,
  VDBWithdrawalsElTableRow,
} from '~/types/api/validator_dashboard'
import type {
  Cursor, TableQueryParams,
} from '~/types/datatable'

const {
  elWithdrawals,
  elWithdrawalsTotalAmount,
} = defineProps<{
  elWithdrawals?: GetValidatorDashboardExecutionLayerWithdrawalsResponse,
  elWithdrawalsTotalAmount?: GetValidatorDashboardTotalExecutionWithdrawalsResponse,
  isLoading: boolean,
}>()

const {
  isGuestDashboard,
} = useDashboardKey()
const { t: $t } = useTranslation()

const validatorDashboardOverviewStore = useValidatorDashboardOverviewStore()
const {
  hasValidators,
} = storeToRefs(validatorDashboardOverviewStore)

const { width } = useWindowSize()
const isMobile = computed(() => {
  return width.value < 768
})

const elWithdrawalsWithIdentifiers = computed(() =>
  addIdentifier(elWithdrawals, 'block_queued', 'tx_index_queued', 'itx_index_queued'),
)
const tableData = computed(() => {
  if (!elWithdrawalsWithIdentifiers.value?.data?.length) {
    return
  }

  return {
    data: [
      {
        amount: elWithdrawalsTotalAmount?.data.total_amount,
        isTotalAmountRow: true,
      },
      ...elWithdrawalsWithIdentifiers.value.data,
    ],
    paging: elWithdrawalsWithIdentifiers.value.paging,
  }
})

const query = defineModel<TableQueryParams>('query')

const onSort = (sort: DataTableSortEvent) => {
  query.value = setQuerySort(sort, query.value)
}
const setCursor = (cursor: Cursor) => {
  query.value = setQueryCursor(cursor, query.value)
}
const setPageSize = (limit: number) => {
  query.value = setQueryPageSize(limit, query.value)
}
const setSearch = (value?: string) => {
  query.value = {
    ...query.value,
    search: value,
  }
}

const { groups } = useValidatorDashboardGroups()
const getGroupName = (groupId: number) => {
  return groups.value.find(group => group.id === groupId)?.name
}
const v1Domain = useV1Domain()
const emit = defineEmits<{
  (e: 'add-validator'): void,
}>()
</script>

<template>
  <BcTableControl
    :title="$t('dashboard.validator.el_withdrawals.title')"
    :search-placeholder="
      $t(
        isGuestDashboard
          ? 'dashboard.validator.el_withdrawals.search_placeholder_guest_dashboard'
          : 'dashboard.validator.el_withdrawals.search_placeholder_private_dashboard',
      )
    "
    @set-search="setSearch"
  >
    <template #table>
      <ClientOnly fallback-tag="span">
        <BcTable
          :data="tableData"
          expandable
          table-class="dashboard-table-el-withdrawals"
          data-key="identifier"
          :selected-sort="query?.sort"
          :cursor="query?.cursor"
          :page-size="query?.limit"
          :row-class="(row: VDBWithdrawalsElTableRow) => row.status === 'queued' ? 'grayed-out-row' : ''"
          :is-row-expandable="(row: VDBWithdrawalsElTableRow) => row.index !== undefined"
          @set-cursor="setCursor"
          @sort="onSort"
          @set-page-size="setPageSize"
        >
          <Column
            sortable
            body-class="dashboard-table-el-withdrawals__age-cell"
            field="timestamp"
          >
            <template #header>
              <BcTableAgeHeader />
            </template>
            <template #body="slotProps">
              <span v-if="slotProps.data.isTotalAmountRow">Σ</span>
              <BcTableDateTime
                v-else-if="slotProps.data.timestamp_queued"
                :unix-timestamp="slotProps.data.timestamp_queued"
              />
              <span v-else>-</span>
            </template>
          </Column>
          <Column
            v-if="!isMobile"
            field="validator_index"
            :header="$t('dashboard.validator.col.validator_index')"
          >
            <template #body="slotProps">
              <BcIcon
                v-if="!slotProps.data.isTotalAmountRow"
                name="desktop"
                size="sm"
                class="dashboard-table-el-withdrawals__desktop-icon"
              />
              <BcLink
                v-if="!slotProps.data.isTotalAmountRow"
                :to="`${v1Domain}/validator/${slotProps.data.index}`"
                target="_blank"
                class="link"
              >
                {{ slotProps.data.index }}
              </BcLink>
            </template>
          </Column>
          <Column
            v-if="!isMobile"
            field="group_id"
            :header="$t('dashboard.validator.col.group')"
          >
            <template #body="slotProps">
              <span v-if="!slotProps.data.isTotalAmountRow">
                {{ getGroupName(slotProps.data.group_id) }}
              </span>
            </template>
          </Column>
          <Column
            v-if="!isMobile"
            field="consolidator"
            :header="$t('dashboard.validator.col.withdrawer')"
          >
            <template #body="slotProps">
              <BcFormatHash
                v-if="!slotProps.data.isTotalAmountRow"
                :hash="slotProps.data.withdrawer.hash"
                type="address"
                :no-wrap="true"
                :ens="slotProps.data.withdrawer.ens"
              />
            </template>
          </Column>
          <Column
            v-if="!isMobile"
            field="status"
            :header="$t('table.status')"
          >
            <template #body="slotProps">
              <DashboardTableElWithdrawalsStatus
                v-if="!slotProps.data.isTotalAmountRow"
                :status="slotProps.data.status"
              />
            </template>
          </Column>
          <Column
            field="amount"
            sortable
            :header="$t('table.amount')"
          >
            <template #body="slotProps">
              <BcTooltip
                v-if="slotProps.data.status === 'queued'"
                tooltip-width="216px"
                tooltip-text-align="left"
              >
                <BcFormatAmount
                  :value="slotProps.data.amount"
                />
                <template #tooltip>
                  <slot name="tooltip">
                    {{ $t('dashboard.validator.el_withdrawals.info_amount_estimated') }}
                  </slot>
                </template>
              </BcTooltip>
              <BcFormatAmount
                v-else
                :value="slotProps.data.amount"
                has-tooltip
              />
            </template>
          </Column>
          <Column
            v-if="isMobile"
            field="status"
          >
            <template #body="slotProps">
              <DashboardTableElWithdrawalsStatus
                v-if="!slotProps.data.isTotalAmountRow"
                :status="slotProps.data.status"
                is-compact
              />
            </template>
          </Column>
          <template #expansion="slotProps">
            <div class="dashboard-table-el-withdrawals__details">
              <div class="dashboard-table-el-withdrawals__details-grid">
                <div
                  v-if="isMobile"
                  class="dashboard-table-el-withdrawals__details-row"
                >
                  <div class="dashboard-table-el-withdrawals__details-label">
                    {{ $t("dashboard.validator.col.validator_index") }}
                  </div>
                  <div class="dashboard-table-el-withdrawals__details-value">
                    <BcIcon
                      name="desktop"
                      size="sm"
                      class="dashboard-table-el-withdrawals__desktop-icon"
                    />
                    <BcLink
                      :to="`${v1Domain}/validator/${slotProps.data.index}`"
                      target="_blank"
                      class="link"
                    >
                      {{ slotProps.data.index }}
                    </BcLink>
                  </div>
                </div>
                <div
                  v-if="isMobile"
                  class="dashboard-table-el-withdrawals__details-row"
                >
                  <div class="dashboard-table-el-withdrawals__details-label">
                    {{ $t("dashboard.validator.col.group") }}
                  </div>
                  <div class="value">
                    {{ getGroupName(slotProps.data.group_id) }}
                  </div>
                </div>
                <div class="dashboard-table-el-withdrawals__details-row">
                  <div class="dashboard-table-el-withdrawals__details-label">
                    {{ $t("block.col.transaction_hash") }}
                  </div>
                  <BcFormatHash
                    :hash="slotProps.data.tx_hash"
                    :no-wrap="true"
                    type="tx"
                  />
                </div>
                <div class="dashboard-table-el-withdrawals__details-row">
                  <div class="dashboard-table-el-withdrawals__details-label">
                    {{ $t("dashboard.validator.col.block_queued") }}
                  </div>
                  <BcLink
                    v-if="slotProps.data.block_queued !== undefined"
                    :to="`${v1Domain}/block/${slotProps.data.block_queued}`"
                    target="_blank"
                    class="link"
                  >
                    <BcFormatNumber :value="slotProps.data.block_queued" />
                  </BcLink>
                  <span v-else>-</span>
                </div>
                <div
                  class="dashboard-table-el-withdrawals__details-row"
                >
                  <div class="dashboard-table-el-withdrawals__details-label">
                    {{ $t("dashboard.validator.col.block_processed") }}
                  </div>
                  <div class="dashboard-table-el-withdrawals__details-value">
                    <BcLink
                      v-if="slotProps.data.block_processed !== undefined"
                      :to="`${v1Domain}/block/${slotProps.data.block_processed}`"
                      target="_blank"
                      class="link"
                    >
                      <BcFormatNumber :value="slotProps.data.block_processed" />
                    </BcLink>
                    <span v-else>-</span>

                    <BcTooltip
                      v-if="slotProps.data.status === 'queued'"
                      tooltip-width="200px"
                      tooltip-text-align="left"
                      class="dashboard-table-el-withdrawals__detail-value-tooltip"
                    >
                      <BcIcon name="circle-info" />
                      <template #tooltip>
                        <slot name="tooltip">
                          {{ $t('dashboard.validator.el_withdrawals.info_block_processed_estimated') }}
                        </slot>
                      </template>
                    </BcTooltip>
                  </div>
                </div>
                <div
                  v-if="isMobile"
                  class="dashboard-table-el-withdrawals__details-row"
                >
                  <div class="dashboard-table-el-withdrawals__details-label">
                    {{ $t("dashboard.validator.col.withdrawer") }}
                  </div>
                  <div class="dashboard-table-el-withdrawals__details-value">
                    <BcFormatHash
                      v-if="!slotProps.data.isTotalAmountRow"
                      :hash="slotProps.data.withdrawer.hash"
                      type="address"
                      :no-wrap="true"
                      :ens="slotProps.data.withdrawer.ens"
                    />
                  </div>
                </div>
                <div
                  v-if="isMobile"
                  class="dashboard-table-el-withdrawals__details-row"
                >
                  <div class="dashboard-table-el-withdrawals__details-label">
                    {{ $t("table.status") }}
                  </div>
                  <div class="dashboard-table-el-withdrawals__details-value">
                    <DashboardTableElConsolidationsStatus
                      :status="slotProps.data.status"
                    />
                  </div>
                </div>
                <div
                  class="dashboard-table-el-withdrawals__details-row"
                >
                  <div class="dashboard-table-el-withdrawals__details-label">
                    {{ $t("table.withdrawal_request_fee") }}
                  </div>
                  <div class="dashboard-table-el-withdrawals__details-value">
                    <BcFormatAmount
                      :value="slotProps.data.fee"
                      has-tooltip
                    />
                  </div>
                </div>
              </div>
            </div>
          </template>
          <template #empty>
            <DashboardTableAddValidator
              v-if="!hasValidators"
              @add-validator="emit('add-validator')"
            />
          </template>
        </BcTable>
      </ClientOnly>
    </template>
  </BcTableControl>
</template>

<style lang="scss" scoped>
@use '~/assets/css/breakpoints' as *;

.dashboard-table-el-withdrawals__desktop-icon {
  margin-right: var(--padding);
  color: var(--text-color-discreet)
}

.dashboard-table-el-withdrawals__details {
  padding: var(--padding-medium);
  color: var(--container-color);
  font-size: var(--small_text_font_size);
  background-color: var(--container-background);
}

.dashboard-table-el-withdrawals__details-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, max-content));
  gap: var(--padding-medium) calc(var(--padding-large) * 2);

  @media (min-width: $breakpoint-md) {
    grid-template-columns: repeat(4, minmax(0, max-content));
    grid-template-rows: repeat(3, auto);
    grid-auto-flow: column;
  }
}

.dashboard-table-el-withdrawals__details-row {
  display: grid;
  grid-template-columns: subgrid;
  grid-column: span 2;
  column-gap: var(--padding-large);
  align-items: center;
}

.dashboard-table-el-withdrawals__details-label {
  font-weight: var(--standard_text_bold_font_weight);
}

.dashboard-table-el-withdrawals__details-value {
  display: flex;
}

.dashboard-table-el-withdrawals__desktop-icon {
  margin-right: var(--padding);
  color: var(--text-color-discreet);
}

.dashboard-table-el-withdrawals__detail-value-tooltip {
  margin-left: var(--padding);
}

:deep(.dashboard-table-el-withdrawals) {
  .grayed-out-row > td > *  {
    opacity: 0.5;
  }

  .dashboard-table-el-withdrawals__age-cell {
    padding-top: 0 !important;
    padding-bottom: 0 !important;
  }
}
</style>
