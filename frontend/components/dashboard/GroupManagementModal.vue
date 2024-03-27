<script lang="ts" setup>
import {
  faAdd,
  faTrash
} from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { orderBy } from 'lodash-es'
import type { DataTableSortEvent } from 'primevue/datatable'
import { warn } from 'vue'
import { BcDialogConfirm } from '#components'
import { useValidatorDashboardOverviewStore } from '~/stores/dashboard/useValidatorDashboardOverviewStore'
import type { ApiPagingResponse } from '~/types/api/common'
import type { VDBOverviewGroup } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import type { Cursor } from '~/types/datatable'
import { getSortOrder } from '~/utils/table'

const { t: $t } = useI18n()
const { fetch } = useCustomFetch()
const dialog = useDialog()

interface Props {
  dashboardKey: DashboardKey;
}
const props = defineProps<Props>()

const { width, isMobile } = useWindowSize()

const visible = defineModel<boolean>()

const overviewStore = useValidatorDashboardOverviewStore()
const { getOverview } = overviewStore
const { overview } = storeToRefs(overviewStore)

const dashboardStore = useUserDashboardStore()
const { dashboards } = storeToRefs(dashboardStore)

const cursor = ref<Cursor>(0)
const pageSize = ref<number>(5)
const newGroupName = ref<string>('')
const search = ref<string>()
const sortField = ref<string>()
const sortOrder = ref<number | null>()

const data = computed<ApiPagingResponse<VDBOverviewGroup>>(() => {
  let groups = (overview.value?.groups ?? [])
  if (search.value?.length) {
    const s = search.value.toLowerCase()
    groups = groups.filter(g => g.name.toLowerCase().includes(s) || parseInt(s) === g.id)
  }
  if (sortField.value?.length && sortOrder.value) {
    groups = orderBy(groups, sortField.value, getSortOrder())
  }
  const totalCount = groups.length
  return {
    paging: {
      total_count: totalCount
    },
    data: groups.slice(cursor.value as number, pageSize.value)
  }
})

const size = computed(() => {
  return {
    showSubTitle: width.value >= 760
  }
})

const resetData = () => {
  search.value = ''
  newGroupName.value = ''
  cursor.value = 0
}

const onClose = () => {
  visible.value = false
  resetData()
}

const addGroup = async () => {
  if (!newGroupName.value?.length) {
    return
  }
  await fetch(API_PATH.DASHBOARD_VALIDATOR_GROUPS, { method: 'POST', body: { name: newGroupName.value } }, { dashboardKey: props.dashboardKey })
  await getOverview(props.dashboardKey)
  newGroupName.value = ''
}

const editGroup = (row: VDBOverviewGroup, newName?: string) => {
  // TODO: Implement group renaming once the backend supports it.
  warn(`Edit group ${row.name} [${row.id}] -> ${newName}`)
}

const removeGroupConfirmed = async (row: VDBOverviewGroup) => {
  await fetch(API_PATH.DASHBOARD_VALIDATOR_GROUP_DELETE, undefined, { dashboardKey: props.dashboardKey, groupId: row.id })
  getOverview(props.dashboardKey)
}

const removeGroup = (row: VDBOverviewGroup) => {
  dialog.open(BcDialogConfirm, {
    props: {
      header: $t('dashboard.validator.group_management.remove_title')
    },
    onClose: response => response?.data && removeGroupConfirmed(row),
    data: {
      question: $t('dashboard.validator.group_management.remove_text', { group: row.name })
    }
  })
}

const onSort = (sort: DataTableSortEvent) => {
  sortField.value = sort.sortField as string
  sortOrder.value = sort.sortOrder
}

const setCursor = (value: Cursor) => {
  cursor.value = value
}

const setPageSize = (value: number) => {
  pageSize.value = value
}

const setSearch = (value?: string) => {
  search.value = value
}

const dashboardName = computed(() => {
  return dashboards.value?.validator_dashboards?.find(d => `${d.id}` === props.dashboardKey)?.name || props.dashboardKey
})

// TODO: once we have a user management we need to check how to get the real premium limit
const MaxGroupsPerDashboard = 40
const premiumLimit = computed(() => (data.value?.paging?.total_count ?? 0) >= MaxGroupsPerDashboard)

</script>

