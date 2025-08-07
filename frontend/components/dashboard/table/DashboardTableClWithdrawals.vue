<script setup lang="ts">
import type { DataTableSortEvent } from 'primevue/datatable'
import type {
  GetValidatorDashboardConsensusLayerWithdrawalsResponse,
  GetValidatorDashboardTotalConsensusWithdrawalsResponse,
  VDBWithdrawalsClTableRow,
} from '~/types/api/validator_dashboard'
import type {
  Cursor, TableQueryParams,
} from '~/types/datatable'

const {
  clWithdrawals,
  clWithdrawalsTotalAmount,
} = defineProps<{
  clWithdrawals?: GetValidatorDashboardConsensusLayerWithdrawalsResponse,
  clWithdrawalsTotalAmount?: GetValidatorDashboardTotalConsensusWithdrawalsResponse,
  isLoading: boolean,
}>()

const {
  isGuestDashboard,
} = useDashboardKey()
const { t: $t } = useTranslation()
const { getTimestampFromSlot } = useNetwork()

const { width } = useWindowSize()
const isMobile = computed(() => {
  return width.value < 768
})

const validatorDashboardOverviewStore = useValidatorDashboardOverviewStore()
const {
  hasValidators,
} = storeToRefs(validatorDashboardOverviewStore)

