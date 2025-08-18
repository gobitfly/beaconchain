<script lang="ts" setup>
// import type { DataTableSortEvent } from 'primevue/datatable'
import { warn } from 'vue'
import { FetchError } from 'ofetch'
import {
  BcDialogConfirm,
  BcPremiumModal,
  DashboardGroupSelectionDialog,
} from '#components'
import type {
  // GetValidatorDashboardValidatorsResponse,
  PostValidatorDashboardValidatorsRequest,
  VDBManageValidatorsTableRow,
  // VDBPostValidatorsData,
} from '~/types/api/validator_dashboard'
import type { Cursor } from '~/types/datatable'

// import type { PathValues } from '~/types/customFetch'
import type { InternalPostSearchResponse } from '~/types/api/search'
import { ERROR_CODE } from '~/shared/utils/helper'
// import type { Query } from '~/types/table'
const { t: $t } = useTranslation()
// const { fetch } = useCustomFetch()

const { width } = useWindowSize()

const dialog = useDialog()

const visible = defineModel<boolean>()

// const validatorDashboardOverviewStore = useValidatorDashboardOverviewStore()
// const {
//   overview,
// } = storeToRefs(validatorDashboardOverviewStore)
// const { refreshOverview } = validatorDashboardOverviewStore

const cursor = ref<Cursor>()
// const pageSize = ref<number>(25)
const selectedGroup = ref()

const {
  premium_perks,
} = useUserStore()

// const initialQuery = {
//   limit: pageSize.value,
//   sort: 'index:asc',
// }

// const {
//   bounce: setQuery,
//   instant: instantQuery,
//   temp: tempQuery,
//   value: query,
// } = useDebounceValue<PathValues | undefined>(initialQuery, 500)

// const data = ref<GetValidatorDashboardValidatorsResponse | undefined>()
const selection = ref<VDBManageValidatorsTableRow[]>([])
const selectedValidators = computed(() => {
  return selection.value?.map(row => `${row.index}`) ?? []
})
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

const resetData = () => {
  // data.value = undefined
  selection.value = []
  selectedGroup.value = -1
  cursor.value = undefined
  // instantQuery(initialQuery)
}

const onClose = () => {
  resetData()
  visible.value = false
}

const {
  key,
  setValidators,
  totalValidators,
  validatorIds,
  variant,
} = useDashboard()

const {
  add,
  // refresh: refreshValidators,
  remove,
  validators,
} = useValidators()
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

  if (key.value !== undefined) {
    await add(key.value, body)
      .then(() => {
        emit('change-validators', [])
      })
  }
  // await $api(`/api/bff/validator-dashboards/${key.value}/validators`, {
  //   body,
  //   method: 'post',
  // }).then(() => {
  //   refreshValidators()
  // })

  // await fetch<VDBPostValidatorsData>(
  //   'DASHBOARD_VALIDATOR_MANAGEMENT',
  //   {
  //     body,
  //     method: 'POST',
  //   },
  //   { dashboardKey: id.value },
  // ).then(() => {
  //   loadData(id.value)
  //   refreshOverview(id.value)
  // })
}

const editSelected = () => {
  hasNoOpenDialogs.value = false
  dialog.open(DashboardGroupSelectionDialog, {
    data: {
      groupId: selection.value?.[0]?.group_id ?? undefined,
      selectedValidators: selection.value?.length,
      totalValidators: totalValidators?.value,
    },
    onClose: (response) => {
      hasNoOpenDialogs.value = true
      if (response?.data !== undefined) {
        changeGroup(
          { validators: selectedValidators.value },
          response?.data,
        )
      }
    },
  })
}

// const onSort = (sort: DataTableSortEvent) => {
//   setQuery(setQuerySort(sort, query?.value))
// }

// const setCursor = (value: Cursor) => {
//   cursor.value = value
//   setQuery(setQueryCursor(value, query?.value))
// }

// const setPageSize = (value: number) => {
//   pageSize.value = value
//   setQuery(setQueryPageSize(value, query?.value))
// }

