<script setup lang="ts">
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'
import { faArrowUpRightFromSquare } from '@fortawesome/pro-solid-svg-icons'
import IconValidator from '../icon/IconValidator.vue'
import IconAccount from '../icon/IconAccount.vue'
import type { Cursor } from '~/types/datatable'
import type { DashboardType } from '~/types/dashboard'
import type { ChainIDs } from '~/types/network'
import type { NotificationDashboardsTableRow } from '~/types/api/notifications'
import { NotificationsDashboardDialogEntity } from '#components'

defineEmits<{ (e: 'openDialog'): void }>()

const cursor = ref<Cursor>()
const pageSize = ref<number>(10)
const { t: $t } = useTranslation()

// TODO: replace currentNetwork with selection from NETWORK_SWITCHER_COMPONENT that has yet to be implemented
const { currentNetwork } = useNetworkStore()
const networkId = ref<ChainIDs>(currentNetwork.value ?? 1)

const {
  isLoading,
  notificationsDashboards,
  onSort,
  query,
  setCursor,
  setPageSize,
  setSearch,
} = useNotificationsDashboardStore(networkId)

const { width } = useWindowSize()
const colsVisible = computed(() => {
  return {
    dashboard: width.value >= 640,
    groups: width.value >= 640,
    notifications: width.value > 1024,
  }
})

const getDashboardType = (isAccount: boolean): DashboardType => isAccount ? 'account' : 'validator'
const { overview } = useNotificationsDashboardOverviewStore()
const mapEventtypeToText = (eventType: NotificationDashboardsTableRow['event_types'][number]) => {
  switch (eventType) {
    case 'attestation_missed':
      return $t('notifications.dashboards.event_type.attestation_missed')
    case 'group_offline':
      return $t('notifications.dashboards.event_type.group_offline')
    case 'group_online':
      return $t('notifications.dashboards.event_type.group_online')
    case 'incoming_tx':
      return $t('notifications.dashboards.event_type.incoming_tx')
    case 'max_collateral':
      return $t('notifications.dashboards.event_type.max_collateral')
    case 'min_collateral':
      return $t('notifications.dashboards.event_type.min_collateral')
    case 'outgoing_tx':
      return $t('notifications.dashboards.event_type.outgoing_tx')
    case 'proposal_missed':
      return $t('notifications.dashboards.event_type.proposal_missed')
    case 'proposal_success':
      return $t('notifications.dashboards.event_type.proposal_success')
    case 'proposal_upcoming':
      return $t('notifications.dashboards.event_type.proposal_upcoming')
    case 'sync':
      return $t('notifications.dashboards.event_type.sync')
    case 'transfer_erc20':
      return $t('notifications.dashboards.event_type.transfer_erc20')
    case 'transfer_erc721':
      return $t('notifications.dashboards.event_type.transfer_erc721')
    case 'transfer_erc1155':
      return $t('notifications.dashboards.event_type.transfer_erc1155')
    case 'validator_got_slashed':
      return $t('notifications.dashboards.event_type.validator_got_slashed')
    case 'validator_has_slashed':
      return $t('notifications.dashboards.event_type.validator_has_slashed')
    case 'validator_offline':
      return $t('notifications.dashboards.event_type.validator_offline')
    case 'validator_online':
      return $t('notifications.dashboards.event_type.validator_online')
    case 'withdrawal':
      return $t('notifications.dashboards.event_type.withdrawal')
    default:
      logError(`Unknown dashboard notification event_type: ${eventType}`)
      return eventType
  }
}
const textDashboardNotifications = (event_types: NotificationDashboardsTableRow['event_types']) => {
  return event_types.map(mapEventtypeToText).join(', ')
}

const dialog = useDialog()

const showDialog = (row: { identifier: string } & NotificationDashboardsTableRow) => {
  dialog.open(NotificationsDashboardDialogEntity, {
    data: {
      dashboard_id: row.dashboard_id,
      epoch: row.epoch,
      group_id: row.group_id,
      group_name: row.group_name,
      identifier: row.identifier,
    },
  })
}
</script>