const clWithdrawalsWithIdentifiers = computed(() =>
  addIdentifier(clWithdrawals, 'slot', 'slot_index'),
)
const tableData = computed(() => {
  if (!clWithdrawalsWithIdentifiers.value?.data?.length) {
    return
  }

  return {
    data: [
      {
        amount: clWithdrawalsTotalAmount?.data.total_amount,
        isTotalAmountRow: true,
      },
      ...clWithdrawalsWithIdentifiers.value.data,
    ],
    paging: clWithdrawalsWithIdentifiers.value.paging,
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
    :title="$t('dashboard.validator.cl_withdrawals.title')"
    :search-placeholder="
      $t(
        isGuestDashboard
          ? 'dashboard.validator.cl_withdrawals.search_placeholder_guest_dashboard'
          : 'dashboard.validator.cl_withdrawals.search_placeholder_private_dashboard',
      )
    "
    @set-search="setSearch"
  >
    <template #table>
      <ClientOnly fallback-tag="span">
        <BcTable
          :data="tableData"
          expandable
          table-class="dashboard-table-cl-withdrawals"
          data-key="identifier"
          :selected-sort="query?.sort"
          :cursor="query?.cursor"
          :page-size="query?.limit"
          :row-class="(row: VDBWithdrawalsClTableRow) => row.status === 'queued' ? 'grayed-out-row' : ''"
          :is-row-expandable="(row: VDBWithdrawalsClTableRow) => row.index !== undefined"
          @set-cursor="setCursor"
          @sort="onSort"
          @set-page-size="setPageSize"
        >
          <Column
            sortable
            body-class="dashboard-table-cl-withdrawals__age-cell"
            field="timestamp"
          >
            <template #header>
              <BcTableAgeHeader />
            </template>
            <template #body="slotProps">
              <span v-if="slotProps.data.isTotalAmountRow">Σ</span>
              <BcTableDateTime
                v-else-if="slotProps.data.slot !== undefined"
                :unix-timestamp="getTimestampFromSlot(slotProps.data.slot)"
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
                class="dashboard-table-cl-withdrawals__desktop-icon"
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
            :header="$t('dashboard.validator.col.type')"
          >
            <template #body="slotProps">
              <div class="dashboard-table-cl-withdrawals__type-cell">
                <BcBadge
                  v-if="!slotProps.data.isTotalAmountRow"
                  class="dashboard-table-cl-withdrawals__type-badge"
                  color="gray"
                >
                  {{
                    slotProps.data.type === 'manual'
                      ? $t('dashboard.validator.cl_withdrawals.type.manual')
                      : $t('dashboard.validator.cl_withdrawals.type.auto')
                  }}
                </BcBadge>
                <BcTooltip
                  v-if="
                    (slotProps.data.type === 'skimming' || slotProps.data.type === 'system')
                      && slotProps.data.status === 'queued'"
                  tooltip-width="280px"
                  tooltip-text-align="left"
                  class="dashboard-table-cl-withdrawals__type-tooltip-trigger"
                >
                  <BcIcon
                    name="circle-info"
                  />
                  <template #tooltip>
                    <slot name="tooltip">
                      {{ $t('dashboard.validator.cl_withdrawals.info_auto_queued') }}
                    </slot>
                  </template>
                </BcTooltip>
              </div>
            </template>
          </Column>
          <Column
            v-if="!isMobile"
            field="status"
            :header="$t('table.status')"
          >
            <template #body="slotProps">
              <DashboardTableClWithdrawalsStatus
                v-if="!slotProps.data.isTotalAmountRow"
                :status="slotProps.data.status"
                :reject-reason="slotProps.data.reject_reason"
              />
            </template>
          </Column>
          <Column
            field="amount"
            sortable
            :header="$t('table.amount')"
          >
            <template #body="slotProps">
              <BcFormatAmount
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
              <DashboardTableClWithdrawalsStatus
                v-if="!slotProps.data.isTotalAmountRow"
                :status="slotProps.data.status"
                is-compact
              />
            </template>
          </Column>
          <template #expansion="slotProps">
            <div class="dashboard-table-cl-withdrawals__details">
              <div class="dashboard-table-cl-withdrawals__details-grid">
                <div
                  v-if="isMobile"
                  class="dashboard-table-cl-withdrawals__details-row"
                >
                  <div class="dashboard-table-cl-withdrawals__details-label">
                    {{ $t("dashboard.validator.col.validator_index") }}
                  </div>
                  <div class="dashboard-table-cl-withdrawals__details-value">
                    <BcIcon
                      name="desktop"
                      size="sm"
                      class="dashboard-table-cl-withdrawals__desktop-icon"
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
                  class="dashboard-table-cl-withdrawals__details-row"
                >
                  <div class="dashboard-table-cl-withdrawals__details-label">
                    {{ $t("dashboard.validator.col.group") }}
                  </div>
                  <div class="value">
                    {{ getGroupName(slotProps.data.group_id) }}
                  </div>
                </div>
                <div
                  v-if="isMobile"
                  class="dashboard-table-cl-withdrawals__details-row"
                >
                  <div class="dashboard-table-cl-withdrawals__details-label">
                    {{ $t("dashboard.validator.col.type") }}
                  </div>
                  <div class="dashboard-table-cl-withdrawals__details-value">
                    <BcBadge
                      v-if="slotProps.data.index !== undefined"
                      class="dashboard-table-cl-withdrawals__type-badge"
                      color="gray"
                    >
                      {{
                        slotProps.data.type === 'manual'
                          ? $t('dashboard.validator.cl_withdrawals.type.manual')
                          : $t('dashboard.validator.cl_withdrawals.type.auto')
                      }}
                    </BcBadge>
                    <BcTooltip
                      v-if="slotProps.data.type === 'auto' && slotProps.data.status === 'queued'"
                      tooltip-width="280px"
                      tooltip-text-align="left"
                      class="dashboard-table-cl-withdrawals__type-tooltip-trigger"
                    >
                      <BcIcon
                        name="circle-info"
                      />
                      <template #tooltip>
                        <slot name="tooltip">
                          {{ $t('dashboard.validator.cl_withdrawals.info_auto_queued') }}
                        </slot>
                      </template>
                    </BcTooltip>
                  </div>
                </div>
                <div
                  v-if="isMobile"
                  class="dashboard-table-cl-withdrawals__details-row"
                >
                  <div class="dashboard-table-cl-withdrawals__details-label">
                    {{ $t("table.status") }}
                  </div>
                  <div class="dashboard-table-cl-withdrawals__details-value">
                    <DashboardTableClWithdrawalsStatus
                      :status="slotProps.data.status"
                      :reject-reason="slotProps.data.reject_reason"
                    />
                  </div>
                </div>
                <div class="dashboard-table-cl-withdrawals__details-row">
                  <div class="dashboard-table-cl-withdrawals__details-label">
                    {{ $t("dashboard.validator.col.slot_queued") }}
                  </div>
                  <BcLink
                    v-if="slotProps.data.slot_queued !== undefined"
                    :to="`${v1Domain}/slot/${slotProps.data.slot_queued}`"
                    target="_blank"
                    class="link"
                  >
                    <BcFormatNumber :value="slotProps.data.slot_queued" />
                  </BcLink>
                  <span v-else>-</span>
                </div>
                <div
                  class="dashboard-table-cl-withdrawals__details-row"
                >
                  <div class="dashboard-table-cl-withdrawals__details-label">
                    {{ $t("dashboard.validator.col.slot_processed") }}
                  </div>
                  <div class="dashboard-table-cl-withdrawals__details-value">
                    <BcLink
                      v-if="slotProps.data.slot_processed !== undefined"
                      :to="`${v1Domain}/slot/${slotProps.data.slot_processed}`"
                      target="_blank"
                      class="link"
                    >
                      <BcFormatNumber :value="slotProps.data.slot_processed" />
                    </BcLink>
                    <span v-else>-</span>

                    <BcTooltip
                      v-if="slotProps.data.status === 'queued'"
                      tooltip-width="200px"
                      tooltip-text-align="left"
                      class="dashboard-table-cl-withdrawals__detail-value-tooltip"
                    >
                      <BcIcon name="circle-info" />
                      <template #tooltip>
                        <slot name="tooltip">
                          {{ $t('dashboard.validator.cl_withdrawals.info_slot_processed_estimated') }}
                        </slot>
                      </template>
                    </BcTooltip>
                  </div>
                </div>
                <div class="dashboard-table-cl-withdrawals__details-row">
                  <div class="dashboard-table-cl-withdrawals__details-label">
                    {{ $t("dashboard.validator.col.recipient") }}:
                  </div>
                  <BcFormatHash
                    type="address"
                    :hash="slotProps.data.recipient.hash"
                    :ens="slotProps.data.recipient.ens"
                  />
                </div>
                <div class="dashboard-table-cl-withdrawals__details-row">
                  <div class="dashboard-table-cl-withdrawals__details-label">
                    {{ $t("dashboard.validator.col.public_key") }}
                  </div>
                  <BcFormatHash
                    :hash="slotProps.data.public_key"
                    type="public_key"
                    :no-wrap="true"
                  />
                </div>
                <div class="dashboard-table-cl-withdrawals__details-row">
                  <div class="dashboard-table-cl-withdrawals__details-label">
                    {{ $t("dashboard.validator.col.withdrawal_credential") }}
                  </div>
                  <BcFormatHash
                    :hash="slotProps.data.withdrawal_credentials"
                    type="withdrawal_credentials"
                    :no-wrap="true"
                  />
                </div>
              </div>
              <span
                v-if="slotProps.data.type === 'skimming'"
                class="dashboard-table-cl-withdrawals__details-footer-text"
              >
                {{ $t('dashboard.validator.cl_withdrawals.info_type_skimming') }}
              </span>
              <span
                v-else-if="slotProps.data.type === 'system'"
                class="dashboard-table-cl-withdrawals__details-footer-text"
              >
                {{ $t('dashboard.validator.cl_withdrawals.info_type_system') }}
              </span>
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

.dashboard-table-cl-withdrawals__details {
  padding: var(--padding-medium);
  color: var(--container-color);
  font-size: var(--small_text_font_size);
  background-color: var(--container-background);
}

.dashboard-table-cl-withdrawals__details-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, max-content));
  gap: var(--padding-medium) calc(var(--padding-large) * 2);

  @media (min-width: $breakpoint-md) {
    grid-template-columns: repeat(4, minmax(0, max-content));
    grid-template-rows: repeat(3, auto);
    grid-auto-flow: column;
  }
}

