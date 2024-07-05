<script lang="ts" setup>import {
  faTrash
} from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

import type { DataTableSortEvent } from 'primevue/datatable'
import type { Cursor, TableQueryParams } from '~/types/datatable'
import { useNotificationsManagementDashboardStore } from '~/stores/notifications/useNotificationsManagementDashboardStore'
import { getGroupLabel } from '~/utils/dashboard/group'
import { type NotificationsManagementDashboardRow } from '~/types/notifications/management'

const cursor = ref<Cursor>()
const pageSize = ref<number>(10)
const { t: $t } = useI18n()

const { dashboards, query: lastQuery, isLoading, getDashboards } = useNotificationsManagementDashboardStore()
const { value: query, temp: tempQuery, bounce: setQuery } = useDebounceValue<TableQueryParams | undefined>({ limit: pageSize.value, sort: 'dashboard_id:desc' }, 500)

const { groups } = useValidatorDashboardGroups()

const { width } = useWindowSize()
const colsVisible = computed(() => {
  return {
    networks: width.value > 1101,
    webhook: width.value >= 945,
    subscriptions: width.value >= 725
  }
})

const loadData = (query?: TableQueryParams) => {
  if (!query) {
    query = { limit: pageSize.value, sort: 'dashboard_id:desc' }
  }
  setQuery(query, true, true)
}

watch(query, (q) => {
  getDashboards(q)
}, { immediate: true })

const groupNameLabel = (groupId?: number) => {
  return getGroupLabel($t, groupId, groups.value, 'Σ')
}

const onSort = (sort: DataTableSortEvent) => {
  loadData(setQuerySort(sort, lastQuery.value))
}

const setCursor = (value: Cursor) => {
  cursor.value = value
  loadData(setQueryCursor(value, lastQuery.value))
}

const setPageSize = (value: number) => {
  pageSize.value = value
  loadData(setQueryPageSize(value, lastQuery.value))
}

const setSearch = (value?: string) => {
  loadData(setQuerySearch(value, lastQuery.value))
}

const wrappedDashboards = computed(() => {
  if (!dashboards.value) {
    return
  }
  return {
    paging: dashboards.value.paging,
    data: dashboards.value.data.map(d => ({ ...d, identifier: `${d.dashboard_type}-${d.dashboard_id}-${d.group_id}` }))
  }
})

const onEdit = (col: 'delete' | 'subscriptions' | 'webhook' | 'networks', row: NotificationsManagementDashboardRow) => {
  switch (col) {
    case 'subscriptions':
      alert('TODO: edit subscriptions' + row.group_id)
      break
    case 'webhook':
      alert('TODO: edit webhook' + row.group_id)
      break
    case 'networks':
      alert('TODO: edit networks' + row.group_id)
      break
    case 'delete':
      alert('TODO: delete' + row.group_id)
      break
  }
}

</script>

