<script lang="ts" setup>
import { warn } from 'vue'
import {
  BcDialogConfirm,
  BcPremiumModal,
  DashboardGroupSelectionDialog,
} from '#components'
import { useValidatorDashboardOverviewStore } from '~/stores/dashboard/useValidatorDashboardOverviewStore'
import type {
  PostValidatorDashboardValidatorsRequest,
  VDBManageValidatorsTableRow,
  VDBPostValidatorsData,
} from '~/types/api/validator_dashboard'
import type { NumberOrString } from '~/types/value'

import type { InternalPostSearchResponse } from '~/types/api/search'

const { t: $t } = useTranslation()
const { fetch } = useCustomFetch()

const { width } = useWindowSize()

const dialog = useDialog()

const validatorDashboardOverviewStore = useValidatorDashboardOverviewStore()
const {
  overview,
} = storeToRefs(validatorDashboardOverviewStore)
const { refreshOverview } = validatorDashboardOverviewStore

const selectedGroup = ref<number>(-1)
const {
  addEntities,
  dashboardKey,
  isGuestDashboard,
  removeEntities,
} = useDashboardKey()
const {
  premium_perks,
} = useUserStore()

const {
  data,
  isLoading,
  query,
  refresh,
} = useTable('DASHBOARD_VALIDATOR_MANAGEMENT', {
  apiPath: 'DASHBOARD_VALIDATOR_MANAGEMENT',
  pathValues: {
    dashboardKey: dashboardKey.value,
  },
})

const selected = ref<VDBManageValidatorsTableRow[]>()
const hasNoOpenDialogs = ref(true)

const size = computed(() => {
  return {
    expandable: width.value < 1060,
    showBalance: width.value >= 1060,
    showGroup: width.value >= 925,
    showPublicKey: width.value >= 570,
    showWithdrawalCredentials: width.value >= 750,
  }
})

const mapIndexOrPubKey = (
  validators?: VDBManageValidatorsTableRow[],
) => {
  return [ ...new Set(validators?.map(
    validator => validator.index ?? validator.public_key)) ]
}

const changeGroup = async (body: PostValidatorDashboardValidatorsRequest, groupId?: number) => {
  if (
    !body.validators?.length
    && !body.deposit_address
    && !body.graffiti
    && !body.withdrawal_credential
  ) {
    warn('no validators selected to change group')
    return
  }
  body.group_id = groupId && groupId !== -1 ? groupId : 0

  await fetch<VDBPostValidatorsData>(
    'DASHBOARD_VALIDATOR_MANAGEMENT',
    {
      body,
      method: 'POST',
    },
    { dashboardKey: dashboardKey.value },
  ).then(() => {
    refresh()
    refreshOverview(dashboardKey.value)
  }).then(() => {
    resetInput()
  }).catch((error) => {
    if (error.statusCode === 403) {
      dialog.open(BcPremiumModal, {
        data: {
          description: $t('dashboard.validator.management.validators_limit_exceeded'),
        },
      })
    }
  })
}
const toast = useBcToast()

const removeValidators = async (validators: NumberOrString[]) => {
  if (!validators?.length) {
    warn('no validators selected to change group')
    return
  }
  if (isGuestDashboard.value) {
    removeEntities(validators.map(v => v.toString()))
    return
  }

  await fetch(
    'DASHBOARD_VALIDATOR_MANAGEMENT_DELETE',
    {
      body: { validators },
      method: 'POST',
    },
    { dashboardKey: dashboardKey.value },
  ).then(
    () => {
      refresh()
      refreshOverview(dashboardKey.value)
    },
  ).catch(() => {
    toast.showError({
      detail: $t('dashboard.validator.management.removal_failed.detail'),
      summary: $t('dashboard.validator.management.removal_failed.summary'),
    })
  })
}

const editSelected = () => {
  hasNoOpenDialogs.value = false
  dialog.open(DashboardGroupSelectionDialog, {
    data: {
      groupId: selected.value?.[0]?.group_id ?? undefined,
      selectedValidators: selected.value?.length,
      totalValidators: totalValidators?.value,
    },
    onClose: (response) => {
      hasNoOpenDialogs.value = true
      if (response?.data !== undefined) {
        changeGroup(
          { validators: mapIndexOrPubKey(selected.value) },
          response?.data,
        )
      }
    },
  })
}

