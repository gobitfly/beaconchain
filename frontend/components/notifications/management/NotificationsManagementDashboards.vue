<script lang="ts" setup>import {
  faTrash,
  faDesktop,
  faUser
} from '@fortawesome/pro-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

import { getGroupLabel } from '~/utils/dashboard/group'
import { API_PATH } from '~/types/customFetch'
import type { ApiPagingResponse, ApiErrorResponse } from '~/types/api/common'
import type { NotificationSettingsDashboardsTableRow, NotificationSettingsValidatorDashboard, NotificationSettingsAccountDashboard } from '~/types/api/notifications'
import type { DashboardType } from '~/types/dashboard'
import { useNotificationsManagementDashboards } from '~/composables/notifications/useNotificationsManagementDashboards'
import { useUserDashboardStore } from '~/stores/dashboard/useUserDashboardStore'
import { NotificationsManagementSubscriptionDialog } from '#components'

type AllOptions = NotificationSettingsValidatorDashboard & NotificationSettingsAccountDashboard

interface WrappedRow extends NotificationSettingsDashboardsTableRow {
  dashboard_type: DashboardType,
  dashboard_name: string,
  subscriptions: string[],
  identifier: string
}

interface SettingsWithContext {
  row: WrappedRow,
  settings: AllOptions
}

// #### CONFIGURATION RELATED TO THE SUBSCRIPTION DIALOGS ####

const KeysIndicatingASubscription : Array<keyof AllOptions> = [
  'is_validator_offline_subscribed', 'group_offline_threshold', 'is_attestations_missed_subscribed', 'is_block_proposal_subscribed',
  'is_upcoming_block_proposal_subscribed', 'is_sync_subscribed', 'is_withdrawal_processed_subscribed', 'is_slashed_subscribed', 'is_real_time_mode_enabled',
  'is_incoming_transactions_subscribed', 'is_outgoing_transactions_subscribed', 'is_erc20_token_transfers_subscribed', 'is_erc721_token_transfers_subscribed',
  'is_erc1155_token_transfers_subscribed', 'is_ignore_spam_transactions_enabled'
]
const TimeoutForSavingFailures = 2300 // ms. We cannot let the user close the dialog and later interrupt his/her new activities with "we lost your preferences half a minute ago, we hope you remember them and do not mind going back to that dialog"
const MinimumTimeBetweenAPIcalls = 700 // ms. Any change ends-up saved anyway, so we can prevent useless requests with a delay larger than usual.

// #### END OF CONFIGURATION RELATED TO THE SUBSCRIPTION DIALOGS ####

const { fetch, setTimeout } = useCustomFetch()
const toast = useBcToast()
const { t: $t } = useI18n()
const dialog = useDialog()
const { dashboardGroups, query, cursor, pageSize, isLoading, onSort, setCursor, setPageSize, setSearch } = useNotificationsManagementDashboards()
const { getDashboardLabel } = useUserDashboardStore()
const { groups } = useValidatorDashboardGroups()
const { width } = useWindowSize()

const debouncer = useDebounceValue<SettingsWithContext>({} as SettingsWithContext, MinimumTimeBetweenAPIcalls)
watch(debouncer.value as Ref<SettingsWithContext>, saveUserSettings)

const colsVisible = computed(() => {
  return {
    networks: width.value > 1101,
    webhook: width.value >= 945,
    subscriptions: width.value >= 725
  }
})

const groupNameLabel = (groupId?: number) => {
  return getGroupLabel($t, groupId, groups.value, 'Σ')
}

