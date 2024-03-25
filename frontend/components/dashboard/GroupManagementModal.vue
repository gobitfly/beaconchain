<script lang="ts" setup>
import {
  faAdd,
  faTrash
} from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { orderBy } from 'lodash-es'
import type { DataTableSortEvent } from 'primevue/datatable'
import { warn } from 'vue'
import { useValidatorDashboardOverviewStore } from '~/stores/dashboard/useValidatorDashboardOverviewStore'
import type { ApiPagingResponse } from '~/types/api/common'
import type { VDBOverviewGroup } from '~/types/api/validator_dashboard'
import type { DashboardKey } from '~/types/dashboard'
import type { Cursor } from '~/types/datatable'
import { getSortOrder } from '~/utils/table'

const { t: $t } = useI18n()
// const { fetch } = useCustomFetch()

interface Props {
  dashboardKey: DashboardKey;
}
const props = defineProps<Props>()

const { width } = useWindowSize()

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

const onClose = () => {
  visible.value = false
}

const addGroup = async () => {
  // TODO call API to add Group
  warn(`Add group ${newGroupName.value}`)
  await getOverview(props.dashboardKey)
}

/*
TODO: add edit group once we have our edit component
const editGroup = (row: VDBOverviewGroup, newName?: string) => {
  // TODO open modal to edit multiple
  warn(`Edit group ${row.name} [${row.id}] -> ${newName}`)
} */

const removeGroup = async (row: VDBOverviewGroup) => {
  // TODO: display confirm modal if user really wants to remove validator.
  // If multiple are selected ask if he wnats to remove all selected
  alert(`remove val ${row.id}`)
  await getOverview(props.dashboardKey)
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
const total = ref(40)
const premiumLimit = computed(() => (data.value?.paging?.total_count ?? 0) >= total.value)

</script>

<template>
  <BcDialog
    v-model="visible"
    :header="$t('dashboard.validator.group_management.title')"
    class="validator-group-managment-modal-container"
  >
    <template v-if="!size.showSubTitle" #header>
      <span class="hdden-title" />
    </template>
    <BcTableControl :search-placeholder="$t('dashboard.validator.group_management.search_placeholder')" @set-search="setSearch">
      <template #header-left>
        <span v-if="size.showSubTitle"> {{ $t('dashboard.validator.group_management.sub_title', {dashboardName}) }}</span>
        <span v-else class="small-title">{{ $t('dashboard.validator.group_management.title') }}</span>
      </template>
      <template #bc-table-sub-header>
        <div class="add-row">
          <InputText v-model="newGroupName" class="search-input" />
          <Button class="p-button-icon-only" style="display: inline;" @click="addGroup">
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
            <Column field="name" :sortable="true" :header="$t('dashboard.validator.group_management.col.name')">
              <template #body="slotProps">
                {{ slotProps.data.name }}
              </template>
            </Column>
            <Column field="id" :sortable="true" :header="$t('dashboard.validator.group_management.col.id')" />

            <Column field="count" :sortable="true" :header="$t('dashboard.validator.group_management.col.count')">
              <template #body="slotProps">
                <!-- TODO: add formating-->
                {{ slotProps.data.count ?? 0 }}
              </template>
            </Column>
            <Column field="action">
              <template #body="slotProps">
                <div class="action-col">
                  <FontAwesomeIcon v-if="slotProps.data.id" :icon="faTrash" class="link" @click="removeGroup(slotProps.data)" />
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
          <div v-if="total" class="labels" :class="premiumLimit">
            <span>{{ data.paging.total_count }}/{{ total }}</span>
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
}

:global(.validator-group-managment-modal-container .bc-table-header .side:first-child) {
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
</style>
