<script lang="ts" setup>
import {
  faAdd,
  faEdit,
  faTrash
} from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import type { DataTableSortEvent } from 'primevue/datatable'
import { warn } from 'vue'
import { uniq } from 'lodash-es'
import { BcDialogConfirm, DashboardGroupSelectionDialog } from '#components'
import { useValidatorDashboardOverviewStore } from '~/stores/dashboard/useValidatorDashboardOverviewStore'
import type { InternalGetValidatorDashboardValidatorsResponse, VDBManageValidatorsTableRow } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import type { Cursor } from '~/types/datatable'
import type { NumberOrString } from '~/types/value'

const { t: $t } = useI18n()
const { fetch } = useCustomFetch()

interface Props {
  dashboardKey: DashboardKey;
}
const props = defineProps<Props>()

const { width } = useWindowSize()

const dialog = useDialog()

const visible = defineModel<boolean>()

const overviewStore = useValidatorDashboardOverviewStore()
const { getOverview } = overviewStore
const { overview } = storeToRefs(overviewStore)

const { value: query, bounce: setQuery } = useDebounceValue<PathValues | undefined>(undefined, 500)

const cursor = ref<Cursor>()
const pageSize = ref<number>(5)
const selectedGroup = ref<number>(-1)
const selectedValidator = ref<string>('')

const data = ref<InternalGetValidatorDashboardValidatorsResponse | undefined>()
const selected = ref<VDBManageValidatorsTableRow[]>()

const size = computed(() => {
  return {
    expandable: width.value < 960,
    showBalance: width.value >= 960,
    showGroup: width.value >= 760,
    showWithdrawalCredentials: width.value >= 560
  }
})

const onClose = () => {
  visible.value = false
}

const mapIndexOrPubKey = (validators?: VDBManageValidatorsTableRow[], notInGroup?: number):NumberOrString[] => {
  if (notInGroup !== undefined) {
    validators = validators?.filter(v => v.group_id !== notInGroup)
  }

  return uniq(validators?.map(vali => vali.index?.toString() ?? vali.public_key) ?? [])
}

const changeGroup = async (validators?: NumberOrString[], groupId?: number) => {
  if (!validators?.length) {
    warn('no validators selected to change group')
    return
  }
  const targetGroupId = groupId !== -1 ? groupId?.toString() : '0'

  await fetch<any>(API_PATH.DASHBOARD_VALIDATOR_MANAGEMENT, { method: 'POST', body: { validators, group_id: targetGroupId } }, { dashboardKey: props.dashboardKey })
  loadData()
  getOverview(props.dashboardKey)
}

const removeValidators = async (validators?: NumberOrString[]) => {
  if (!validators?.length) {
    warn('no validators selected to change group')
    return
  }

  await fetch<any>(API_PATH.DASHBOARD_VALIDATOR_MANAGEMENT, { method: 'DELETE', body: { validators } }, { dashboardKey: props.dashboardKey })
  loadData()
  getOverview(props.dashboardKey)
}

const addValidator = () => {
  changeGroup([selectedValidator.value], selectedGroup.value)
}

const editSelected = () => {
  dialog.open(DashboardGroupSelectionDialog, {
    onClose: (response) => {
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
    data.value = await fetch<InternalGetValidatorDashboardValidatorsResponse>(API_PATH.DASHBOARD_VALIDATOR_MANAGEMENT, undefined, { dashboardKey: props.dashboardKey }, query.value)
    selected.value = []
  }
}

watch(() => [props.dashboardKey, visible.value, query.value], () => {
  if (visible.value) {
    loadData()
  }
}, { immediate: true })

const switchValidatorGroup = (row: VDBManageValidatorsTableRow, group: number) => {
  changeGroup(mapIndexOrPubKey([row].concat(selected.value ?? []), group), group)
}

const removeRow = (row: VDBManageValidatorsTableRow) => {
  const list = mapIndexOrPubKey([row].concat(selected.value ?? []))
  if (!list?.length) {
    warn('no validator to remove')
  }

  dialog.open(BcDialogConfirm, {
    props: {
      header: $t('dashboard.validator.management.remove_title')
    },
    onClose: response => response?.data && removeValidators(list),
    data: {
      question: $t('dashboard.validator.management.remove_text', { validator: list[0] }, list.length),
      yesLabel: $t('navigation.remove'),
      noLabel: $t('navigation.dismiss')
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
    class="validator-managment-modal-container"
  >
    <template v-if="!size.showWithdrawalCredentials" #header>
      <span class="hdden-title" />
    </template>
    <BcTableControl :search-placeholder="$t('dashboard.validator.summary.search_placeholder')" @set-search="setSearch">
      <template #header-left>
        <span v-if="size.showWithdrawalCredentials"> {{ $t('dashboard.validator.management.sub_title') }}</span>
        <span v-else class="small-title">{{ $t('dashboard.validator.manage-validators') }}</span>
      </template>
      <template #bc-table-sub-header>
        <div class="add-row">
          <DashboardGroupSelection v-model="selectedGroup" :include-all="true" class="small group-selection" />
          <!-- TODO: replace input once Searchbar is finished -->
          <InputText v-model="selectedValidator" class="search-input" placeholder="Placeholder input (will be replaced once the searchbar is finished)" />
          <Button class="p-button-icon-only" style="display: inline;" :disabled="!selectedValidator" @click="addValidator">
            <FontAwesomeIcon :icon="faAdd" />
          </Button>
          <!-- end of temp -->
        </div>
      </template>
      <template #table>
        <ClientOnly fallback-tag="span">
          <BcTable
            v-model:selection="selected"
            :data="data"
            data-key="group_id"
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
            <span>{{ data?.paging?.total_count ?? 0 }}/{{ MaxValidatorsPerDashboard }}</span>
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
}

:global(.validator-managment-modal-container .bc-table-header .side:first-child) {
  display: contents;
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

  .search-input {
    flex-shrink: 1;
    flex-grow: 1;
    width: 50px;
  }
}

.footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: var(--padding);
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
