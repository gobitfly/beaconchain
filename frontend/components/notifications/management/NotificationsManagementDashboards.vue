<script lang="ts" setup>
import {
  faDesktop, faTrash, faUser,
} from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

import type { ApiPagingResponse } from '~/types/api/common'
import type {
  NotificationSettingsAccountDashboard,
  NotificationSettingsDashboardsTableRow,
  NotificationSettingsValidatorDashboard,
} from '~/types/api/notifications'
import type { DashboardType } from '~/types/dashboard'
import { useNotificationsManagementDashboards } from '~/composables/notifications/useNotificationsManagementDashboards'
import {
  NotificationsManagementModalDashboardsDelete,
  NotificationsManagementModalWebhook,
  NotificationsManagementSubscriptionDialog,
} from '#components'

interface WrappedRow extends NotificationSettingsDashboardsTableRow {
  dashboard_name: string,
  dashboard_type: DashboardType,
  identifier: string,
  subscriptions: string[],
}

const toast = useBcToast()
const { t: $t } = useTranslation()
const dialog = useDialog()
const { refreshOverview } = useNotificationsDashboardOverviewStore()
const {
  cursor,
  dashboards,
  deleteDashboardNotifications,
  isLoading,
  onSort,
  pageSize,
  query,
  saveSubscriptions,
  setCursor,
  setPageSize,
  setSearch,
} = useNotificationsManagementDashboards()
const { width } = useWindowSize()

const colsVisible = computed(() => {
  return {
    networks: width.value > 1101,
    subscriptions: width.value >= 725,
    webhook: width.value >= 945,
  }
})

const wrappedDashboards: ComputedRef<
  ApiPagingResponse<WrappedRow> | undefined
> = computed(() => {
  if (!dashboards.value) {
    return
  }
  return {
    data: dashboards.value.data.map(dashboard => ({
      ...dashboard,
      dashboard_type: dashboardType(dashboard),
      identifier: `${dashboardType(dashboard)}-${dashboard.dashboard_id}-${dashboard.group_id}`,
      subscriptions: getSubscriptions(dashboard),
    })),
    paging: dashboards.value.paging,
  }

  function dashboardType(
    row: NotificationSettingsDashboardsTableRow,
  ): DashboardType {
    return row.is_account_dashboard ? 'account' : 'validator'
  }

  function getSubscriptions(
    row: NotificationSettingsDashboardsTableRow,
  ): string[] {
    const result: string[] = []
    if (row.is_account_dashboard) {
      const settingsAccountDashboard = row.settings as NotificationSettingsAccountDashboard
      if (settingsAccountDashboard.is_incoming_transactions_subscribed) {
        result.push($t('notifications.subscriptions.accounts.incoming_transactions.label'))
      }
      if (settingsAccountDashboard.is_outgoing_transactions_subscribed) {
        result.push($t('notifications.subscriptions.accounts.outgoing_transactions.label'))
      }
      if (settingsAccountDashboard.is_erc20_token_transfers_subscribed) {
        result.push($t('notifications.subscriptions.accounts.erc20_token_transfers.label'))
      }
      if (settingsAccountDashboard.is_erc721_token_transfers_subscribed) {
        result.push($t('notifications.subscriptions.accounts.erc721_token_transfers.label'))
      }
      if (settingsAccountDashboard.is_erc1155_token_transfers_subscribed) {
        result.push($t('notifications.subscriptions.accounts.erc1155_token_transfers.label'))
      }
      if (settingsAccountDashboard.is_ignore_spam_transactions_enabled) {
        result.push($t('notifications.subscriptions.accounts.ignore_spam_transactions.label'))
      }
      return result
    }
    const settingsValidatorDashboard = row.settings as NotificationSettingsValidatorDashboard
    if (settingsValidatorDashboard.is_validator_offline_subscribed) {
      result.push($t('notifications.subscriptions.validators.validator_is_offline.label'))
    }
    if (settingsValidatorDashboard.is_attestations_missed_subscribed) {
      result.push($t('notifications.subscriptions.validators.attestation_missed.label'))
    }
    if (settingsValidatorDashboard.is_block_proposal_subscribed) {
      result.push($t('notifications.subscriptions.validators.block_proposal.label'))
    }
    if (settingsValidatorDashboard.is_upcoming_block_proposal_subscribed) {
      result.push($t('notifications.subscriptions.validators.upcoming_block_proposal.label'))
    }
    if (settingsValidatorDashboard.is_sync_subscribed) {
      result.push($t('notifications.subscriptions.validators.sync_committee.label'))
    }
    if (settingsValidatorDashboard.is_withdrawal_processed_subscribed) {
      result.push($t('notifications.subscriptions.validators.withdrawal_processed.label'))
    }
    if (settingsValidatorDashboard.is_slashed_subscribed) {
      result.push($t('notifications.subscriptions.validators.validator_got_slashed.label'))
    }
    if (settingsValidatorDashboard.is_min_collateral_subscribed) {
      result.push($t('notifications.subscriptions.validators.min_collateral_reached.label'))
    }
    if (settingsValidatorDashboard.is_group_efficiency_below_subscribed) {
      result.push($t('notifications.subscriptions.accounts.group_efficiency.label'))
    }
    if (settingsValidatorDashboard.is_max_collateral_subscribed) {
      result.push($t('notifications.subscriptions.validators.max_collateral_reached.label'))
    }
    return result
  }
})

const handleSubscriptionChange = (
  settings: Omit<NotificationSettingsValidatorDashboard, 'is_webhook_discord_enabled' | 'webhook_url'>,
  row: WrappedRow,
) => {
  saveSubscriptions({
    dashboard_id: row.dashboard_id,
    group_id: row.group_id,
    settings: {
      ...settings,
      is_webhook_discord_enabled: row.settings.is_webhook_discord_enabled,
      webhook_url: row.settings.webhook_url,
    },
  })
}

