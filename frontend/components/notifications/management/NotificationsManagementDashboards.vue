<script lang="ts" setup>
import type { ApiPagingResponse } from '~/types/api/common'
import type {
  NotificationSettingsDashboardsTableRow,
  NotificationSettingsValidatorDashboard,
} from '~/types/api/notifications'
import type { DashboardType } from '~/types/dashboard'
import {
  NotificationsManagementModalDashboardsDelete,
  NotificationsManagementModalWebhook,
  NotificationsManagementSubscriptionDialog,
} from '#components'
import type { WebhookSettings } from '~/components/notifications/management/modal/NotificationsManagementModalWebhook.vue'
import type { SubscriptionSettings } from '~/components/notifications/management/NotificationsManagementSubscriptionDialog.vue'

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

const { width } = useWindowSize()

const query = useDefaultQuery({
  limit: 10,
  sort: 'dashboard_id:desc',
})

const {
  data: dashboards,
  refresh: refreshNotificationDashboardSettings,
  status,
} = useApi('/api/bff/users/me/notifications/settings/dashboards', {
  query,
})

const colsVisible = computed(() => {
  return {
    networks: width.value > 1101,
    subscriptions: width.value >= 725,
    webhook: width.value >= 945,
  }
})

const wrappedDashboards: ComputedRef<
ApiPagingResponse<WrappedRow> | null>
   = computed(() => {
     if (!dashboards.value) {
       return null
     }
     return {
       data: dashboards.value.data.map(dashboard => ({
         ...dashboard,
         dashboard_type: 'validator',
         identifier: `validator-${dashboard.dashboard_id}-${dashboard.group_id}`,
         subscriptions: getSubscriptions(dashboard),
       })),
       paging: dashboards.value.paging,
     }

     function getSubscriptions(
       row: NotificationSettingsDashboardsTableRow,
     ): string[] {
       const result: string[] = []
       const settingsValidatorDashboard = row.settings as NotificationSettingsValidatorDashboard
       if (settingsValidatorDashboard.is_validator_offline_subscribed) {
         result.push($t('notifications.subscriptions.validators.validator_is_offline.label'))
       }
       if (settingsValidatorDashboard.is_attestations_missed_subscribed) {
         result.push($t('notifications.subscriptions.validators.attestation_missed.label'))
       }
       if (settingsValidatorDashboard.is_block_proposal_missed_subscribed) {
         result.push($t('notifications.subscriptions.validators.block_proposal_missed.label'))
       }
       if (settingsValidatorDashboard.is_block_proposal_success_subscribed) {
         result.push($t('notifications.subscriptions.validators.block_proposal_success.label'))
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

const { $api } = useNuxtApp()

const editNotificationSettings = async (
  dashboardId: number,
  groupId: number,
  settings: NotificationSettingsDashboardsTableRow['settings'],
) => {
  try {
    await $api(
      `/api/bff/users/me/notifications/settings/validator-dashboards/${dashboardId}/groups/${groupId}`,
      {
        body: settings,
        method: 'PUT',
      })
    refreshNotificationDashboardSettings()
    refreshOverview()
  }
  catch {
    toast.showError({
      detail: $t('notifications.dashboards.error_message'),
      summary: $t('notifications.dashboards.error_title'),
    })
  }
}

type Dialog = 'delete' | 'networks' | 'subscriptions' | 'webhook'
const onEdit = (col: Dialog, row: WrappedRow) => {
  switch (col) {
    case 'delete':
      return dialog.open(NotificationsManagementModalDashboardsDelete, {
        data: row,
        emits: {
          onDelete: async (
          ) => {
            await editNotificationSettings(
              row.dashboard_id,
              row.group_id,
              {
                group_efficiency_below_threshold: 0,
                is_attestations_missed_subscribed: false,
                is_block_proposal_missed_subscribed: false,
                is_block_proposal_success_subscribed: false,
                is_group_efficiency_below_subscribed: false,
                is_max_collateral_subscribed: false,
                is_min_collateral_subscribed: false,
                is_slashed_subscribed: false,
                is_sync_subscribed: false,
                is_upcoming_block_proposal_subscribed: false,
                is_validator_offline_subscribed: false,
                is_webhook_discord_enabled: false,
                is_withdrawal_processed_subscribed: false,
                max_collateral_threshold: 0,
                min_collateral_threshold: 0,
                webhook_url: '',
              })
          },
        },
      })
    case 'subscriptions':
      dialog.open(NotificationsManagementSubscriptionDialog, {
        data: row.settings,
        emits: {
          onChangeSettings: async (
            subscriptinSettings: SubscriptionSettings,
          ) => {
            await editNotificationSettings(
              row.dashboard_id,
              row.group_id,
              {
                ...subscriptinSettings,
                is_webhook_discord_enabled: row.settings.is_webhook_discord_enabled,
                webhook_url: row.settings.webhook_url,
              })
          },
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
            webhookSettings: WebhookSettings,
          ) => {
            await editNotificationSettings(
              row.dashboard_id,
              row.group_id,
              {
                ...row.settings,
                ...webhookSettings,
              })
          },
        },
      })
      break
  }
}

const isDeleteButtonDisabled = (dashboard: WrappedRow) => {
  const hasSubscriptions = dashboard.subscriptions?.length
  return !hasSubscriptions || dashboard.is_archived
}
</script>

<template>
  <BcTableControl
    v-model:search="query.search"
    :search-placeholder="$t('notifications.dashboards.search_placeholder')"
  >
    <template #table>
      <ClientOnly fallback-tag="span">
        <BcTable
          :data="wrappedDashboards"
          :query
          data-key="identifier"
          :expandable="!colsVisible.networks"
          class="notifications-management-dashboard-table"
          :is-loading="status === 'pending'"
        >
          <Column
            field="dashboard_id"
            body-class="dashboard-col"
            header-class="dashboard-col"
            sortable
            :header="$t('notifications.col.dashboard')"
          >
            <template #body="slotProps">
              <BcTooltip
                v-if="slotProps.data.is_archived"
                fit-content
                tooltip-text-align="left"
                class="disabled-text"
                :text="$t('notifications.dashboards.archived')"
              >
                <BcIcon
                  :name="slotProps.data.dashboard_type === 'validator' ? 'desktop' : 'user'"
                  class="type-icon"
                />
                {{ slotProps.data.dashboard_name }}
              </BcTooltip>
              <span
                v-else
              >
                <BcIcon
                  :name="slotProps.data.dashboard_type === 'validator' ? 'desktop' : 'user'"
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
            :header="$t('notifications.col.group')"
          >
            <template #body="slotProps">
              <span :class="{ 'text-disabled': slotProps.data.is_archived }">
                {{ slotProps.data.group_name }}
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
                :class="{ 'text-disabled': slotProps.data.is_archived }"
                :is-disabled="slotProps.data.is_archived"
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
                :class="{ 'text-disabled': slotProps.data.is_archived }"
                :is-disabled="slotProps.data.is_archived"
                :truncate-text="true"
                :label="slotProps.data.settings.webhook_url"
                @on-edit="() => onEdit('webhook', slotProps.data)"
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
                  :screenreader-text="{
                    key: 'notifications.clients.settings.screenreader.delete_notifications_for_dashboard_id',
                    interpolation: { dashboard_id: slotProps.data.dashboard_name },
                  }"
                  :disabled="isDeleteButtonDisabled(slotProps.data)"
                  class="link"
                  name="trash"
                  @click="onEdit('delete', slotProps.data)"
                />
              </div>
            </template>
          </Column>
          <template #expansion="slotProps">
            <div class="expansion">
              <div class="info">
                <div
                  class="label"
                  :class="{ 'text-disabled': slotProps.data.is_archived }"
                >
                  {{ $t("notifications.col.subscriptions") }}
                </div>

                <BcTablePopoutEdit
                  class="value"
                  :class="{ 'text-disabled': slotProps.data.is_archived }"
                  :is-disabled="slotProps.data.is_archived"
                  :label="slotProps.data.subscriptions.join(', ')"
                  @on-edit="onEdit('subscriptions', slotProps.data)"
                />
              </div>
              <div class="info">
                <div
                  class="label"
                  :class="{ 'text-disabled': slotProps.data.is_archived }"
                >
                  {{ $t("notifications.col.webhook") }}
                </div>

                <BcTablePopoutEdit
                  :class="{ 'text-disabled': slotProps.data.is_archived }"
                  :is-disabled="slotProps.data.is_archived"
                  class="value"
                  :label="slotProps.data.settings.webhook_url"
                  truncate-text
                  @on-edit="() => onEdit('webhook', slotProps.data)"
                />
              </div>
            </div>
          </template>
        </BcTable>
      </ClientOnly>
    </template>
  </BcTableControl>
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

.disabled-text {
  color: var(--text-color-disabled)
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

  .networks-col {
    @include utils.set-all-width(156px);
  }
}

:deep(.webhook-col),
:deep(.subscriptions-col) {
  @include utils.set-all-width(340px);

  @media (max-width: 1300px) {
    @include utils.set-all-width(260px);
  }

  @media (max-width: 1200px) {
    @include utils.set-all-width(240px);
  }
}
</style>