const switchValidatorGroup = (
  row: VDBManageValidatorsTableRow,
  group: number,
) => {
  changeGroup(
    { validators: mapIndexOrPubKey([ row ].concat(selected.value ?? [])) },
    group,
  )
}

const removeRow = (row: VDBManageValidatorsTableRow) => {
  const list = mapIndexOrPubKey([ row ].concat(selected.value ?? []))
  if (!list?.length) {
    warn('no validator to remove')
  }

  hasNoOpenDialogs.value = false
  dialog.open(BcDialogConfirm, {
    data: {
      question: $t(
        'dashboard.validator.management.remove_text',
        { validator: list[0] },
        list.length,
      ),
      title: $t('dashboard.validator.management.remove_title'),
    },
    onClose: (response) => {
      hasNoOpenDialogs.value = true
      if (response?.data) {
        removeValidators(list)
      }
    },
  })
}

const totalValidators = computed(() => {
  // this is necessary after an `typescript update`
  // for types created by api, we should use `types` instead of `interfaces`
  return addUpValues(overview.value?.validators as unknown as Record<string, number>)
})

const {
  premiumProducts,
} = useProductsStore()

const latestEffectiveBalance = computed(() => {
  return overview.value?.balances.effective_latest
})

const effectiveBalanceLimitPerDashboard = computed(() => {
  const freeProduct = premiumProducts.value['Free']
  const effectiveBalanceLimitFreeProduct = freeProduct?.premium_perks.effective_balance_per_dashboard

  return premium_perks.value?.effective_balance_per_dashboard ?? effectiveBalanceLimitFreeProduct
})

const hasReachedLimit = computed(() => {
  if (!latestEffectiveBalance.value || !effectiveBalanceLimitPerDashboard.value) {
    return false
  }
  return isGreaterEquals(latestEffectiveBalance.value, effectiveBalanceLimitPerDashboard.value)
})

const hasPremiumPerkBulkAdding = computed(() => !!premium_perks.value?.bulk_adding)

const handleInvalidSubmit = () => {
  dialog.open(BcPremiumModal, {})
}
const resetInput = () => {
  inputValidator.value = ''
}
const handleSubmit = (item: InternalPostSearchResponse['data'][number] | undefined) => {
  if (!item) return
  const {
    type,
    value,
  } = item
  if (hasReachedLimit.value) {
    handleInvalidSubmit()
    return
  }
  if (
    !hasPremiumPerkBulkAdding.value
    && (type !== 'validator' && type !== 'validator_list')
  ) {
    handleInvalidSubmit()
    return
  }
  if (isGuestDashboard.value) {
    if (item.type === 'validator') {
      addEntities([ `${item.value.index}` ])
      resetInput()
      return
    }
    if (item.type === 'validator_list') {
      addEntities(
        item.value.validators
          .map(validator => `${validator}`),
      )
      resetInput()
      return
    }
    handleInvalidSubmit()
    return
  }
  changeGroup({
    ...(type === 'validator' && { validators: [ value.index ] }),
    ...(type === 'validator_list' && { validators: value.validators }),
    ...(type === 'validators_by_deposit_address' && { deposit_address: value.deposit_address }),
    ...(type === 'validators_by_withdrawal_credential' && { withdrawal_credential: value.withdrawal_credential }),
    ...(type === 'validators_by_graffiti' && { graffiti: value.graffiti }),
  },
  selectedGroup.value,
  )
}
const inputValidator = ref('')

const emit = defineEmits<{
  (e: 'close'): void,
}>()
const isVisible = ref(true)
const close = () => {
  isVisible.value = false
  emit('close')
}
const onSetSearch = (value?: string) => {
  query.value.search = value
}
const filterGroups = (value?: number) => {
  query.value.group_id = value
}
</script>

