<script setup lang="ts">
import type { DataTableSortEvent } from 'primevue/datatable'
import type {
  GetValidatorDashboardConsensusLayerDepositsResponse,
  GetValidatorDashboardTotalConsensusDepositsResponse,
  VDBConsensusDepositsTableRow,
} from '~/types/api/validator_dashboard'
import type {
  Cursor, TableQueryParams,
} from '~/types/datatable'
import { useValidatorDashboardOverviewStore } from '~/stores/dashboard/useValidatorDashboardOverviewStore'
import { getGroupLabel } from '~/utils/dashboard/group'
import { useNetworkStore } from '~/stores/useNetworkStore'

const {
  clDeposits,
  clDepositsTotalAmount,
} = defineProps<{
  clDeposits?: GetValidatorDashboardConsensusLayerDepositsResponse,
  clDepositsTotalAmount?: GetValidatorDashboardTotalConsensusDepositsResponse,
  isLoading: boolean,
}>()

const {
  isGuestDashboard,
} = useDashboardKey()

const { t: $t } = useTranslation()

const {
  getTimestampFromSlot,
} = useNetworkStore()

const validatorDashboardOverviewStore = useValidatorDashboardOverviewStore()
const {
  hasValidators,
} = storeToRefs(validatorDashboardOverviewStore)
const { groups } = useValidatorDashboardGroups()

const query = defineModel<TableQueryParams>('query')

const { width } = useWindowSize()
const isMobile = computed(() => {
  return width.value < 768
})

const tableData = computed(() => {
  if (!clDeposits?.data?.length) {
    return
  }

  return {
    data: [
      {
        amount: clDepositsTotalAmount?.data.total_amount,
        isTotalAmountRow: true,
      },
      ...clDeposits.data,
    ],
    paging: clDeposits.paging,
  }
})

const groupNameLabel = (groupId?: number) => {
  return getGroupLabel($t, groupId, groups.value)
}

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

const getRowClass = (row: VDBConsensusDepositsTableRow) => {
  if (row.index === undefined) {
    return 'total-row'
  }
  if (row.status === 'queued') {
    return 'grayed-out-row'
  }
}

const isRowExpandable = (row: VDBConsensusDepositsTableRow) => {
  return row.index !== undefined
}

const {
  displayCurrencyDefault,
  selectedCurrencyMain,
} = useCurrency()
</script>