// watch(selectedGroup, (value) => {
//   setQuery({
//     ...query?.value,
//     group_id: value,
//   })
// })

// const loadData = async (dashboardKey: string) => {
//   if (dashboardKey) {
// const testQ = JSON.stringify(query.value)
// const result = await fetch<GetValidatorDashboardValidatorsResponse>(
//   'DASHBOARD_VALIDATOR_MANAGEMENT',
//   undefined,
//   { dashboardKey },
//   query.value,
// )

// Make sure that during loading the query did not change
// if (testQ === JSON.stringify(query.value)) {
// data.value = result
// selection.value = []
// }
// }
//   else {
//     data.value = {
//       data: [],
//       paging: {},
//     }
//   }
// }

// watch(
//   () => [
//     visible.value,
//     query.value,
//   ],
//   () => {
//     if (visible.value) {
//       loadData(id.value)
//     }
//   },
//   { immediate: true },
// )

const switchValidatorGroup = (
  row: VDBManageValidatorsTableRow,
  group: number,
) => {
  changeGroup(
    { validators: selectedValidators.value },
    group,
  )
}
const removeRow = (row: VDBManageValidatorsTableRow) => {
  selection.value = [ ...new Set([
    row,
    ...selection.value,
  ]) ]
  // hasNoOpenDialogs.value = false
  dialog.open(BcDialogConfirm, {
    data: {
      question: $t(
        'dashboard.validator.management.remove_text',
        { validator: selectedValidators.value[0] },
        selectedValidators.value.length,
      ),
      title: $t('dashboard.validator.management.remove_title'),

    },
    onClose: (response) => {
      const shouldRemove = response?.data
      if (shouldRemove) {
        if (variant.value === 'guest-dashboard') {
          if (validators.value?.data) {
            validators.value.data = validators.value.data.filter(
              validator => !selectedValidators.value.includes(`${validator.index}`),
            )
          }
          emit('change-validators', validatorIds.value.filter(validatorId => !selectedValidators.value.includes(validatorId)))
          selection.value = []
          return
        }
        if (key.value) {
          remove(key.value, selectedValidators.value.map(Number))
        }
      }
      // selection.value = []
      // hasNoOpenDialogs.value = true
      // if (response?.data) {
      //   removeValidators(selectedValidators.value)
      // }

      // if (variant.value === 'guest-dashboard') {
      //   setValidators(validators)
      //   return
      // }

      // await fetch(
      //   'DASHBOARD_VALIDATOR_MANAGEMENT_DELETE',
      //   {
      //     body: JSON.stringify({ validators }),
      //     method: 'POST',
      //   },
      //   { dashboardKey: id.value },
      // )
    },
    props: {
      modal: true,
    },
  })
}

// const overviewTest = useFetchedData(`/api/bff/validator-dashboards/${validatorListEncoded.value}/validators`)

const overview = useFetchedData('dashboardOverview')

const latestEffectiveBalance = computed(() => {
  return overview.value?.balances.effective_latest
})

const { productFree } = useProduct()