<template>
  <BcDialog
    v-model="isVisible"
    :header="$t('dashboard.validator.management.title')"
    :close-on-escape="hasNoOpenDialogs"
    class="validator-managment-modal-container"
    @hide="close"
  >
    <template
      v-if="!size.showWithdrawalCredentials"
      #header
    >
      <span />
    </template>
    <BcTableControl
      :search-placeholder="
        $t(
          isGuestDashboard
            ? 'dashboard.validator.summary.search_placeholder_public'
            : 'dashboard.validator.summary.search_placeholder',
        )
      "
      @set-search="onSetSearch"
    >
      <template #header-left>
        <span v-if="size.showWithdrawalCredentials">
          {{ $t("dashboard.validator.management.sub_title") }}</span>
        <span
          v-else
          class="small-title"
        >{{
          $t("dashboard.validator.manage_validators")
        }}</span>
      </template>
      <template #bc-table-sub-header>
        <div class="add-row">
          <DashboardGroupSelection
            v-model="selectedGroup"
            :include-all="true"
            class="small group-selection"
            @update:model-value="filterGroups"
          />
          <DashboardValidatorManagementModalSearch
            v-model="inputValidator"
            class="search-bar"
            :has-premium-perk-bulk-adding
            :has-reached-limit
            :is-guest-dashboard
            @submit="handleSubmit"
          />
        </div>
      </template>
      <template #table>
        <ClientOnly fallback-tag="span">
          <BcTable
            v-model:selection="selected"
            v-model:query="query"
            :data="data ?? undefined"
            :loading="isLoading"
            data-key="public_key"
            :expandable="size.expandable"
            has-selection-mode
            class="management-table"
          >
            <Column
              field="index"
              :sortable="true"
              :header="$t('dashboard.validator.col.index')"
            />

            <Column
              v-if="size.showPublicKey"
              field="public_key"
              :sortable="!size.expandable"
              :header="$t('dashboard.validator.col.public_key')"
            >
              <template #body="slotProps">
                <BcFormatHash
                  :hash="slotProps.data.public_key"
                  type="public_key"
                  class="public-key"
                />
              </template>
            </Column>
            <Column
              v-if="size.showGroup"
              field="group_id"
              :sortable="!size.expandable"
              :header="$t('dashboard.validator.col.group')"
            >
              <template #body="slotProps">
                <DashboardGroupSelection
                  v-model="slotProps.data.group_id"
                  class="small group-selection"
                  @set-group="
                    (id: number) => switchValidatorGroup(slotProps.data, id)
                  "
                />
              </template>
            </Column>
            <Column
              v-if="size.showBalance"
              field="balance"
              :sortable="true"
              :header="$t('dashboard.validator.col.balance')"
            >
              <template #body="slotProps">
                <div class="balance-col">
                  <BcFormatAmount
                    :value="slotProps.data.balance"
                  />
                </div>
              </template>
            </Column>
            <Column
              field="status"
              :sortable="!size.expandable"
              header-class="status-col"
              :header="$t('dashboard.validator.col.status')"
            >
              <template #body="slotProps">
                <ValidatorTableStatus
                  :status="slotProps.data.status"
                  :hide-label="size.expandable"
                />
              </template>
            </Column>
            <Column
              v-if="size.showWithdrawalCredentials"
              field="withdrawal_credential"
              :sortable="!size.expandable"
              :header="$t('dashboard.validator.col.withdrawal_credential')"
            >
              <template #body="slotProps">
                <div class="withdrawal-col">
                  <BcFormatHash
                    :hash="slotProps.data.withdrawal_credential"
                    type="withdrawal_credentials"
                  />
                </div>
              </template>
            </Column>
            <Column field="action">
              <template #header>
                <Button
                  v-show="selected?.length"
                  class="edit-button"
                  @click.stop.prevent="editSelected()"
                >
                  <span class="edit-label">{{ $t("common.edit") }}</span>
                  <BcIcon
                    class="edit-icon"
                    name="edit"
                  />
                </Button>
              </template>
              <template #body="slotProps">
                <div class="action-col">
                  <BcButtonIcon
                    class="remove-button"
                    :screenreader-text="{
                      key: 'dashboard.validator.management.remove_validator',
                      interpolation: { validatorIndex: slotProps.data.index },
                    }"
                    name="trash"
                    @click="removeRow(slotProps.data)"
                  />
                </div>
              </template>
            </Column>
            <template #expansion="slotProps">
              <div class="expansion">
                <div class="info">
                  <div class="label">
                    {{ $t("dashboard.validator.col.public_key") }}
                  </div>
                  <BcFormatHash
                    :hash="slotProps.data.public_key"
                    type="public_key"
                    class="public-key"
                  />
                </div>
                <div class="info">
                  <div class="label">
                    {{ $t("dashboard.validator.col.balance") }}
                  </div>
                  <BcFormatAmount :value="slotProps.data.balance" />
                </div>
                <div class="info">
                  <div class="label">
                    {{ $t("dashboard.validator.col.group") }}
                  </div>
                  <DashboardGroupSelection
                    v-model="slotProps.data.group_id"
                    class="small"
                    @set-group="
                      (id: number) => switchValidatorGroup(slotProps.data, id)
                    "
                  />
                </div>
                <div class="info">
                  <div class="label">
                    {{ $t("dashboard.validator.col.status") }}
                  </div>
                  <ValidatorTableStatus
                    :status="slotProps.data.status"
                  />
                </div>
                <div class="info">
                  <div class="label" />
                  <BcFormatHash
                    :hash="slotProps.data.withdrawal_credential"
                    type="withdrawal_credentials"
                  />
                </div>
              </div>
            </template>

            <template #bc-table-footer-left>
              <div
                class="left"
              >
                <div
                  class="labels"
                  :class="{ 'premium-limit': hasReachedLimit }"
                >
                  <span>
                    <BcFormatAmount
                      :value="latestEffectiveBalance ?? '0'"
                      :maximum-fraction-digits="0"
                      target-currency="mainDisplayCurrency"
                    />
                    /
                    <BcFormatAmount
                      :value="effectiveBalanceLimitPerDashboard || '0'"
                      :maximum-fraction-digits="0"
                      target-currency="mainDisplayCurrency"
                    />
                  </span>
                </div>
                <BcPremiumGem />
              </div>
            </template>

            <template #bc-table-footer-right>
              <Button
                :label="$t('navigation.done')"
                @click="close"
              />
            </template>
          </BcTable>
        </ClientOnly>
      </template>
    </BcTableControl>
  </BcDialog>
