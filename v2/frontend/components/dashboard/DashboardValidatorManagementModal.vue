<script lang="ts" setup>
import {
  faEdit, faTrash,
} from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import type { DataTableSortEvent } from 'primevue/datatable'
import { warn } from 'vue'
import {
  BcDialogConfirm,
  BcPremiumModal,
  DashboardGroupSelectionDialog,
} from '#components'
import { useValidatorDashboardOverviewStore } from '~/stores/dashboard/useValidatorDashboardOverviewStore'
import type {
  GetValidatorDashboardValidatorsResponse,
  VDBManageValidatorsTableRow,
  VDBPostValidatorsData,
} from '~/types/api/validator_dashboard'
import type { Cursor } from '~/types/datatable'
import type { NumberOrString } from '~/types/value'

import {
  API_PATH, type PathValues,
} from '~/types/customFetch'
import type { InternalPostSearchResponse } from '~/types/api/search'
// import type { InternalPostSearchResponse } from '~/types/api/search'

const { t: $t } = useTranslation()
const { fetch } = useCustomFetch()

const { width } = useWindowSize()

const dialog = useDialog()

const visible = defineModel<boolean>()

const {
  overview, refreshOverview,
} = useValidatorDashboardOverviewStore()

const cursor = ref<Cursor>()
const pageSize = ref<number>(25)
const selectedGroup = ref<number>(-1)
const {
  addEntities,
  dashboardKey,
  isPublic: isPublicDashboard,
  removeEntities,
}
  = useDashboardKey()
const {
  user,
} = useUserStore()

const initialQuery = {
  limit: pageSize.value,
  sort: 'index:asc',
}

const {
  bounce: setQuery,
  instant: instantQuery,
  temp: tempQuery,
  value: query,
} = useDebounceValue<PathValues | undefined>(initialQuery, 500)

const data = ref<GetValidatorDashboardValidatorsResponse | undefined>()
const selected = ref<VDBManageValidatorsTableRow[]>()
const hasNoOpenDialogs = ref(true)

type ValidatorUpdateBody = {
  deposit_address?: string,
  graffiti?: string,
  group_id?: number,
  validators?: number[],
  withdrawal_address?: string,
}

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
  data.value = undefined
  selected.value = []
  selectedGroup.value = -1
  cursor.value = undefined
  instantQuery(initialQuery)
}

const onClose = () => {
  resetData()
  visible.value = false
}

const mapIndexOrPubKey = (
  validators?: VDBManageValidatorsTableRow[],
) => {
  return [ ...new Set(validators?.map(
    validator => validator.index ?? validator.public_key)) ]
}

const changeGroup = async (body: ValidatorUpdateBody, groupId?: number) => {
  if (
    !body.validators?.length
    && !body.deposit_address
    && !body.graffiti
    && !body.withdrawal_address
  ) {
    warn('no validators selected to change group')
    return
  }
  body.group_id = groupId && groupId !== -1 ? groupId : 0

  await fetch<VDBPostValidatorsData>(
    API_PATH.DASHBOARD_VALIDATOR_MANAGEMENT,
    {
      body,
      method: 'POST',
    },
    { dashboardKey: dashboardKey.value },
  )

  loadData()
  refreshOverview(dashboardKey.value)
}

const removeValidators = async (validators?: NumberOrString[]) => {
  if (!validators?.length) {
    warn('no validators selected to change group')
    return
  }
  if (isPublicDashboard.value) {
    removeEntities(validators.map(v => v.toString()))
    return
  }

  await fetch(
    API_PATH.DASHBOARD_VALIDATOR_MANAGEMENT_DELETE,
    {
      body: JSON.stringify({ validators }),
      method: 'POST',
    },
    { dashboardKey: dashboardKey.value },
  )

  loadData()
  refreshOverview(dashboardKey.value)
}

const { premium_perks } = useUserStore()

