<script setup lang="ts">
import type { VDBExecutionDepositsTableRow } from '~/types/api/validator_dashboard'

const {
  hasValidators,
  key,
  variant,
} = useDashboard()

const { t: $t } = useTranslation()

const { width } = useWindowSize()
const colsVisible = computed(() => {
  return {
    block: width.value >= 768,
    depositor: width.value >= 1042,
    group: width.value > 768,
    validatorIndex: width.value >= 768,
    validity: width.value >= 768,
  }
})

const query = useDefaultQuery({
  sort: 'timestamp:desc',
})
const {
  data: elDeposits,
  status,
} = useApi(`/api/bff/validator-dashboards/${key.value}/execution-layer-deposits`, {
  immediate: key.value !== undefined,
  query,
})

const tableData = computed(() => {
  if (!elDeposits.value?.data.length) {
    return null
  }

  return {
    data: [
      {
        amount: elDeposits.value?.total_amount,
        isTotalAmountRow: true,
      },
      ...elDeposits.value.data,
    ],
    paging: elDeposits.value.paging,
  }
})

const { groups } = useValidatorDashboardGroups()
const getGroupName = (groupId: number) => {
  return groups.value.find(group => group.id === groupId)?.name
}

const {
  displayCurrencyDefault,
  selectedCurrencyMain,
} = useCurrency()
const v1Domain = useV1Domain()
const emit = defineEmits<{
  (e: 'add-validator'): void,
}>()
</script>