</template>

<style lang="scss" scoped>
@use "~/assets/css/main.scss";
@use "~/assets/css/utils.scss";
@use "~/assets/css/fonts.scss";
@use '~/assets/css/breakpoints' as *;

:global(.validator-managment-modal-container) {
  width: 1060px;
  height: 800px;
}

:global(.validator-managment-modal-container .p-dialog-content) {
  display: flex;
  flex-direction: column;
  flex-grow: 1;
}

:global(.validator-managment-modal-container .bc-table-header) {
  height: unset !important;
  padding: var(--padding) 0 !important;
  @include fonts.subtitle_text;
}

:global(
    .validator-managment-modal-container .bc-table-header .side:first-child
  ) {
  display: contents;
}

.small-title {
  @include utils.truncate-text;
  @include fonts.big_text;
}

.remove-button {
  color: var(--blue);
}

.group-selection {
  width: 6rem;
  @media (min-width: $breakpoint-md) {
    width: 10rem;
  }
}

.management-table {
  @include main.container;
  flex-grow: 1;
  display: flex;
  flex-direction: column;
  overflow-y: hidden;
  justify-content: space-between;
  padding-bottom: var(--padding-medium);

  :deep(.p-datatable-wrapper) {
    flex-grow: 1;
  }
}

.add-row {
  position: relative;
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--padding);
  gap: var(--padding);

  .search-bar {
    flex-shrink: 1;
    flex-grow: 1;
  }
}

.left {
  display: flex;
  margin-top: 4px;
  gap: var(--padding-small);

  .labels {
    display: flex;
    gap: var(--padding-small);

    &.premium-limit {
      color: var(--negative-color);
    }

    @media (max-width: 450px) {
      flex-direction: column;
    }
  }

  .gem {
    color: var(--primary-color);
  }
}

.edit-icon {
  margin-left: var(--padding-small);
}

.balance-col {
  width: 110px;
}

.withdrawal-col {
  width: 200px;
}

.action-col {
  width: 10px;
}

@media (max-width: 959px) {
  :deep(.edit-button) {
    padding: 8px 6px;

    .edit-label {
      display: none;
    }
  }

  .public-key {
    width: unset;
  }

  :deep(.status-col) {
    .p-column-title {
      width: 35px;
    }
  }
}

.expansion {
  @include main.container;
  padding: var(--padding);
  display: flex;
  flex-direction: column;
  gap: var(--padding);
  font-size: var(--small_text_font_size);

  .info {
    display: flex;
    align-items: center;
    gap: var(--padding);

    .label {
      font-weight: var(--standard_text_bold_font_weight);
      width: 100px;
    }

    :nth-child(2) {
      max-width: 160px;
    }
  }
}
</style>