const editSelected = () => {
  hasNoOpenDialogs.value = false
  dialog.open(DashboardGroupSelectionDialog, {
    data: {
      groupId: selected.value?.[0]?.group_id ?? undefined,
      selectedValidators: selected.value?.length,
      totalValidatorsValidators: totalValidators?.value,
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

const onSort = (sort: DataTableSortEvent) => {
  setQuery(setQuerySort(sort, query?.value))
}

const setCursor = (value: Cursor) => {
  cursor.value = value
  setQuery(setQueryCursor(value, query?.value))
}

const setPageSize = (value: number) => {
  pageSize.value = value
  setQuery(setQueryPageSize(value, query?.value))
}

const setSearch = (value?: string) => {
  setQuery(setQuerySearch(value, query?.value))
}

watch(selectedGroup, (value) => {
  setQuery({
    ...query?.value,
    group_id: value,
  })
})

const loadData = async () => {
  if (dashboardKey.value) {
    const testQ = JSON.stringify(query.value)
    const result = await fetch<GetValidatorDashboardValidatorsResponse>(
      API_PATH.DASHBOARD_VALIDATOR_MANAGEMENT,
      undefined,
      { dashboardKey: dashboardKey.value },
      query.value,
    )

    // Make sure that during loading the query did not change
    if (testQ === JSON.stringify(query.value)) {
      data.value = result
      selected.value = []
    }
  }
  else {
    data.value = {
      data: [],
      paging: {},
    }
  }
}

watch(
  () => [
    dashboardKey.value,
    visible.value,
    query.value,
  ],
  () => {
    if (visible.value) {
      loadData()
    }
  },
  { immediate: true },
)

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
      response?.data && removeValidators(list)
    },
  })
}

const totalValidators = computed(() => addUpValues(overview.value?.validators))

const maxValidatorsPerDashboard = computed(() =>
  isPublicDashboard.value || !user.value?.premium_perks?.validators_per_dashboard
    ? 20
    : user.value.premium_perks.validators_per_dashboard,
)

const premiumLimit = computed(
  () => totalValidators.value >= maxValidatorsPerDashboard.value,
)
// const hasTooManyValidators = computed(() => totalValidators.value + 1 > maxValidatorsPerDashboard.value)
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
  if (
    totalValidators.value + 1 > maxValidatorsPerDashboard.value
    || (type === 'validator_list' && totalValidators.value + value.validators.length > maxValidatorsPerDashboard.value)
  ) {
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
  if (isPublicDashboard.value) {
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
  resetInput()
}
const inputValidator = ref('')
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
      :search-placeholder="
        $t(
          isPublicDashboard
            ? 'dashboard.validator.summary.search_placeholder_public'
            : 'dashboard.validator.summary.search_placeholder',
        )
      "
      @set-search="setSearch"
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
          />
          <DashboardValidatorManagementModalSearch
            v-model="inputValidator"
            class="search-bar"
            :has-premium-perk-bulk-adding
            :total-validators
            :max-validators-per-dashboard
            :is-public-dashboard
            @submit="handleSubmit"
          />
        </div>
      </template>
      <template #table>
        <ClientOnly fallback-tag="span">
          <BcTable
            v-model:selection="selected"
            :data
            data-key="public_key"
            :expandable="size.expandable"
            selection-mode="multiple"
            class="management-table"
            :cursor
            :page-size
            :selected-sort="tempQuery?.sort as string"
            @set-cursor="setCursor"
            @sort="onSort"
            @set-page-size="setPageSize"
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
                  <BcFormatValue :value="slotProps.data.balance" />
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
                  :position="slotProps.data.queue_position"
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
                  <FontAwesomeIcon
                    class="edit-icon"
                    :icon="faEdit"
                  />
                </Button>
              </template>
              <template #body="slotProps">
                <div class="action-col">
                  <FontAwesomeIcon
                    :icon="faTrash"
                    class="link"
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
                  <BcFormatValue :value="slotProps.data.balance" />
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
                    :position="slotProps.data.queue_position"
                  />
                </div>
                <div class="info">
                  <div class="label">
                    {{ $t("dashboard.validator.col.withdrawal_credential") }}
                  </div>
                  <BcFormatHash
                    :hash="slotProps.data.withdrawal_credential"
                    type="withdrawal_credentials"
                  />
                </div>
              </div>
            </template>

            <template #bc-table-footer-left>
              <div
                v-if="maxValidatorsPerDashboard"
                class="left"
              >
                <div
                  class="labels"
                  :class="{ premiumLimit }"
                >
                  <span>
                    <BcFormatNumber
                      :value="totalValidators"
                      default="0"
                    /> /
                    <BcFormatNumber
                      :value="maxValidatorsPerDashboard"
                      default="0"
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

    &.premiumLimit {
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

.public-key {
  width: 134px;
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
