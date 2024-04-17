<script lang="ts" setup>
import {
  faEdit,
  faTrash
} from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import type { DataTableSortEvent } from 'primevue/datatable'
import { warn } from 'vue'
import { uniq } from 'lodash-es'
import { BcDialogConfirm, DashboardGroupSelectionDialog } from '#components'
import { useValidatorDashboardOverviewStore } from '~/stores/dashboard/useValidatorDashboardOverviewStore'
import type { InternalGetValidatorDashboardValidatorsResponse, VDBManageValidatorsTableRow, VDBPostValidatorsData } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import type { Cursor } from '~/types/datatable'
import type { NumberOrString } from '~/types/value'
import { type SearchBar, SearchbarStyle, SearchbarPurpose, ResultType, type ResultSuggestion, pickHighestPriorityAmongBestMatchings } from '~/types/searchbar'
import { ChainIDs } from '~/types/networks'

const { t: $t } = useI18n()
const { fetch } = useCustomFetch()

interface Props {
  dashboardKey: DashboardKey;
}
const props = defineProps<Props>()

const { width } = useWindowSize()

const dialog = useDialog()

const visible = defineModel<boolean>()

const { overview, refreshOverview } = useValidatorDashboardOverviewStore()

const cursor = ref<Cursor>()
const pageSize = ref<number>(5)
const selectedGroup = ref<number>(-1)
const selectedValidator = ref<string>('')

const { value: query, bounce: setQuery } = useDebounceValue<PathValues | undefined>({ limit: pageSize.value }, 500)

const data = ref<InternalGetValidatorDashboardValidatorsResponse | undefined>()
const selected = ref<VDBManageValidatorsTableRow[]>()
const searchBar = ref<SearchBar>()
const hasNoOpenDialogs = ref(true)

const size = computed(() => {
  return {
    expandable: width.value < 960,
    showBalance: width.value >= 960,
    showGroup: width.value >= 760,
    showWithdrawalCredentials: width.value >= 560
  }
})

const resetData = () => {
  data.value = undefined
  selected.value = []
  selectedGroup.value = -1
  cursor.value = undefined
}

const onClose = () => {
  resetData()
  visible.value = false
}

const mapIndexOrPubKey = (validators?: VDBManageValidatorsTableRow[]):NumberOrString[] => {
  return uniq(validators?.map(vali => vali.index?.toString() ?? vali.public_key) ?? [])
}

const changeGroup = async (validators?: NumberOrString[], groupId?: number) => {
  if (!validators?.length) {
    warn('no validators selected to change group')
    return
  }
  const targetGroupId = groupId !== -1 ? groupId?.toString() : '0'

  await fetch< VDBPostValidatorsData >(API_PATH.DASHBOARD_VALIDATOR_MANAGEMENT, { method: 'POST', body: { validators, group_id: targetGroupId } }, { dashboardKey: props.dashboardKey })

  loadData()
  refreshOverview(props.dashboardKey)
}

const removeValidators = async (validators?: NumberOrString[]) => {
  if (!validators?.length) {
    warn('no validators selected to change group')
    return
  }

  await fetch(API_PATH.DASHBOARD_VALIDATOR_MANAGEMENT, { method: 'DELETE', query: { validators: validators.join(',') } }, { dashboardKey: props.dashboardKey })

  loadData()
  refreshOverview(props.dashboardKey)
}

const addValidator = (result : ResultSuggestion) => {
  switch (result.type) {
    case ResultType.ValidatorsByIndex : // `result.queryParam` contains the index of the validator
    case ResultType.ValidatorsByPubkey : // `result.queryParam` contains the pubkey of the validator
      selectedValidator.value = result.queryParam
      break
    // The following types can correspond to several validators. The search bar doesn't know the list of indices and pubkeys :
    case ResultType.ValidatorsByDepositAddress : // `result.queryParam` contains the address that was used to deposit
    case ResultType.ValidatorsByDepositEnsName : // `result.queryParam` contains the ENS name that was used to deposit
    case ResultType.ValidatorsByWithdrawalCredential : // `result.queryParam` contains the withdrawal credential
    case ResultType.ValidatorsByWithdrawalAddress : // `result.queryParam` contains the withdrawal address
    case ResultType.ValidatorsByWithdrawalEnsName : // `result.queryParam` contains the ENS name of the withdrawal address
    case ResultType.ValidatorsByGraffiti : // `result.queryParam` contains the graffiti used to sign blocks
      selectedValidator.value = result.queryParam // TODO: maybe handle these cases differently? (because `result.queryParam` identifies a list of validators instead of a single index/pubkey)
      break
    default :
      return
  }
  // When the result is a batch of validators, result.count is the size of the batch.

  changeGroup([selectedValidator.value], selectedGroup.value)

  // The following method hides the result in the drop-down, so the user can easily identify which validators he can still add:
  searchBar.value!.hideResult(result)
  // You do not have to call it here, you can do it later, for example after getting confirmation that the validator is added into the database.

  // Because of props `:keep-dropdown-open="true"` in the template, the dropdown does not close when the user selects a validator.
  // If it happens that you want to close the dropdown, you can call this method:
  // searchBar.value!.closeDropdown()
  // Or, if you are sure that the dropdown should always be closed when the user selects something, simply remove `:keep-dropdown-open="true"`.
}

