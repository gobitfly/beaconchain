<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import {
  faArrowUpRightFromSquare
} from '@fortawesome/pro-solid-svg-icons'
import IconValidator from '../icon/IconValidator.vue'
import IconAccount from '../icon/IconAccount.vue'
import type { Cursor } from '~/types/datatable'
import { getGroupLabel } from '~/utils/dashboard/group'

defineEmits<{(e: 'openDialog'): void }>()

const cursor = ref<Cursor>()
const pageSize = ref<number>(10)
const { t: $t } = useI18n()

const { onSort, setCursor, setPageSize, setSearch, notificationsDashboards, query, isLoading } = useNotificationsDashboardStore()

const { groups } = useValidatorDashboardGroups()

const { width } = useWindowSize()
const colsVisible = computed(() => {
  return {
    notifications: width.value > 1024,
    dashboard: width.value >= 640,
    groups: width.value >= 640
  }
})

const groupNameLabel = (groupId?: number) => {
  return getGroupLabel($t, groupId, groups.value, 'Σ')
}

const openDialog = () => {
  // TODO: implement dialog
  alert('not implemented yet 😪')
}

const notificationsDashboardsWithUniqueIdentifier = computed(() => {
  if (!notificationsDashboards.value) {
    return
  }
  return {
    paging: notificationsDashboards.value.paging,
    // TODO: set unique identifier after backend is ready
    data: notificationsDashboards.value.data
      .map((item, index) => ({ ...item, identifier: index }))
      // .filter(() => false) // comment in to test empty table
  }
})
</script>

<template>
  <div>
    <BcTableControl
      :title="$t('notifications.dashboards.title')"
      :search-placeholder="$t('notifications.dashboards.search_placeholder')"
      @set-search="setSearch"
    >
      <template #header-left>
        NETWORK_SWITCHER_COMPONENT
      </template>
      <template #table>
        <ClientOnly fallback-tag="span">
          <BcTable
            :data="notificationsDashboardsWithUniqueIdentifier"
            data-key="dashboardId"
            :expandable="!colsVisible.notifications"
            :cursor="cursor"
            :page-size="pageSize"
            :selected-sort="query?.sort"
            :loading="isLoading"
            @set-cursor="setCursor"
            @sort="onSort"
            @set-page-size="setPageSize"
          >
            <Column
              field="network"
              sortable
              header-class="col-header-network"
              body-class="col-network"
            >
              <template #body="slotProps">
                <div class="icon-wrapper">
                  <IconNetwork colored :chain-id="slotProps.data.dashboardNetwork" class="icon-network" />
                </div>
              </template>
            </Column>
            <Column
              field="age"
              sortable
              header-class="col-age"
              body-class="col-age"
            >
              <template #header>
                <BcTableAgeHeader />
              </template>
              <template #body="slotProps">
                <BcFormatTimePassed :value="slotProps.data.timestamp" type="go-timestamp" />
              </template>
            </Column>
            <Column
              v-if="colsVisible.dashboard"
              field="dashboard"
              :sortable="true"
              header-class="col-dashboard"
              body-class="col-dashboard"
              :header="$t('notifications.col.dashboard')"
            >
              <template #body="slotProps">
                <NotificationsDashboardsTableItemDashboard
                  :type="slotProps.data.entity.type"
                  :dashboard-id="slotProps.data.dashboardId"
                  :dashboard-name="slotProps.data.dashboardName"
                />
              </template>
            </Column>
            <Column
              v-if="colsVisible.groups"
              field="group_id"
              body-class="col-group"
              header-class="col-group"
              :header="$t('notifications.col.group')"
            >
              <template #body="slotProps">
                <span>
                  {{ groupNameLabel(slotProps.data.group_id) }}
                </span>
              </template>
            </Column>
            <Column
              field="entity"
              sortable
              header-class="col-entity"
              body-class="col-entity"
              :header="$t('notifications.dashboards.col.entity')"
            >
              <template #body="slotProps">
                <div class="entity">
                  <template v-if="slotProps.data.entity.type === 'validator'">
                    <IconValidator class="icon-dashboard-type" />
                    {{ slotProps.data.entity.count }}
                    <span>
                      {{ $t('notifications.dashboards.entity.validators', slotProps.data.entity.count) }}
                    </span>
                  </template>
                  <template v-if="slotProps.data.entity.type === 'account'">
                    <IconAccount class="icon-dashboard-type" />
                    {{ slotProps.data.entity.count }}
                    <span>
                      {{ $t('notifications.dashboards.entity.accounts', slotProps.data.entity.count) }}
                    </span>
                  </template>
                  <FontAwesomeIcon
                    class="link"
                    :icon="faArrowUpRightFromSquare"
                    @click="openDialog"
                  />
                </div>
              </template>
            </Column>

            <Column
              v-if="colsVisible.notifications"
              field="notification"
              body-class="notification"
              header-class="notification"
              :header="$t('notifications.dashboards.col.notification')"
            >
              <template #body="slotProps">
                {{ slotProps.data.notification.join(', ') }}
              </template>
            </Column>
            <template #expansion="slotProps">
              <div class="expansion">
                <div class="label-dashboard">
                  {{ $t('notifications.dashboards.expansion.label-dashboard') }}
                </div>
                <NotificationsDashboardsTableItemDashboard
                  :type="slotProps.data.entity.type"
                  :dashboard-id="slotProps.data.dashboardId"
                  :dashboard-name="slotProps.data.dashboardName"
                />
                <div class="label-group">
                  {{ $t('notifications.dashboards.expansion.label-group') }}
                </div>
                <div class="group">
                  {{ groupNameLabel(slotProps.data.group_id) }}
                </div>
                <div class="label-notification">
                  {{ $t('notifications.dashboards.expansion.label-notification') }}
                </div>
                <div class="notification">
                  {{ slotProps.data.notification.join(', ') }}
                </div>
              </div>
            </template>
            <template #empty>
              <NotificationsDashboardsTableEmpty
                v-if="!notificationsDashboardsWithUniqueIdentifier?.data.length"
                @open-dialog="$emit('openDialog')"
              />
            </template>
            <!-- TODO: implement number of subscriptions -->
            <template #bc-table-footer-right>
              <template v-if="width < 1024">
                {{ $t('notifications.dashboards.footer.subscriptions.validators_shortened', { count: 1}) }} | {{ $t('notifications.dashboards.footer.subscriptions.accounts_shortened', { count: 1}) }}
              </template>
              <template v-else>
                <div>{{ $t('notifications.dashboards.footer.subscriptions.validators', { count: 1}) }}</div>
                <div>{{ $t('notifications.dashboards.footer.subscriptions.accounts', { count: 1}) }}</div>
              </template>
            </template>
          </BcTable>
        </ClientOnly>
      </template>
    </BcTableControl>
  </div>