<template>
  <div>
    <Teleport to="#notifications-management-search-placholder">
      <BcContentFilter :search-placeholder="$t('placeholder')" class="search" @filter-changed="setSearch" />
    </Teleport>

    <ClientOnly fallback-tag="span">
      <BcTable
        :data="wrappedDashboards"
        data-key="identifier"
        :expandable="!colsVisible.networks"
        class="notifications-management-dashboard-table"
        :cursor="cursor"
        :page-size="pageSize"
        :selected-sort="tempQuery?.sort"
        :loading="isLoading"
        @set-cursor="setCursor"
        @sort="onSort"
        @set-page-size="setPageSize"
      >
        <Column
          field="dashboard_id"
          body-class="dashboard-col"
          header-class="dashboard-col"
          :sortable="true"
          :header="$t('notifications.col.dashboard')"
        >
          <template #body="slotProps">
            <span>
              {{ slotProps.data.dashboard_name }}
            </span>
          </template>
        </Column>
        <Column
          field="group_id"
          body-class="group-col"
          header-class="group-col"
          :sortable="true"
          :header="$t('notifications.col.group')"
        >
          <template #body="slotProps">
            <span>
              {{ groupNameLabel(slotProps.data.group_id) }}
            </span>
          </template>
        </Column>
        <Column
          v-if="colsVisible.subscriptions"
          field="subscriptions"
          body-class="subscriptions-col"
          header-class="subscriptions-col"
          :header="$t('notifications.col.subscriptions')"
        >
          <template #body="slotProps">
            <BcTablePopoutEdit
              :truncate-text="true"
              :label="slotProps.data.subscriptions.join(', ')"
              @on-edit="onEdit('subscriptions', slotProps.data)"
            />
          </template>
        </Column>
        <Column
          v-if="colsVisible.webhook"
          field="webhook"
          body-class="webhook-col"
          header-class="webhook-col"
          :header="$t('notifications.col.webhook')"
        >
          <template #body="slotProps">
            <BcTablePopoutEdit :truncate-text="true" :label="slotProps.data.webhook.url" @on-edit="() => onEdit('webhook', slotProps.data)" />
          </template>
        </Column>
        <Column
          v-if="colsVisible.networks"
          field="networks"
          body-class="networks-col"
          header-class="networks-col"
          :header="$t('notifications.col.networks')"
        >
          <template #body="slotProps">
            <BcTablePopoutEdit
              :truncate-text="true"
              :no-icon="slotProps.data.dashboard_type === 'validator'"
              @on-edit="onEdit('networks', slotProps.data)"
            >
              <template #content>
                <IconNetwork
                  v-for="chainId in slotProps.data.networks"
                  :key="chainId"
                  :colored="true"
                  class="network-icon"
                  :chain-id="chainId"
                />
              </template>
            </BcTablePopoutEdit>
          </template>
        </Column>
        <Column field="action" body-class="action-col" header-class="action-col">
          <template #body="slotProps">
            <!--TODO: once we have our api check how to identify 'deleted' rows-->
            <div class="action-row">
              <FontAwesomeIcon
                :disabled="!slotProps.data.subscriptions?.length ? true : null"
                :icon="faTrash"
                class="link"
                @click="onEdit('delete', slotProps.data)"
              />
            </div>
          </template>
        </Column>
        <template #expansion="slotProps">
          <div class="expansion">
            <div class="info">
              <div class="label">
                {{ $t('notifications.col.subscriptions') }}
              </div>

              <BcTablePopoutEdit
                class="value"
                :label="slotProps.data.subscriptions.join(', ')"
                @on-edit="onEdit('subscriptions', slotProps.data)"
              />
            </div>
            <div class="info">
              <div class="label">
                {{ $t('notifications.col.webhook') }}
              </div>

              <BcTablePopoutEdit
                class="value"
                :label="slotProps.data.webhook.url"
                @on-edit="() => onEdit('webhook', slotProps.data)"
              />
            </div>
            <div class="info">
              <div class="label">
                {{ $t('notifications.col.networks') }}
              </div>

              <BcTablePopoutEdit
                class="value"
                :no-icon="slotProps.data.dashboard_type === 'validator'"
                @on-edit="onEdit('networks', slotProps.data)"
              >
                <template #content>
                  <div class="newtork-row">
                    <IconNetwork
                      v-for="chainId in slotProps.data.networks"
                      :key="chainId"
                      :colored="true"
                      class="network-icon"
                      :chain-id="chainId"
                    />
                  </div>
                </template>
              </BcTablePopoutEdit>
            </div>
          </div>
        </template>
      </BcTable>
    </ClientOnly>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/main.scss";
@use "~/assets/css/utils.scss";

.expansion {
  @include main.container;
  padding: var(--padding);
  display: flex;
  flex-direction: column;
  gap: var(--padding);
  font-size: var(--small_text_font_size);

  .info {
    display: flex;
    gap: var(--padding);

    .label {
      flex-shrink: 0;
      font-weight: var(--standard_text_bold_font_weight);
      width: 100px;
    }

    .value {
      width: 197px;
    }
  }
}

.network-icon {
  margin-right: var(--padding);
  height: 20px;
  width: 20px;
}
.newtork-row{
  display: flex;
}

.action-row {
  display: flex;
  justify-content: flex-end;
}

:deep(.notifications-management-dashboard-table) {

  .dashboard-col,
  .group-col {
    @include utils.truncate-text;
    @include utils.set-all-width(210px);

    @media (max-width: 1460px) {
      @include utils.set-all-width(140px);
    }

    @media (max-width: 5205px) {
      @include utils.set-all-width(120px);
    }
  }

  .webhook-col,
  .subscriptions-col {
    @include utils.set-all-width(340px);

    @media (max-width: 1360px) {
      @include utils.set-all-width(240px);
    }
  }

  /*.action-col{
    @include utils.set-all-width(40px);
  }*/
  .networks-col {
    @include utils.set-all-width(156px);
  }
}
</style>