<template>
  <BcTableControl
    :title="$t('dashboard.validator.cl_deposits.title')"
    :search-placeholder="
      $t(
        isGuestDashboard
          ? 'dashboard.validator.cl_deposits.search_placeholder_guest_dashboard'
          : 'dashboard.validator.cl_deposits.search_placeholder_private_dashboard',
      )
    "
    @set-search="setSearch"
  >
    <template #table>
      <ClientOnly fallback-tag="span">
        <BcTable
          :data="tableData"
          data-key="identifier"
          expandable
          table-class="dashboard-table-cl-deposits"
          :cursor="query?.cursor"
          :page-size="query?.limit"
          :row-class="getRowClass"
          :is-row-expandable
          :is-loading
          @set-cursor="setCursor"
          @sort="onSort"
          @set-page-size="setPageSize"
        >
          <Column
            field="timestamp"
            body-class="age-field"
            sortable
          >
            <template #header>
              <BcTableAgeHeader />
            </template>
            <template #body="slotProps">
              <span v-if="slotProps.data.isTotalAmountRow">Σ</span>
              <BcTableDateTime
                v-else-if="slotProps.data.slot_processed !== undefined"
                :unix-timestamp="getTimestampFromSlot(slotProps.data.slot_processed)"
              />
              <span v-else>-</span>
            </template>
          </Column>
          <Column
            v-if="!isMobile"
            :header="$t('dashboard.validator.col.validator_index')"
          >
            <template #body="slotProps">
              <BcIcon
                v-if="!slotProps.data.isTotalAmountRow"
                name="desktop"
                size="sm"
                class="dashboard-table-cl-deposits__desktop-icon"
              />
              <BcLink
                v-if="!slotProps.data.isTotalAmountRow"
                :to="`/validator/${slotProps.data.index}`"
                target="_blank"
                class="link"
              >
                {{ slotProps.data.index }}
              </BcLink>
            </template>
          </Column>
          <Column
            v-if="!isMobile"
            body-class="group-id"
            header-class="group-id"
            :header="$t('dashboard.validator.col.group')"
          >
            <template #body="slotProps">
              <span v-if="!slotProps.data.isTotalAmountRow">
                {{ groupNameLabel(slotProps.data.group_id) }}
              </span>
            </template>
          </Column>
          <Column
            v-if="!isMobile"
            :header="$t('dashboard.validator.col.type')"
          >
            <template #body="slotProps">
              <BcBadge
                v-if="!slotProps.data.isTotalAmountRow"
                class="dashboard-table-cl-deposits__type-badge"
                color="gray"
              >
                {{
                  slotProps.data.type === 'manual'
                    ? $t('dashboard.validator.cl_deposits.type.manual')
                    : $t('dashboard.validator.cl_deposits.type.auto')
                }}
              </BcBadge>
            </template>
          </Column>
          <Column
            v-if="!isMobile"
            :header="$t('table.status')"
          >
            <template #body="slotProps">
              <div
                v-if="!slotProps.data.isTotalAmountRow"
                class="status-cell-content"
              >
                <DashboardTableClDepositsStatus :status="slotProps.data.status" />
              </div>
            </template>
          </Column>
          <Column
            field="amount"
            :header="$t('table.amount')"
            sortable
          >
            <template #body="slotProps">
              <BcTooltip
                fit-content
              >
                <BcFormatAmount
                  :value="slotProps.data.amount"
                  target-currency="clDisplayCurrency"
                  has-tooltip
                />
                <template
                  v-if="displayCurrencyDefault.executionLayer !== selectedCurrencyMain"
                  #tooltip
                >
                  <BcFormatAmount
                    :value="slotProps.data.amount"
                    has-higher-precision
                  />
                </template>
              </BcTooltip>
            </template>
          </Column>
          <Column
            v-if="isMobile"
          >
            <template #body="slotProps">
              <div
                v-if="!slotProps.data.isTotalAmountRow"
                class="status-cell-content"
              >
                <DashboardTableClDepositsStatus
                  :status="slotProps.data.status"
                  is-mobile
                />
              </div>
            </template>
          </Column>
          <template #expansion="slotProps">
            <div class="dashboard-table-cl-deposits__details">
              <div class="dashboard-table-cl-deposits__details-grid">
                <div
                  v-if="isMobile"
                  class="dashboard-table-cl-deposits__details-row"
                >
                  <div class="dashboard-table-cl-deposits__details-label">
                    {{ $t("dashboard.validator.col.validator_index") }}
                  </div>
                  <div class="dashboard-table-cl-deposits__details-value">
                    <BcIcon
                      v-if="!slotProps.data.isTotalAmountRow"
                      name="desktop"
                      size="sm"
                      class="dashboard-table-cl-deposits__desktop-icon"
                    />
                    <BcLink
                      :to="`/validator/${slotProps.data.index}`"
                      target="_blank"
                      class="link"
                    >
                      {{ slotProps.data.index }}
                    </BcLink>
                  </div>
                </div>
                <div
                  v-if="isMobile"
                  class="dashboard-table-cl-deposits__details-row"
                >
                  <div class="dashboard-table-cl-deposits__details-label">
                    {{ $t("dashboard.validator.col.group") }}
                  </div>
                  <div class="dashboard-table-cl-deposits__details-value">
                    {{ groupNameLabel(slotProps.data.group_id) }}
                  </div>
                </div>
                <div class="dashboard-table-cl-deposits__details-row">
                  <div class="dashboard-table-cl-deposits__details-label">
                    {{ $t("dashboard.validator.col.slot_queued") }}
                  </div>
                  <BcLink
                    v-if="slotProps.data.slot_queued !== undefined"
                    :to="`/slot/${slotProps.data.slot_queued}`"
                    target="_blank"
                    class="link"
                  >
                    <BcFormatNumber :value="slotProps.data.slot_queued" />
                  </BcLink>
                  <span v-else>-</span>
                </div>
                <div class="dashboard-table-cl-deposits__details-row">
                  <div class="dashboard-table-cl-deposits__details-label">
                    {{ $t("dashboard.validator.col.slot_processed") }}
                  </div>
                  <div class="dashboard-table-cl-deposits__details-value">
                    <BcLink
                      v-if="slotProps.data.slot_processed !== undefined"
                      :to="`/slot/${slotProps.data.slot_processed}`"
                      target="_blank"
                      class="link"
                    >
                      <BcFormatNumber :value="slotProps.data.slot_processed" />
                    </BcLink>
                    <span v-else>-</span>

                    <BcTooltip
                      v-if="slotProps.data.status === 'queued'"
                      tooltip-width="175px"
                      tooltip-text-align="left"
                      class="dashboard-table-cl-deposits__detail-value-tooltip"
                    >
                      <BcIcon name="circle-info" />
                      <template #tooltip>
                        <slot name="tooltip">
                          {{ $t('dashboard.validator.cl_deposits.info_slot_processed_estimated') }}
                        </slot>
                      </template>
                    </BcTooltip>
                  </div>
                </div>
                <div
                  v-if="isMobile"
                  class="dashboard-table-cl-deposits__details-row"
                >
                  <div class="dashboard-table-cl-deposits__details-label">
                    {{ $t("dashboard.validator.col.type") }}
                  </div>
                  <BcBadge
                    v-if="slotProps.data.index !== undefined"
                    class="dashboard-table-cl-deposits__type-badge"
                    color="gray"
                  >
                    {{
                      slotProps.data.type === 'manual'
                        ? $t('dashboard.validator.cl_deposits.type.manual')
                        : $t('dashboard.validator.cl_deposits.type.auto')
                    }}
                  </BcBadge>
                </div>
                <div
                  v-if="isMobile"
                  class="dashboard-table-cl-deposits__details-row"
                >
                  <div class="dashboard-table-cl-deposits__details-label">
                    {{ $t("dashboard.validator.col.status") }}
                  </div>
                  <DashboardTableClDepositsStatus :status="slotProps.data.status" />
                </div>
                <div class="dashboard-table-cl-deposits__details-row">
                  <div class="dashboard-table-cl-deposits__details-label">
                    {{ $t("dashboard.validator.col.public_key") }}
                  </div>
                  <BcFormatHash
                    :hash="slotProps.data.public_key"
                    type="public_key"
                    :no-wrap="true"
                  />
                </div>
                <div class="dashboard-table-cl-deposits__details-row">
                  <div class="dashboard-table-cl-deposits__details-label">
                    {{ $t("dashboard.validator.col.withdrawal_credential") }}
                  </div>
                  <BcFormatHash
                    :hash="slotProps.data.withdrawal_credential"
                    type="withdrawal_credentials"
                    :no-wrap="true"
                  />
                </div>
                <div class="dashboard-table-cl-deposits__details-row">
                  <div class="dashboard-table-cl-deposits__details-label">
                    {{ $t("dashboard.validator.col.signature") }}
                  </div>
                  <BcFormatHash
                    v-if="slotProps.data.index !== undefined"
                    :hash="slotProps.data.signature"
                    :no-wrap="true"
                  />
                </div>
              </div>
              <span
                v-if="slotProps.data.status === 'queued'"
                class="dashboard-table-cl-deposits__details-footer-text"
              >
                <span>
                  {{ $t('dashboard.validator.cl_deposits.info_compound_event_deposit') }}
                </span>
              </span>
            </div>
          </template>
          <template #empty>
            <DashboardTableAddValidator v-if="!hasValidators" />
          </template>
        </BcTable>
      </ClientOnly>
    </template>
  </BcTableControl>