</template>

<style lang="scss" scoped>
@use "~/assets/css/utils.scss";

$breakpoint-sm: 640px;
$breakpoint-lg: 1024px;

:deep(.col-header-network .p-column-header-content){
  justify-content: center;
}

:deep(.expander) {
  @include utils.set-all-width(22px);
  @media (max-width: $breakpoint-sm) {
    padding-inline: 4px !important;
  }
}
:deep(.col-header-network) {
  @include utils.set-all-width(35px);
  padding-left: 0px !important;
}
:deep(.col-network) {
  @include utils.set-all-width(35px);
  padding-inline: 0px !important;
}
:deep(.col-age) {
  @media (max-width: $breakpoint-lg) {
    @include utils.set-all-width(78px);
  }
  *:not([data-pc-section="sort"]) {
    @include utils.truncate-text;
  }
}
:deep(.col-dashboard) {
  @media (max-width: $breakpoint-lg) {
    @include utils.set-all-width(105px);
  }
  span:not([data-pc-section="sort"]) {
    @include utils.truncate-text;
  }
}
:deep(.col-group) {
  @include utils.truncate-text;
  @media (max-width: $breakpoint-lg) {
    @include utils.set-all-width(80px);
  }
}
:deep(.col-entity) {
  padding-right: 3px !important;
  @media (max-width: $breakpoint-lg) {
    @include utils.set-all-width(85px);
    padding-left: 0px !important;
  }
  *:not([data-pc-section="sort"]) {
    @include utils.truncate-text;
  }
}

:deep(.bc-table-header) {
  .h1 {
    display: none;
  }

  @media (min-width: $breakpoint-lg) {
    .h1 {
      display: block;
    }
  }
}
:deep(.right-info){
  flex-direction: column;
  justify-content: center;
}
.icon-wrapper {
  text-align: center;

  .icon-network {
    height: 14px;
    width: 14px;
  }
}
svg {
  flex-shrink: 0;
}
.entity {
  display: flex;
  align-items: center;
  gap: var(--padding-small);
  @media (min-width: $breakpoint-lg) {
    gap: var(--padding);
  }
}
.expansion {
  //1. duplicating primevue padding
  display: grid;
  grid-template-columns: 120px 1fr;
  column-gap: var(--padding-xl);
  row-gap: 14px; //1.
  background-color: var(--table-header-background);
  padding: 14px 7px; //1.
  @media (min-width: $breakpoint-sm) {
    padding-left: 14px !important;
  }
}
.label-group, .label-notification {
  font-weight: var(--standard_text_medium_font_weight);
}
</style>
