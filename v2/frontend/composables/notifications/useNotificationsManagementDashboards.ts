import type { TableQueryParams } from '~/types/datatable'

import type {
  GetUserNotificationSettingsDashboardsResponse,
  NotificationSettingsAccountDashboard,
  NotificationSettingsDashboardsTableRow,
  NotificationSettingsValidatorDashboard,
  PutUserNotificationSettingsAccountDashboardResponse,
  PutUserNotificationSettingsValidatorDashboardResponse,
} from '~/types/api/notifications'

export function useNotificationsManagementDashboards() {
  const { fetch } = useCustomFetch()
  const { refreshOverview } = useNotificationsDashboardOverviewStore()
  const data = ref<GetUserNotificationSettingsDashboardsResponse>()
  const {
    cursor,
    isStoredQuery,
    onSort,
    pageSize,
    pendingQuery,
    query,
    setCursor,
    setPageSize,
    setSearch,
    setStoredQuery,
  } = useTableQuery({
    limit: 10,
    sort: 'dashboard_id:desc',
  }, 10)
  const isLoading = ref(false)

  const dashboards = computed(() => data.value)

  async function getDashboards(q?: TableQueryParams) {
    isLoading.value = true
    setStoredQuery(q)
    const res
      = await fetch<GetUserNotificationSettingsDashboardsResponse>(
        'GET_NOTIFICATIONS_SETTINGS_DASHBOARD',
        undefined,
        undefined,
        q,
      )

    isLoading.value = false
    if (!isStoredQuery(q)) {
      return // in case some query params change while loading
    }

    data.value = res
    return res
  }

  watch(
    query,
    (q) => {
      getDashboards(q)
    },
    { immediate: true },
  )
  const clearSettings = (
    {
      is_account_dashboard,
      settings,
    }:
    {
      is_account_dashboard: boolean,
      settings: NotificationSettingsAccountDashboard | NotificationSettingsValidatorDashboard,
    },
  ) => {
    settings.webhook_url = ''
    settings.is_webhook_discord_enabled = false
    if (is_account_dashboard) {
      const accountDashboardSettings = settings as NotificationSettingsAccountDashboard
      accountDashboardSettings.erc20_token_transfers_value_threshold = 0
      accountDashboardSettings.is_erc1155_token_transfers_subscribed = false
      accountDashboardSettings.is_erc20_token_transfers_subscribed = false
      accountDashboardSettings.is_erc721_token_transfers_subscribed = false
      accountDashboardSettings.is_ignore_spam_transactions_enabled = false
      accountDashboardSettings.is_incoming_transactions_subscribed = false
      accountDashboardSettings.is_outgoing_transactions_subscribed = false
      return
    }
    const accountDashboardSettings = settings as NotificationSettingsValidatorDashboard
    accountDashboardSettings.group_efficiency_below_threshold = 0
    accountDashboardSettings.is_attestations_missed_subscribed = false
    accountDashboardSettings.is_block_proposal_subscribed = false
    accountDashboardSettings.is_group_efficiency_below_subscribed = false
    accountDashboardSettings.is_max_collateral_subscribed = false
    accountDashboardSettings.is_min_collateral_subscribed = false
    accountDashboardSettings.is_slashed_subscribed = false
    accountDashboardSettings.is_sync_subscribed = false
    accountDashboardSettings.is_upcoming_block_proposal_subscribed = false
    accountDashboardSettings.is_validator_offline_subscribed = false
    accountDashboardSettings.is_withdrawal_processed_subscribed = false
    accountDashboardSettings.max_collateral_threshold = 0
    accountDashboardSettings.min_collateral_threshold = 0
  }
  const deleteDashboardNotifications = async (
    {
      dashboard_id,
      group_id,
      is_account_dashboard,
      settings,
    }:
    Pick<
      NotificationSettingsDashboardsTableRow,
      | 'dashboard_id'
      | 'group_id'
      | 'is_account_dashboard'
      | 'settings'
    >,
  ) => {
    clearSettings({
      is_account_dashboard,
      settings,
    })
    if (is_account_dashboard) {
      return await fetch<PutUserNotificationSettingsAccountDashboardResponse>(
        'NOTIFICATIONS_MANAGEMENT_DASHBOARD_ACCOUNT_SET_NOTIFICATION',
        {
          body: settings,
        },
        {
          dashboard_id,
          group_id,
        },
      )
    }
    return await fetch<PutUserNotificationSettingsValidatorDashboardResponse>(
      'NOTIFICATIONS_MANAGEMENT_DASHBOARD_VALIDATOR_SET_NOTIFICATION',
      {
        body: settings,
      },
      {
        dashboard_id,
        group_id,
      },
    )
  }

  function patchDashboardSettings(
    {
      dashboard_id,
      group_id,
      settings,
    }:
    {
      dashboard_id: number,
      group_id: number,
      settings: NotificationSettingsValidatorDashboard,
    },
  ) {
    const currentDashboard = dashboards.value?.data.find((dashboard) => {
      if (dashboard.dashboard_id === dashboard_id && dashboard.group_id === group_id) {
        return dashboard
      }
    })
    if (currentDashboard) {
      currentDashboard.settings = settings
    }
  }
  async function saveSubscriptions(
    {
      dashboard_id,
      group_id,
      settings,
    }:
    {
      dashboard_id: number,
      group_id: number,
      settings: NotificationSettingsValidatorDashboard,
    },
  ) {
    await fetch<PutUserNotificationSettingsValidatorDashboardResponse>(
      'SAVE_VALIDATOR_DASHBOARDS_SETTINGS',
      {
        body: {
          ...settings,
        },
        method: 'PUT',
      },
      {
        dashboard_id,
        group_id,
      },
    ).then(({ data: settings }) => {
      patchDashboardSettings({
        dashboard_id,
        group_id,
        settings,
      })
    }).then(() => refreshOverview())
  }

  return {
    cursor,
    dashboards,
    deleteDashboardNotifications,
    isLoading,
    onSort,
    pageSize,
    query: pendingQuery,
    saveSubscriptions,
    setCursor,
    setPageSize,
    setSearch,
  }
}