<template>
  <div>
    <BcTableControl
      :title="$t('notifications.dashboards.title')"
      :search-placeholder="$t('notifications.dashboards.search_placeholder')"
      @set-search="setSearch"
    >
      <template #table>
        <ClientOnly fallback-tag="span">
          <BcTable
            :data="addIdentifier(notificationsDashboards, 'is_account_dashboard', 'dashboard_id', 'group_id', 'epoch')"
            data-key="identifier"
            :expandable="!colsVisible.notifications"
            :cursor
            :page-size
            :selected-sort="query?.sort"
            :loading="isLoading"
            @set-cursor="setCursor"
            @sort="onSort"
            @set-page-size="setPageSize"
          >
            <Column
              field="chain_id"
              sortable
              header-class="col-header-network"
              body-class="col-network"
            >
              <template #body="slotProps">
                <div class="icon-wrapper">
                  <IconNetwork
                    colored
                    :chain-id="slotProps.data.chain_id"
                    class="icon-network"
                  />
                </div>
              </template>
            </Column>
            <Column
              field="epoch"
              sortable
              header-class="col-age"
              body-class="col-age"
            >
              <template #header>
                <BcTableAgeHeader />
              </template>
              <template #body="slotProps">
                <BcFormatTimePassed
                  :value="slotProps.data.epoch"
                  type="epoch"
                />
              </template>
            </Column>
            <Column
              v-if="colsVisible.dashboard"
              field="dashboard_id"
              :sortable="true"
              header-class="col-dashboard"
              body-class="col-dashboard"
              :header="$t('notifications.col.dashboard')"
            >
              <template #body="slotProps">
                <NotificationsDashboardsTableItemDashboard
                  :type="getDashboardType(slotProps.data.is_account_dashboard)"
                  :dashboard-id="slotProps.data.dashboard_id"
                  :dashboard-name="slotProps.data.dashboard_name"
                />
              </template>
            </Column>
            <Column
              v-if="colsVisible.groups"
              field="group_name"
              body-class="col-group"
              header-class="col-group"
              :header="$t('notifications.col.group')"
            >
              <template #body="slotProps">
                <span>
                  {{ slotProps.data.group_name }}
                </span>
              </template>
            </Column>
            <Column
              field="entity_count"
              header-class="col-entity"
              body-class="col-entity"
              :header="$t('notifications.dashboards.col.entity')"
            >
              <template #body="slotProps">
                <div class="entity">
                  <template v-if="!slotProps.data.is_account_dashboard">
                    <IconValidator class="icon-dashboard-type" />
                    {{ slotProps.data.entity_count }}
                    <span>
                      {{
                        $t(
                          "notifications.dashboards.entity.validators",
                          slotProps.data.entity_count,
                        )
                      }}
                    </span>
                  </template>
                  <template v-else>
                    <IconAccount class="icon-dashboard-type" />
                    {{ slotProps.data.entity_count }}
                    <span>
                      {{
                        $t(
                          "notifications.dashboards.entity.accounts",
                          slotProps.data.entity_count,
                        )
                      }}
                    </span>
                  </template>
                  <BcButtonIcon
                    screenreader-text="Open notification details"
                    @click="showDialog(slotProps.data)"
                  >
                    <FontAwesomeIcon
                      class="link"
                      :icon="faArrowUpRightFromSquare"
                    />
                  </BcButtonIcon>
                </div>
              </template>
            </Column>

            <Column
              v-if="colsVisible.notifications"
              field="notification"
              body-class="col-notification"
              header-class="col-notification"
              :header="$t('notifications.dashboards.col.notification')"
            >
              <template #body="slotProps">
                {{ textDashboardNotifications(slotProps.data.event_types) }}
              </template>
            </Column>
            <template #expansion="slotProps">
              <div class="expansion">
                <div class="label-dashboard">
                  {{ $t("notifications.dashboards.expansion.label_dashboard") }}
                </div>
                <NotificationsDashboardsTableItemDashboard
                  :type="getDashboardType(slotProps.data.is_account_dashboard)"
                  :dashboard-id="slotProps.data.dashboard_id"
                  :dashboard-name="slotProps.data.dashboard_name"
                />
                <div class="label-group">
                  {{ $t("notifications.dashboards.expansion.label_group") }}
                </div>
                <div class="group">
                  {{ slotProps.data.group_name }}
                </div>
                <div class="label-notification">
                  {{
                    $t("notifications.dashboards.expansion.label_notification")
                  }}
                </div>
                <div class="notification">
                  {{ slotProps.data.event_types.join(", ") }}
                </div>
              </div>
            </template>
            <template #empty>
              <NotificationsDashboardsTableEmpty
                v-if="!notificationsDashboards?.data.length"
                @open-dialog="$emit('openDialog')"
              />
            </template>
            <template #bc-table-footer-right>
              <template v-if="width < 1024">
                {{
                  $t(
                    "notifications.dashboards.footer.subscriptions.validators_shortened",
                    { count: overview?.vdb_subscriptions_count })
                }}
                |
                {{
                  $t(
                    "notifications.dashboards.footer.subscriptions.accounts_shortened",
                    { count: overview?.adb_subscriptions_count })
                }}
              </template>
              <template v-else>
                <div>
                  {{
                    $t(
                      "notifications.dashboards.footer.subscriptions.validators",
                      { count: overview?.vdb_subscriptions_count })

                  }}
                </div>
                <BcFeatureFlag
                  feature="feature-account_dashboards"
                >
                  <div>
                    {{
                      $t(
                        "notifications.dashboards.footer.subscriptions.accounts",
                        { count: overview?.adb_subscriptions_count })

                    }}
                  </div>
                </BcFeatureFlag>
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

:deep(.col-header-network .p-column-header-content) {
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
:deep(.col-notification) {
  @include utils.set-all-width(240px);
  @include utils.truncate-text;
}

:deep(.right-info) {
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
.label-group,
.label-notification {
  font-weight: var(--standard_text_medium_font_weight);
}
</style>