.dashboard-table-cl-withdrawals__details-row {
  display: grid;
  grid-template-columns: subgrid;
  grid-column: span 2;
  column-gap: var(--padding-large);
  align-items: center;
}

.dashboard-table-cl-withdrawals__details-label {
  font-weight: var(--standard_text_bold_font_weight);
}

.dashboard-table-cl-withdrawals__details-value {
  display: flex;
}

.dashboard-table-cl-withdrawals__details-footer-text {
  display: flex;
  color: var(--text-color-discreet);
  gap: var(--padding);
  padding-top: var(--padding-medium);
  line-height: 1.25rem;
}

.dashboard-table-cl-withdrawals__detail-value-tooltip {
  margin-left: var(--padding);
}

.dashboard-table-cl-withdrawals__desktop-icon {
  margin-right: var(--padding);
  color: var(--text-color-discreet)
}

.dashboard-table-cl-withdrawals__type-badge {
  width: 5rem;
}

.dashboard-table-cl-withdrawals__type-tooltip-trigger {
  display: flex;
  margin-left: var(--padding-medium);
}

:deep(.dashboard-table-cl-withdrawals) {
  .grayed-out-row > td > *  {
    opacity: 0.5;
  }

  .dashboard-table-cl-withdrawals__age-cell {
    padding-top: 0 !important;
    padding-bottom: 0 !important;
  }

  .dashboard-table-cl-withdrawals__type-cell {
    display: flex;
    align-items: center;
  }
}
</style>