<template>
  <BcTableControl
    v-model:search="query.search"
    :title="$t('dashboard.validator.el_deposits.title')"
    :search-placeholder="$t(
      variant === 'guest-dashboard'
        ? 'dashboard.validator.el_deposits.search_placeholder_guest_dashboard'
        : 'dashboard.validator.el_deposits.search_placeholder_private_dashboard',
    )
    "
  >
    <template #table>
      <ClientOnly fallback-tag="span">
        <BcTable
          :data="tableData"
          :query
          data-key="identifier"
          expandable
          table-class="dashboard-table-el-deposits"
          :cursor="query?.cursor"
          :selected-sort="query?.sort"
          :page-size="query?.limit"
          :row-class="(row: VDBExecutionDepositsTableRow) => row.index === undefined ? 'total-row' : ''"
          :is-row-expandable="(row: VDBExecutionDepositsTableRow) => row.index !== undefined"
          :is-loading="status === 'pending'"
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
                v-else
                :unix-timestamp="slotProps.data.timestamp"
              />
            </template>
          </Column>
          <Column
            v-if="colsVisible.validatorIndex"
            field="index"
            :header="$t('dashboard.validator.col.validator_index')"
          >
            <template #body="slotProps">
              <BcIcon
                v-if="!slotProps.data.isTotalAmountRow"
                name="desktop"
                size="sm"
                class="dashboard-table-el-deposits__desktop-icon"
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
            v-if="colsVisible.group"
            field="group_id"
            body-class="group-id"
            header-class="group-id"
            :header="$t('dashboard.validator.col.group')"
          >
            <template #body="slotProps">
              <span v-if="!slotProps.data.isTotalAmountRow">
                {{ getGroupName(slotProps.data.group_id) }}
              </span>
            </template>
          </Column>
          <Column
            v-if="colsVisible.block"
            field="block"
            sortable
            :header="$t('common.block')"
          >
            <template #body="slotProps">
              <BcLink
                v-if="!slotProps.data.isTotalAmountRow"
                :to="`${v1Domain}/block/${slotProps.data.block}`"
                target="_blank"
                class="link"
              >
                <BcFormatNumber :value="slotProps.data.block" />
              </BcLink>
            </template>
          </Column>
          <Column
            v-if="colsVisible.depositor"
            field="depositor"
            :header="$t('dashboard.validator.col.depositor')"
          >
            <template #body="slotProps">
              <BcFormatHash
                v-if="!slotProps.data.isTotalAmountRow"
                :hash="slotProps.data.depositor.hash"
                :ens="slotProps.data.depositor.ens"
                :no-wrap="true"
                type="address"
              />
            </template>
          </Column>
          <Column
            v-if="colsVisible.validity"
            field="validity"
            :header="$t('table.validity')"
          >
            <template #body="slotProps">
              <div
                v-if="!slotProps.data.isTotalAmountRow"
                class="status-cell-content"
              >
                <DashboardTableElDepositsValidity
                  :validity="slotProps.data.validity"
                />
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
            v-if="!colsVisible.validity"
            field="validity"
          >
            <template #body="slotProps">
              <div v-if="!slotProps.data.isTotalAmountRow">
                <DashboardTableElDepositsValidity
                  is-compact
                  :validity="slotProps.data.validity"
                />
              </div>
            </template>
          </Column>
          <template #expansion="slotProps">
            <div class="expansion">
              <div
                v-if="!colsVisible.validatorIndex"
                class="row"
              >
                <div class="label">
                  {{ $t("dashboard.validator.col.validator_index") }}
                </div>
                <div class="value">
                  <BcIcon
                    v-if="!slotProps.data.isTotalAmountRow"
                    name="desktop"
                    size="sm"
                    class="dashboard-table-el-deposits__desktop-icon"
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
                v-if="!colsVisible.group"
                class="row"
              >
                <div class="label">
                  {{ $t("dashboard.validator.col.group") }}
                </div>
                <div class="value">
                  {{ getGroupName(slotProps.data.group_id) }}
                </div>
              </div>

              <div class="row">
                <div class="label">
                  {{ $t("block.col.transaction_hash") }}
                </div>
                <BcFormatHash
                  v-if="!slotProps.data.isTotalAmountRow"
                  :hash="slotProps.data.tx_hash"
                  :no-wrap="true"
                  type="tx"
                />
              </div>

              <div
                v-if="!colsVisible.block"
                class="row"
              >
                <div class="label">
                  {{ $t("common.block") }}
                </div>
                <BcLink
                  :to="`${v1Domain}/block/${slotProps.data.block}`"
                  target="_blank"
                  class="link"
                >
                  <BcFormatNumber :value="slotProps.data.block" />
                </BcLink>
              </div>

              <div class="row">
                <div class="label">
                  {{ $t("dashboard.validator.col.public_key") }}
                </div>
                <BcFormatHash
                  :hash="slotProps.data.public_key"
                  type="public_key"
                  :no-wrap="true"
                />
              </div>

              <div class="row">
                <div class="label">
                  {{ $t("dashboard.validator.col.withdrawal_credential") }}
                </div>
                <BcFormatHash
                  :hash="slotProps.data.withdrawal_credential"
                  type="withdrawal_credentials"
                  :no-wrap="true"
                />
              </div>

              <div
                v-if="!colsVisible.depositor"
                class="row"
              >
                <div class="label">
                  {{ $t("dashboard.validator.col.depositor") }}
                </div>
                <BcFormatHash
                  v-if="!slotProps.data.isTotalAmountRow"
                  :hash="slotProps.data.depositor.hash"
                  :ens="slotProps.data.depositor.ens"
                  :no-wrap="true"
                  type="address"
                />
              </div>
              <div
                v-if="!colsVisible.validity"
                class="row"
              >
                <div class="label">
                  {{ $t("dashboard.validator.col.validity") }}
                </div>
                <DashboardTableElDepositsValidity
                  :validity="slotProps.data.validity"
                />
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
@use "~/assets/css/utils.scss";
@use '~/assets/css/breakpoints' as *;

:deep(.dashboard-table-el-deposits) {
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
      white-space: nowrap;
      overflow: visible;
    }
  }

  .age-field {
    padding: 0 7px;
    white-space: nowrap;
  }

  .status-cell-content {
    display: flex;
    align-items: center;
  }
}

.dashboard-table-el-deposits__desktop-icon {
  margin-right: var(--padding);
  color: var(--text-color-discreet);
}

.expansion {
  display: grid;
  grid-template-columns: repeat(2, max-content);
  gap: var(--padding-medium) calc(var(--padding-large) * 2);
  padding: var(--padding-medium);
  background-color: var(--container-background);
  color: var(--container-color);
  font-size: var(--small_text_font_size);

  @media (min-width: $breakpoint-md) {
    grid-template-columns: repeat(4, max-content);
    grid-template-rows: repeat(3, auto);
    grid-auto-flow: column;
  }

  .row {
    display: grid;
    grid-template-columns: subgrid;
    grid-column: span 2;
    column-gap: var(--padding-large);
    align-items: center;

    .label {
      font-weight: var(--standard_text_bold_font_weight);
    }

    .value {
      display: flex
    }
  }
}
</style>