type Dialog = 'delete' | 'networks' | 'subscriptions' | 'webhook'
const onEdit = (col: Dialog, row: WrappedRow) => {
  switch (col) {
    case 'delete':
      if (row.dashboard_type === 'validator') {
        return dialog.open(NotificationsManagementModalDashboardsDelete, {
          data: row,
          emits: {
            onDelete: handleDelete,
          },
        })
      }
      alert('TODO: Subscription Dialog for Account Dashboards')
      break
    case 'subscriptions':
      dialog.open(NotificationsManagementSubscriptionDialog, {
        data: row.settings,
        emits: {
          onChangeSettings: (
            settings: Omit<NotificationSettingsValidatorDashboard, 'is_webhook_discord_enabled' | 'webhook_url'>,
          ) => handleSubscriptionChange(settings, row),
        },
      })
      break
    case 'webhook':
      dialog.open(NotificationsManagementModalWebhook, {
        data: {
          is_webhook_discord_enabled: row.settings.is_webhook_discord_enabled,
          webhook_url: row.settings.webhook_url,
        },
        emits: {
          onSave: async (
            webhookData: Pick<NotificationSettingsValidatorDashboard, 'is_webhook_discord_enabled' | 'webhook_url'>,
            closeCallback: () => void,
          ) => {
            try {
              await saveSubscriptions({
                dashboard_id: row.dashboard_id,
                group_id: row.group_id,
                settings: {
                  ...row.settings as NotificationSettingsValidatorDashboard,
                  ...webhookData,
                },
              })
              closeCallback()
            }
            catch {
              toast.showError({
                detail: $t('notifications.subscriptions.error_message'),
                group: $t('notifications.subscriptions.error_group'),
                summary: $t('notifications.subscriptions.error_title'),
              })
            }
          },
        },
      })
      break
  }
}

function getTypeIcon(type: DashboardType) {
  if (type === 'validator') {
    return faDesktop
  }
  return faUser
}
const handleDelete = (payload: Parameters<typeof deleteDashboardNotifications>[0]) => {
  deleteDashboardNotifications(payload).then(() => refreshOverview())
}
</script>

<template>
  <div>
    <Teleport to="#notifications-management-search-placholder">
      <BcContentFilter
        :search-placeholder="$t('notifications.dashboards.search_placeholder')"
        class="search"
        @filter-changed="setSearch"
      />
    </Teleport>

    <ClientOnly fallback-tag="span">
      <BcTable
        :data="wrappedDashboards"
        data-key="identifier"
        :expandable="!colsVisible.networks"
        class="notifications-management-dashboard-table"
        :cursor
        :page-size
        :selected-sort="query?.sort"
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
              <FontAwesomeIcon
                :icon="getTypeIcon(slotProps.data.dashboard_type)"
                class="type-icon"
              />
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
            {{ slotProps.data.group_name }}
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
            <BcTablePopoutEdit
              :truncate-text="true"
              :label="slotProps.data.settings.webhook_url"
              @on-edit="() => onEdit('webhook', slotProps.data)"
            />
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
            <BcNetworkSelector
              :readonly-networks="slotProps.data.chain_ids"
            />
          </template>
        </Column>
        <Column
          field="action"
          body-class="action-col"
          header-class="action-col"
        >
          <template #body="slotProps">
            <div class="action-row">
              <BcButtonIcon
                :screenreader-text="
                  $t('notifications.clients.settings.screenreader.delete_notifications_for_dashboard_id',
                     { dashboard_id: slotProps.data.dashboard_name },
                  )"
                :disabled="!slotProps.data.subscriptions?.length ? true : null"
                @click="onEdit('delete', slotProps.data)"
              >
                <FontAwesomeIcon
                  :icon="faTrash"
                  class="link"
                />
              </BcButtonIcon>
            </div>
          </template>
        </Column>
        <template #expansion="slotProps">
          <div class="expansion">
            <div class="info">
              <div class="label">
                {{ $t("notifications.col.subscriptions") }}
              </div>

              <BcTablePopoutEdit
                class="value"
                :label="slotProps.data.subscriptions.join(', ')"
                @on-edit="onEdit('subscriptions', slotProps.data)"
              />
            </div>
            <div class="info">
              <div class="label">
                {{ $t("notifications.col.webhook") }}
              </div>

              <BcTablePopoutEdit
                class="value"
                :label="slotProps.data.settings.webhook_url"
                @on-edit="() => onEdit('webhook', slotProps.data)"
              />
            </div>
            <div class="info">
              <div class="label">
                {{ $t("notifications.col.networks") }}
              </div>

              <BcTablePopoutEdit
                class="value"
                :no-icon="!slotProps.data.is_account_dashboard"
              >
                <template #content>
                  <div class="newtork-row">
                    <BcNetworkSelector
                      :readonly-networks="slotProps.data.chain_ids"
                    />
                    &nbsp;
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

.type-icon {
  margin-right: var(--padding);
}

.newtork-row {
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
      @include utils.set-all-width(180px);
    }

    @media (max-width: 1260px) {
      @include utils.set-all-width(140px);
    }

    @media (max-width: 520px) {
      @include utils.set-all-width(130px);
    }
  }

  .webhook-col,
  .subscriptions-col {
    @include utils.set-all-width(340px);

    @media (max-width: 1300px) {
      @include utils.set-all-width(260px);
    }

    @media (max-width: 1200px) {
      @include utils.set-all-width(240px);
    }
  }

  .networks-col {
    @include utils.set-all-width(156px);
  }
}
</style>