const wrappedDashboardGroups: ComputedRef<ApiPagingResponse<WrappedRow>|undefined> = computed(() => {
  if (!dashboardGroups.value) {
    return
  }
  return {
    paging: dashboardGroups.value.paging,
    data: dashboardGroups.value.data.map(d => ({
      ...d,
      dashboard_type: dashboardType(d),
      dashboard_name: getDashboardLabel(String(d.dashboard_id), dashboardType(d)),
      subscriptions: subscriptionList(d),
      identifier: `${dashboardType(d)}-${d.dashboard_id}-${d.group_id}`
    }))
  }

  function dashboardType (row: NotificationSettingsDashboardsTableRow) : DashboardType {
    return row.is_account_dashboard ? 'account' : 'validator'
  }

  function subscriptionList (row: NotificationSettingsDashboardsTableRow) : string[] {
    const result: string[] = []
    for (const key of KeysIndicatingASubscription) {
      if ((row.settings as AllOptions)[key]) {
        result.push($t('notifications.subscriptions.' + dashboardType(row) + 's.' + key + '.option'))
      }
    }
    return result
  }
})

const onEdit = (col: 'delete' | 'subscriptions' | 'webhook' | 'networks', row: WrappedRow) => {
  const dialogProps = {
    dashboardType: row.dashboard_type,
    initialSettings: row.settings,
    saveUserSettings: (settings: AllOptions) => debouncer.bounce({ row, settings }, true, true)
  }
  switch (col) {
    case 'subscriptions':
      dialog.open(NotificationsManagementSubscriptionDialog, { data: dialogProps })
      break
    case 'webhook':
      /* TODO: replace `WebhookDialog` with the name of Marcel's component
      dialog.open(WebhookDialog, { data }) */
      break
    case 'networks':
      alert('TODO: edit networks' + row.group_id)
      break
    case 'delete':
      alert('TODO: delete' + row.group_id)
      break
  }
}

async function saveUserSettings (settingsAndContext: SettingsWithContext) {
  let response: ApiErrorResponse | undefined
  try {
    response = await fetch<ApiErrorResponse>(API_PATH.SAVE_DASHBOARDS_SETTINGS, {
      method: 'PUT',
      signal: AbortSignal.timeout(TimeoutForSavingFailures),
      body: { ...settingsAndContext.row.settings, ...settingsAndContext.settings }
    }, {
      for: settingsAndContext.row.dashboard_type,
      dashboardKey: String(settingsAndContext.row.dashboard_id),
      groupId: String(settingsAndContext.row.group_id)
    })
  } catch {
    response = undefined
  }
  if (!response || response.error) {
    toast.showError({ summary: $t('notifications.subscriptions.error_title'), group: $t('notifications.subscriptions.error_group'), detail: $t('notifications.subscriptions.error_message') })
  }
}

function getTypeIcon (type: DashboardType) {
  if (type === 'validator') {
    return faDesktop
  }
  return faUser
}

</script>

<template>
  <div>
    <Teleport to="#notifications-management-search-placholder">
      <BcContentFilter :search-placeholder="$t('placeholder')" class="search" @filter-changed="setSearch" />
    </Teleport>

    <ClientOnly fallback-tag="span">
      <BcTable
        :data="wrappedDashboardGroups"
        data-key="identifier"
        :expandable="!colsVisible.networks"
        class="notifications-management-dashboard-table"
        :cursor="cursor"
        :page-size="pageSize"
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
              <FontAwesomeIcon :icon="getTypeIcon(slotProps.data.dashboard_type)" class="type-icon" />
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
            <BcTablePopoutEdit
              :truncate-text="true"
              :no-icon="!slotProps.data.is_account_dashboard"
              @on-edit="onEdit('networks', slotProps.data)"
            >
              <template #content>
                <BcNetworkSelector :readonly-networks="slotProps.data.chain_ids" />
                &nbsp;
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
                :label="slotProps.data.settings.webhook_url"
                @on-edit="() => onEdit('webhook', slotProps.data)"
              />
            </div>
            <div class="info">
              <div class="label">
                {{ $t('notifications.col.networks') }}
              </div>

              <BcTablePopoutEdit
                class="value"
                :no-icon="!slotProps.data.is_account_dashboard"
                @on-edit="onEdit('networks', slotProps.data)"
              >
                <template #content>
                  <div class="newtork-row">
                    <BcNetworkSelector :readonly-networks="slotProps.data.chain_ids" />
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