<template>
  <BcDialog
    v-model="visible"
    :header="$t('dashboard.validator.group_management.title')"
    class="validator-group-managment-modal-container"
    @update:visible="(visible: boolean)=>!visible && resetData()"
  >
    <template v-if="!size.showSubTitle" #header>
      <span />
    </template>
    <BcTableControl
      :search-placeholder="$t('dashboard.validator.group_management.search_placeholder')"
      @set-search="setSearch"
    >
      <template #header-left>
        <span v-if="size.showSubTitle"> {{ $t('dashboard.validator.group_management.sub_title', { dashboardName })
        }}</span>
        <span v-else class="small-title">{{ $t('dashboard.validator.group_management.title') }}</span>
      </template>
      <template #bc-table-sub-header>
        <div class="add-row">
          <InputText
            v-model="newGroupName"
            class="search-input"
            :disabled="premiumLimit"
            maxlength="20"
            :placeholder="$t('dashboard.validator.group_management.new_group_placeholder')"
            @keypress.enter="addGroup"
          />
          <Button style="display: inline;" :disabled="!newGroupName.length || premiumLimit" @click="addGroup">
            <FontAwesomeIcon :icon="faAdd" />
          </Button>
        </div>
      </template>
      <template #table>
        <ClientOnly fallback-tag="span">
          <BcTable
            :data="data"
            class="management-table"
            :cursor="cursor"
            :page-size="pageSize"
            @set-cursor="setCursor"
            @sort="onSort"
            @set-page-size="setPageSize"
          >
            <Column field="name" class="edit-group" :sortable="true" :header="$t('dashboard.validator.group_management.col.name')">
              <template #body="slotProps">
                <!-- TODO: wait for the backend to implement group renaming the activate this input and finish the logic -->
                <BcInputLabel
                  class="edit-group truncate-text"
                  :value="slotProps.data.name"
                  :default="slotProps.data.id === 0 ? $t('common.default') : ''"
                  :can-be-empty="slotProps.data.id === 0"
                  :disabled="false"
                  :maxlength="20"
                  @set-value="(name: string) => editGroup(slotProps.data, name)"
                />
              </template>
            </Column>
            <Column field="id" :sortable="!isMobile" :header="$t('dashboard.validator.group_management.col.id')" />

            <Column field="count" :sortable="!isMobile" :header="$t('dashboard.validator.group_management.col.count')">
              <template #body="slotProps">
                <BcFormatNumber :value="slotProps.data.count" default="0" />
              </template>
            </Column>
            <Column field="action">
              <template #body="slotProps">
                <div class="action-col">
                  <FontAwesomeIcon
                    v-if="slotProps.data.id"
                    :icon="faTrash"
                    class="link"
                    @click="removeGroup(slotProps.data)"
                  />
                </div>
              </template>
            </Column>
          </BcTable>
        </ClientOnly>
      </template>
    </BcTableControl>
    <template #footer>
      <div class="footer">
        <div class="left">
          <div v-if="MaxGroupsPerDashboard" class="labels" :class="{premiumLimit}">
            <span>
              <BcFormatNumber :value="data.paging.total_count" default="0" />/
              <BcFormatNumber :value="MaxGroupsPerDashboard" />
            </span>
            <span>{{ $t('dashboard.validator.group_management.groups_added') }}</span>
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

:global(.validator-group-managment-modal-container) {
  width: 960px;
  height: 800px;

}

:global(.validator-group-managment-modal-container .p-dialog-content) {
  display: flex;
  flex-direction: column;
  flex-grow: 1;
}

:global(.validator-group-managment-modal-container .bc-table-header) {
  height: unset;
  padding: var(--padding) 0;
  @include fonts.subtitle_text;
}

:global(.validator-group-managment-modal-container .bc-table-header .side:first-child) {
  display: contents;
}
:global(.validator-group-managment-modal-container .bc-pageinator .left-info) {
  padding-left: var(--padding-large);
}

:global(.validator-group-managment-modal-container .edit-group ){
  max-width: 201px;
  width: 201px;
}

.edit-group {
  max-width: 180px;
}

.small-title {
  @include utils.truncate-text;
  @include fonts.big_text;
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

  .search-input {
    flex-shrink: 1;
    flex-grow: 1;
    width: 50px;
    border-top-right-radius: 0;
    border-bottom-right-radius: 0;
    height: 100%;
  }

  button {
    border-top-left-radius: 0;
    border-bottom-left-radius: 0;
    padding: var(--padding-small) 8px;
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

      &.premiumLimit {
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
  width: 100%;
  display: flex;
  justify-content: flex-end;
}

@media (max-width: 560px) {

  .edit-group {
    max-width: 100px;
  }

  .action-col {
    width: 33px;
  }
}
</style>