const editSelected = () => {
  hasNoOpenDialogs.value = false
  dialog.open(DashboardGroupSelectionDialog, {
    onClose: (response) => {
      hasNoOpenDialogs.value = true
      if (response?.data !== undefined) {
        changeGroup(mapIndexOrPubKey(selected.value), response?.data)
      }
    },
    data: {
      groupId: selected.value?.[0]?.group_id ?? undefined,
      selectedValidators: selected.value?.length,
      totalValidators: total?.value
    }
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
  setQuery({ ...query?.value, group_id: value })
})

const loadData = async () => {
  if (props.dashboardKey) {
    const testQ = JSON.stringify(query.value)
    const result = await fetch<InternalGetValidatorDashboardValidatorsResponse>(API_PATH.DASHBOARD_VALIDATOR_MANAGEMENT, undefined, { dashboardKey: props.dashboardKey }, query.value)

    // Make sure that during loading the query did not change
    if (testQ === JSON.stringify(query.value)) {
      data.value = result
      selected.value = []
    }
  }
}

watch(() => [props.dashboardKey, visible.value, query.value], () => {
  if (visible.value) {
    loadData()
  }
}, { immediate: true })

const switchValidatorGroup = (row: VDBManageValidatorsTableRow, group: number) => {
  changeGroup(mapIndexOrPubKey([row].concat(selected.value ?? [])), group)
}

const removeRow = (row: VDBManageValidatorsTableRow) => {
  const list = mapIndexOrPubKey([row].concat(selected.value ?? []))
  if (!list?.length) {
    warn('no validator to remove')
  }

  hasNoOpenDialogs.value = false
  dialog.open(BcDialogConfirm, {
    onClose: (response) => {
      hasNoOpenDialogs.value = true
      response?.data && removeValidators(list)
    },
    data: {
      title: $t('dashboard.validator.management.remove_title'),
      question: $t('dashboard.validator.management.remove_text', { validator: list[0] }, list.length)
    }
  })
}

const total = computed(() => addUpValues(overview.value?.validators))

// TODO: get this value from the backend based on the logged in user
const MaxValidatorsPerDashboard = 1000

const premiumLimit = computed(() => (data.value?.paging?.total_count ?? 0) >= MaxValidatorsPerDashboard)

</script>

<template>
  <BcDialog
    v-model="visible"
    :header="$t('dashboard.validator.management.title')"
    :close-on-escape="hasNoOpenDialogs"
    class="validator-managment-modal-container"
    @update:visible="(visible: boolean)=>!visible && resetData()"
  >
    <template v-if="!size.showWithdrawalCredentials" #header>
      <span />
    </template>
    <BcTableControl :search-placeholder="$t('dashboard.validator.summary.search_placeholder')" @set-search="setSearch">
      <template #header-left>
        <span v-if="size.showWithdrawalCredentials"> {{ $t('dashboard.validator.management.sub_title') }}</span>
        <span v-else class="small-title">{{ $t('dashboard.validator.manage-validators') }}</span>
      </template>
      <template #bc-table-sub-header>
        <div class="add-row">
          <DashboardGroupSelection v-model="selectedGroup" :include-all="true" class="small group-selection" />
          <!-- TODO: below, replace "[ChainIDs.Ethereum]" with a variable containing the array of chain id(s) that the validators should belong to -->
          <BcSearchbarMain
            ref="searchBar"
            :bar-style="SearchbarStyle.Embedded"
            :bar-purpose="SearchbarPurpose.ValidatorAddition"
            :only-networks="[ChainIDs.Ethereum]"
            :pick-by-default="pickHighestPriorityAmongBestMatchings"
            :keep-dropdown-open="true"
            class="search-bar"
            @go="addValidator"
          />
        </div>
      </template>
      <template #table>
        <ClientOnly fallback-tag="span">
          <BcTable
            v-model:selection="selected"
            :data="data"
            data-key="public_key"
            :expandable="size.expandable"
            selection-mode="multiple"
            class="management-table"
            :cursor="cursor"
            :page-size="pageSize"
            @set-cursor="setCursor"
            @sort="onSort"
            @set-page-size="setPageSize"
          >
            <Column field="index" :sortable="true" :header="$t('dashboard.validator.col.index')" />

            <Column field="public_key" :sortable="!size.expandable" :header="$t('dashboard.validator.col.public_key')">
              <template #body="slotProps">
                <BcFormatHash :hash="slotProps.data.public_key" type="public_key" class="public-key" />
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
                  @set-group="(id: number) => switchValidatorGroup(slotProps.data, id)"
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
                <BcFormatValue :value="slotProps.data.balance" />
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
                <BcFormatHash :hash="slotProps.data.withdrawal_credential" type="withdrawal_credentials" />
              </template>
            </Column>
            <Column field="action">
              <template #header>
                <Button v-show="selected?.length" class="edit-button" @click.stop.prevent="editSelected()">
                  <span class="edit-label">{{ $t('common.edit') }}</span>
                  <FontAwesomeIcon class="edit-icon" :icon="faEdit" />
                </Button>
              </template>
              <template #body="slotProps">
                <div class="action-col">
                  <FontAwesomeIcon :icon="faTrash" class="link" @click="removeRow(slotProps.data)" />
                </div>
              </template>
            </Column>
            <template #expansion="slotProps">
              <div class="expansion">
                <div class="info">
                  <div class="label">
                    {{ $t('dashboard.validator.col.balance') }}
                  </div>
                  <BcFormatValue :value="slotProps.data.balance" />
                </div>
                <div class="info">
                  <div class="label">
                    {{ $t('dashboard.validator.col.group') }}
                  </div>
                  <DashboardGroupSelection
                    v-model="slotProps.data.group_id"
                    class="small"
                    @set-group="(id: number) => switchValidatorGroup(slotProps.data, id)"
                  />
                </div>
                <div class="info">
                  <div class="label">
                    {{ $t('dashboard.validator.col.withdrawal_credential') }}
                  </div>
                  <BcFormatHash :hash="slotProps.data.withdrawal_credential" type="withdrawal_credentials" />
                </div>
              </div>
            </template>
          </BcTable>
        </ClientOnly>
      </template>
    </BcTableControl>
    <template #footer>
      <div class="footer">
        <div v-if="MaxValidatorsPerDashboard" class="left">
          <div class="labels" :class="{premiumLimit}">
            <span><BcFormatNumber :value="data?.paging?.total_count" default="0" />/<BcFormatNumber :value="MaxValidatorsPerDashboard" default="0" /></span>
            <span>{{ $t('dashboard.validator.management.validators_added') }}</span>
          </div>
          <BcPremiumGem />
        </div>
        <Button :label="$t('navigation.done')" @click="onClose" />
      </div>
    </template>
  </BcDialog>
</template>

<style lang="scss" scoped>
@use '~/assets/css/main.scss';
@use '~/assets/css/utils.scss';
@use '~/assets/css/fonts.scss';

:global(.validator-managment-modal-container) {
  width: 960px;
  height: 800px;

}

:global(.validator-managment-modal-container .p-dialog-content) {
  display: flex;
  flex-direction: column;
  flex-grow: 1;
}

:global(.validator-managment-modal-container .bc-table-header) {
  height: unset;
  padding: var(--padding) 0;
  @include fonts.subtitle_text;
}

:global(.validator-managment-modal-container .bc-table-header .side:first-child) {
  display: contents;
}

:global(.validator-managment-modal-container .bc-pageinator .left-info) {
  padding-left: var(--padding-large);
}

.small-title {
  @include utils.truncate-text;
  @include fonts.big_text;
}

.group-selection {
  width: 160px;
}

.management-table {
  @include main.container;
  flex-grow: 1;
  display: flex;
  flex-direction: column;
  overflow-y: hidden;

  :deep(.p-datatable-wrapper) {
    flex-grow: 1;
  }
}

.add-row {
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

.footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: var(--padding-large);
  gap: var(--padding);

  .left {
    display: flex;
    align-items: center;
    gap: var(--padding-small);

    .labels {
      display: flex;
      gap: var(--padding-small);
      &.premiumLimit{
        color: var(--negative-color);
      }

      @media (max-width: 959px) {
        flex-direction: column;
      }
    }

    .gem {
      color: var(--primary-color);
    }
  }
}

.public-key {
  width: 134px;
}

.edit-icon {
  margin-left: var(--padding-small);
}

.action-col {
  width: 94px;
  display: flex;
  justify-content: flex-end;
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

  .action-col {
    width: 33px;
  }

  :deep(.status-col) {
    .p-column-title {
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
      max-width: 20px;
    }

  }
}

.expansion {
  @include main.container;
  padding: var(--padding);
  display: flex;
  flex-direction: column;
  gap: var(--padding);

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