const effectiveBalanceLimitPerDashboard = computed(() => {
  const effectiveBalanceLimitFreeProduct = productFree.value?.premium_perks.effective_balance_per_dashboard

  return variant.value === 'guest-dashboard'
    ? effectiveBalanceLimitFreeProduct
    : premium_perks.value?.effective_balance_per_dashboard
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
const emit = defineEmits<{
  (e: 'change-validators', value: string[]): void,
}>()

const { $api } = useNuxtApp()
const query = useDefaultQuery()

const handleSubmit = async (item: InternalPostSearchResponse['data'][number] | undefined) => {
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
  if (!variant.value || variant.value === 'guest-dashboard') {
    if (item.type === 'validator' || item.type === 'validator_list') {
      const newValidators = new Set(validatorIds.value)

      if ('index' in item.value) {
        newValidators.add(`${item.value.index}`)
      }
      if ('validators' in item.value) {
        item.value.validators.forEach((validator) => {
          newValidators.add(`${validator}`)
        })
      }
      const validatorListEncoded = encodeBase64Url([ ...newValidators ].join(','))
      // console.log('👉', validatorListEncoded, newValidators)
      try {
        const response = await $api(`/api/bff/validator-dashboards/${validatorListEncoded}/validators`)
        validators.value = response
        // emit('change-validators', [ ...newValidators ])
        setValidators([ ...newValidators ])
        resetInput()
      }
      catch (error) {
        if (error instanceof FetchError)
          if (error.statusMessage === ERROR_CODE.EFFECTIVE_BALANCE_EXCEEDS_LIMIT) {
            dialog.open(BcPremiumModal, {
              data: {
                description: $t('dashboard.validator.management.validators_limit_exceeded'),
              },
            })
          }
      }
      // await $api(`/api/bff/validator-dashboards/${validatorListEncoded}/validators`)
      //   .then((response) => {
      //     setValidators([ ...newValidators ])
      //     fetchedDataValidators.value = response
      //     resetInput()
      //   })
      //   .catch((error) => {
      //     if (error.statusMessage === ERROR_CODE.EFFECTIVE_BALANCE_EXCEEDS_LIMIT) {
      //       dialog.open(BcPremiumModal, {
      //         data: {
      //           description: $t('dashboard.validator.management.validators_limit_exceeded'),
      //         },
      //       })
      //     }
      //   })
    }
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
    .then(() => { resetInput() })
    .catch((error) => {
      if (error.statusCode === 403) {
        dialog.open(BcPremiumModal, {
          data: {
            description: $t('dashboard.validator.management.validators_limit_exceeded'),
          },
        })
      }
    })
}
const inputValidator = ref('')
const {
  data,
  status,
} = useFetch(`/api/bff/validator-dashboards/${key.value}/validators`, {
  immediate: !!key.value,
  key: 'validators',
  query,
})

// const { data }
//  = useFetch(`/api/bff/validator-dashboards/${key.value}/validators`, {
//    method: 'post',
//  })

// const test = ref(encodeBase64Url('1,2'))
// const onClick = () => {
//   console.log('Button clicked!')
//   test.value = encodeBase64Url('1,2,3')
// }
// const {
//   data,
//   status,
// } = useApi(() => `/api/bff/validator-dashboards/${test.value}/validators`, {
//   immediate: validatorListEncoded.value.length > 0,
//   // key: 'validators',
//   query,
// })
</script>

<template>
  <BcDialog
    v-model="visible"
    :header="$t('dashboard.validator.management.title')"
    :close-on-escape="hasNoOpenDialogs"
    class="validator-managment-modal-container"
    @update:visible="(visible: boolean) => !visible && resetData()"
  >
    <template
      v-if="!size.showWithdrawalCredentials"
      #header
    >
      <span />
    </template>
    <BcTableControl
      v-model:search="query.search"
      :search-placeholder="
        $t(
          variant === 'guest-dashboard'
            ? 'dashboard.validator.summary.search_placeholder_public'
            : 'dashboard.validator.summary.search_placeholder',
        )
      "
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
            v-model="query.group_id"
            :include-all="true"
            class="small group-selection"
          />
          <DashboardValidatorManagementModalSearch
            v-model="inputValidator"
            class="search-bar"
            :has-premium-perk-bulk-adding
            :has-reached-limit
            :is-guest-dashboard="variant === 'guest-dashboard'"
            @submit="handleSubmit"
          />
        </div>
      </template>
      <template #table>
        <ClientOnly fallback-tag="span">
          <BcTable
            v-model:selection="selection"
            v-model:query="query"
            :is-loading="status ==='pending'"
            :data
            data-key="public_key"
            :expandable="size.expandable"
            selection-mode="multiple"
            class="management-table"
            :cursor
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
                  v-if="selection?.length"
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
                @click="onClose"
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