</template>

<style lang="scss" scoped>
@use "~/assets/css/utils.scss";
@use '~/assets/css/breakpoints' as *;

:deep(.dashboard-table-cl-deposits) {
  > .p-datatable-wrapper {
    min-height: 335px;
  }

  .withdrawal-credentials {
    @include utils.truncate-text;
  }

  .group-id {
    @include utils.set-all-width(120px);
    @include utils.truncate-text;
  }

  .total-row {
    td {
      font-weight: var(--standard_text_medium_font_weight);
      border-bottom-color: var(--primary-color);
    }
  }

  .grayed-out-row > td > * {
    opacity: 0.5;
  }

  .age-field {
    white-space: nowrap;
  }
  tr > td.age-field {
    padding: 0 7px;
    @include utils.set-all-width(151px);
  }
}

.dashboard-table-cl-deposits {
  &__details {
    padding: var(--padding-medium);
    color: var(--container-color);
    font-size: var(--small_text_font_size);
    background-color: var(--container-background);
  }

  &__details-grid {
    display: grid;
    grid-template-columns: repeat(2, max-content);
    gap: var(--padding-medium) calc(var(--padding-large) * 2);

    @media (min-width: $breakpoint-md) {
      grid-template-columns: repeat(4, max-content);
      grid-template-rows: repeat(3, auto);
      grid-auto-flow: column;
    }
  }

  &__details-row {
    display: grid;
    grid-template-columns: subgrid;
    grid-column: span 2;
    column-gap: var(--padding-large);
    align-items: center;
  }

  &__details-label {
    font-weight: var(--standard_text_bold_font_weight);
  }

  &__details-value {
    display: flex;
  }

  &__details-footer-text {
    display: flex;
    color: var(--text-color-discreet);
    gap: var(--padding);
    padding-top: var(--padding-medium);
  }

  &__type-badge {
    width: 5rem;
  }

  &__desktop-icon {
    margin-right: var(--padding);
    color: var(--text-color-discreet);
  }

  &__detail-value-tooltip {
    margin-left: var(--padding);
  }
}
</style>
