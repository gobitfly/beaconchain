import type {
  AdConfiguration,
} from './adConfiguration'
import type {
  GetUserDashboardsResponse,
} from './api/dashboard'
import type {
  InternalGetLatestStateResponse,
} from './api/latest_state'
import type {
  GetUserNotificationClientsResponse,
  GetUserNotificationDashboardsResponse,
  GetUserNotificationMachinesResponse,
  GetUserNotificationNetworksResponse,
  GetUserNotificationSettingsDashboardsResponse,
  GetUserNotificationSettingsResponse,
  GetUserNotificationsResponse,
  GetUserNotificationsValidatorDashboardResponse,
  PutUserNotificationSettingsAccountDashboardResponse,
  PutUserNotificationSettingsGeneralResponse,
  PutUserNotificationSettingsNetworksResponse,
  PutUserNotificationSettingsPairedDevicesResponse,
  PutUserNotificationSettingsValidatorDashboardResponse,

} from './api/notifications'
import type {
  InternalPostSearchResponse,
} from './api/search'
import type {
  GetValidatorDashboardSlotVizResponse,
} from './api/slot_viz'
import type {
  InternalGetProductSummaryResponse,
  InternalGetUserInfoResponse,
  StripeCreateCheckoutSession,
  StripeCustomerPortal,
} from './api/user'
import type {
  GetValidatorDashboardBlocksResponse,
  GetValidatorDashboardConsensusLayerDepositsResponse,
  GetValidatorDashboardDutiesResponse,
  GetValidatorDashboardExecutionLayerDepositsResponse,
  GetValidatorDashboardGroupRewardsResponse,
  GetValidatorDashboardGroupSummaryResponse,
  GetValidatorDashboardResponse,
  GetValidatorDashboardRewardsChartResponse,
  GetValidatorDashboardRewardsResponse,
  GetValidatorDashboardSummaryChartResponse,
  GetValidatorDashboardSummaryResponse,
  GetValidatorDashboardSummaryValidatorsResponse,
  GetValidatorDashboardTotalConsensusDepositsResponse,
  GetValidatorDashboardTotalExecutionDepositsResponse,
  GetValidatorDashboardTotalWithdrawalsResponse,
  GetValidatorDashboardValidatorsResponse,
  GetValidatorDashboardWithdrawalsResponse,
  VDBPostReturnData,
} from './api/validator_dashboard'

export type API_PATH = keyof API_PATH_RESPONSE

export type API_PATH_RESPONSE = {
  AD_CONFIGURATIONs: AdConfiguration,
  DASHBOARD_CL_DEPOSITS: GetValidatorDashboardConsensusLayerDepositsResponse,
  DASHBOARD_CL_DEPOSITS_TOTAL: GetValidatorDashboardTotalConsensusDepositsResponse,
  DASHBOARD_CREATE_ACCOUNT: { data: VDBPostReturnData },
  DASHBOARD_CREATE_VALIDATOR: { data: VDBPostReturnData },
  DASHBOARD_DELETE_ACCOUNT: unknown,
  DASHBOARD_DELETE_VALIDATOR: unknown,
  DASHBOARD_EL_DEPOSITS: GetValidatorDashboardExecutionLayerDepositsResponse,
  DASHBOARD_EL_DEPOSITS_TOTAL: GetValidatorDashboardTotalExecutionDepositsResponse,
  DASHBOARD_OVERVIEW: GetValidatorDashboardResponse,
  DASHBOARD_RENAME_ACCOUNT: unknown,
  DASHBOARD_RENAME_VALIDATOR: unknown,
  DASHBOARD_SLOTVIZ: GetValidatorDashboardSlotVizResponse,
  DASHBOARD_SUMMARY: GetValidatorDashboardSummaryResponse,
  DASHBOARD_SUMMARY_CHART: GetValidatorDashboardSummaryChartResponse,
  DASHBOARD_SUMMARY_DETAILS: GetValidatorDashboardGroupSummaryResponse,
  DASHBOARD_VALIDATOR_BLOCKS: GetValidatorDashboardBlocksResponse,
  DASHBOARD_VALIDATOR_CREATE_PUBLIC_ID: unknown,
  DASHBOARD_VALIDATOR_EDIT_PUBLIC_ID: unknown,
  DASHBOARD_VALIDATOR_EPOCH_DUTY: GetValidatorDashboardDutiesResponse,
  DASHBOARD_VALIDATOR_GROUP_MODIFY: unknown,
  DASHBOARD_VALIDATOR_GROUPS: unknown,
  DASHBOARD_VALIDATOR_INDICES: GetValidatorDashboardSummaryValidatorsResponse,
  DASHBOARD_VALIDATOR_MANAGEMENT: GetValidatorDashboardValidatorsResponse,
  DASHBOARD_VALIDATOR_MANAGEMENT_DELETE: () => void,
  DASHBOARD_VALIDATOR_REWARDS: GetValidatorDashboardRewardsResponse,
  DASHBOARD_VALIDATOR_REWARDS_CHART: GetValidatorDashboardRewardsChartResponse,
  DASHBOARD_VALIDATOR_REWARDS_DETAILS: GetValidatorDashboardGroupRewardsResponse,
  DASHBOARD_VALIDATOR_TOTAL_WITHDRAWALS: GetValidatorDashboardTotalWithdrawalsResponse,
  DASHBOARD_VALIDATOR_WITHDRAWALS: GetValidatorDashboardWithdrawalsResponse,
  GET_NOTIFICATIONS_SETTINGS_DASHBOARD: GetUserNotificationSettingsDashboardsResponse,
  LATEST_STATE: InternalGetLatestStateResponse,
  LOGIN: unknown,
  LOGOUT: unknown,
  NOTIFICATIONS_CLIENTS: GetUserNotificationClientsResponse,
  NOTIFICATIONS_DASHBOARDS: GetUserNotificationDashboardsResponse,
  NOTIFICATIONS_DASHBOARDS_DETAILS_VALIDATOR: GetUserNotificationsValidatorDashboardResponse,
  NOTIFICATIONS_MACHINE: GetUserNotificationMachinesResponse,
  NOTIFICATIONS_MANAGEMENT_CLIENTS_SET_NOTIFICATION: PutUserNotificationSettingsNetworksResponse,
  NOTIFICATIONS_MANAGEMENT_DASHBOARD_ACCOUNT_SET_NOTIFICATION: PutUserNotificationSettingsAccountDashboardResponse,
  NOTIFICATIONS_MANAGEMENT_DASHBOARD_VALIDATOR_SET_NOTIFICATION: PutUserNotificationSettingsValidatorDashboardResponse,
  NOTIFICATIONS_MANAGEMENT_GENERAL: GetUserNotificationSettingsResponse,
  NOTIFICATIONS_MANAGEMENT_NETWORK_SET_NOTIFICATION: PutUserNotificationSettingsNetworksResponse,
  NOTIFICATIONS_MANAGEMENT_PAIRED_DEVICES_DELETE: unknown,
  NOTIFICATIONS_MANAGEMENT_PAIRED_DEVICES_SET_NOTIFICATION: PutUserNotificationSettingsPairedDevicesResponse,
  NOTIFICATIONS_MANAGEMENT_SAVE: PutUserNotificationSettingsGeneralResponse,
  NOTIFICATIONS_NETWORK: GetUserNotificationNetworksResponse,
  NOTIFICATIONS_OVERVIEW: GetUserNotificationsResponse,
  NOTIFICATIONS_TEST_EMAIL: unknown,
  NOTIFICATIONS_TEST_PUSH: unknown,
  NOTIFICATIONS_TEST_WEBHOOK: unknown,
  PRODUCT_SUMMARY: InternalGetProductSummaryResponse,
  REGISTER: unknown,
  SAVE_VALIDATOR_DASHBOARDS_SETTINGS: PutUserNotificationSettingsValidatorDashboardResponse,
  SEARCH: InternalPostSearchResponse,
  STRIPE_CHECKOUT_SESSION: StripeCreateCheckoutSession,
  STRIPE_CUSTOMER_PORTAL: StripeCustomerPortal,
  USER: InternalGetUserInfoResponse,
  USER_CHANGE_EMAIL: unknown,
  USER_CHANGE_PASSWORD: unknown,
  USER_DASHBOARDS: GetUserDashboardsResponse,
  USER_DELETE: unknown,
}

export type PathValues = Record<string, boolean | number | string>

type MappingData = {
  getPath?: (values?: PathValues) => string,
  legacy?: boolean,
  method?: 'DELETE' | 'GET' | 'POST' | 'PUT', // 'GET' will be used as default
  mock?: boolean,
  mockFunction?: MockFunction,
  path: string,
  retunType?: any,
}

interface MockFunction {
  (body?: any, param?: PathValues, query?: PathValues): any,
}

export const mapping: Record<API_PATH, MappingData> = {
  AD_CONFIGURATIONs: {
    getPath: values => `/ad-configurations?keys=${values?.keys}`,
    mock: true,
    path: '/ad-configurations?={keys}',
  },
  DASHBOARD_CL_DEPOSITS: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/consensus-layer-deposits`,
    mock: false,
    path: '/validator-dashboards/{dashboard_id}/consensus-layer-deposits',
  },
  DASHBOARD_CL_DEPOSITS_TOTAL: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/total-consensus-layer-deposits`,
    mock: false,
    path: '/validator-dashboards/{dashboard_id}/total-consensus-layer-deposits',
  },
  DASHBOARD_CREATE_ACCOUNT: {
    method: 'POST',
    mock: true,
    path: '/account-dashboards',
  },
  DASHBOARD_CREATE_VALIDATOR: {
    method: 'POST',
    mock: false,
    path: '/validator-dashboards',
  },
  DASHBOARD_DELETE_ACCOUNT: {
    getPath: values => `/account-dashboards/${values?.dashboardKey}`,
    method: 'DELETE',
    mock: true,
    path: '/account-dashboards/{dashboardKey}',
  },
  DASHBOARD_DELETE_VALIDATOR: {
    getPath: values => `/validator-dashboards/${values?.dashboardKey}`,
    method: 'DELETE',
    mock: false,
    path: '/validator-dashboards/{dashboardKey}',
  },
  DASHBOARD_EL_DEPOSITS: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/execution-layer-deposits`,
    mock: false,
    path: '/validator-dashboards/{dashboard_id}/execution-layer-deposits',
  },
  DASHBOARD_EL_DEPOSITS_TOTAL: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/total-execution-layer-deposits`,
    mock: false,
    path: '/validator-dashboards/{dashboard_id}/total-execution-layer-deposits',
  },
  DASHBOARD_OVERVIEW: {
    getPath: values => `/validator-dashboards/${values?.dashboardKey}`,
    mock: false,
    path: '/validator-dashboards/{dashboardKey}',
  },
  DASHBOARD_RENAME_ACCOUNT: {
    getPath: values => `/account-dashboards/${values?.dashboardKey}/name`,
    method: 'PUT',
    mock: true,
    path: '/account-dashboards/{dashboardKey}/name',
  },
  DASHBOARD_RENAME_VALIDATOR: {
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/name`,
    method: 'PUT',
    mock: false,
    path: '/validator-dashboards/{dashboardKey}/name',
  },
  DASHBOARD_SLOTVIZ: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/slot-viz`,
    mock: false,
    path: '/validator-dashboards/{dashboardKey}/slot-viz',
  },
  DASHBOARD_SUMMARY: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/summary`,
    mock: false,
    path: '/validator-dashboards/{dashboardKey}/summary',
  },
  DASHBOARD_SUMMARY_CHART: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/summary-chart`,
    mock: false,
    path: '/validator-dashboards/{dashboardKey}/summary-chart?',
  },
  DASHBOARD_SUMMARY_DETAILS: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/groups/${values?.groupId}/summary`,
    mock: false,
    path: '/validator-dashboards/{dashboardKey}/groups/{group_id}/summary',
  },
  DASHBOARD_VALIDATOR_BLOCKS: {
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/blocks`,
    mock: false,
    path: 'validator-dashboards/{dashboard_id}/blocks',
  },
  DASHBOARD_VALIDATOR_CREATE_PUBLIC_ID: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/public-ids`,
    method: 'POST',
    mock: false,
    path: '/validator-dashboards/{dashboardKey}/public-ids',
  },
  DASHBOARD_VALIDATOR_EDIT_PUBLIC_ID: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/public-ids/${values?.publicId}`,
    method: 'PUT',
    mock: false,
    path: '/validator-dashboards/{dashboardKey}/public-ids/{publicId}',
  },
  DASHBOARD_VALIDATOR_EPOCH_DUTY: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/duties/${values?.epoch}`,
    mock: false,
    path: '/validator-dashboards/{dashboard_id}/duties/{epoch}:',
  },
  DASHBOARD_VALIDATOR_GROUP_MODIFY: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/groups/${values?.groupId}`,
    method: 'PUT', // can be 'DELETE' = delete group or 'PUT' = modify group
    mock: false,
    path: 'validator-dashboards/{dashboard_id}/groups/{group_id}',
  },
  DASHBOARD_VALIDATOR_GROUPS: {
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/groups`,
    method: 'POST',
    mock: false,
    path: ' ',
  },
  DASHBOARD_VALIDATOR_INDICES: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/summary/validators`,
    mock: false,
    path: '/validator-dashboards/{dashboard_id}/summary/validators',
  },
  DASHBOARD_VALIDATOR_MANAGEMENT: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/validators`,
    mock: false,
    path: 'validator-dashboards/{dashboard_id}/validators',
  },
  DASHBOARD_VALIDATOR_MANAGEMENT_DELETE: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/validators/bulk-deletions`,
    mock: false,
    path: 'validator-dashboards/{dashboard_id}/validators/bulk-deletions',
  },
  DASHBOARD_VALIDATOR_REWARDS: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/rewards`,
    mock: false,
    path: '/validator-dashboards/{dashboardKey}/rewards',
  },
  DASHBOARD_VALIDATOR_REWARDS_CHART: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/rewards-chart`,
    mock: false,
    path: '/validator-dashboards/{dashboardKey}/rewards-chart',
  },
  DASHBOARD_VALIDATOR_REWARDS_DETAILS: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/groups/${values?.groupId}/rewards/${values?.epoch}`,
    mock: false,
    path: '/validator-dashboards/{dashboardKey}/groups/{group_id}/rewards',
  },
  DASHBOARD_VALIDATOR_TOTAL_WITHDRAWALS: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/total-withdrawals`,
    mock: false,
    path: 'validator-dashboards/{dashboard_id}/total-withdrawals',
  },
  DASHBOARD_VALIDATOR_WITHDRAWALS: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/withdrawals`,
    mock: false,
    path: 'validator-dashboards/{dashboard_id}/withdrawals',
  },
  GET_NOTIFICATIONS_SETTINGS_DASHBOARD: {
    mock: false,
    path: '/users/me/notifications/settings/dashboards',
  },
  LATEST_STATE: {
    path: '/latest-state',
  },
  LOGIN: {
    method: 'POST',
    mock: false,
    path: '/login',
  },
  LOGOUT: {
    method: 'POST',
    mock: false,
    path: '/logout',
  },
  NOTIFICATIONS_CLIENTS: {
    method: 'GET',
    path: '/users/me/notifications/clients',
  },
  NOTIFICATIONS_DASHBOARDS: {
    path: '/users/me/notifications/dashboards',
  },
  NOTIFICATIONS_DASHBOARDS_DETAILS_VALIDATOR: {
    getPath: pathValues =>
      `/users/me/notifications/validator-dashboards/${pathValues?.dashboard_id}`
      + `/groups/${pathValues?.group_id}/epochs/${pathValues?.epoch}`,
    path: '/users/me/notifications/validator-dashboards/{dashboard_id}/groups/{group_id}/epochs/{epoch}',
  },
  NOTIFICATIONS_MACHINE: {
    path: '/users/me/notifications/machines',
  },
  NOTIFICATIONS_MANAGEMENT_CLIENTS_SET_NOTIFICATION: {
    getPath: pathValues =>
      `/users/me/notifications/settings/clients/${pathValues?.client_id}`,
    method: 'PUT',
    path: '/users/me/notifications/settings/clients/{client_id}',
  },
  NOTIFICATIONS_MANAGEMENT_DASHBOARD_ACCOUNT_SET_NOTIFICATION: {
    getPath: pathValues =>
      `/users/me/notifications/settings/account-dashboards/${pathValues?.dashboard_id}`
      + `/groups/${pathValues?.group_id}`,
    method: 'PUT',
    path: '/users/me/notifications/settings/account-dashboards/{dashboard_id}/groups/{group_id}',
  },
  NOTIFICATIONS_MANAGEMENT_DASHBOARD_VALIDATOR_SET_NOTIFICATION: {
    getPath: pathValues =>
      `/users/me/notifications/settings/validator-dashboards/${pathValues?.dashboard_id}`
      + `/groups/${pathValues?.group_id}`,
    method: 'PUT',
    path: '/users/me/notifications/settings/validator-dashboards/{dashboard_id}/groups/{group_id}',
  },
  NOTIFICATIONS_MANAGEMENT_GENERAL: {
    path: '/users/me/notifications/settings',
  },
  NOTIFICATIONS_MANAGEMENT_NETWORK_SET_NOTIFICATION: {
    getPath: pathValues =>
      `/users/me/notifications/settings/networks/${pathValues?.network}`,
    method: 'PUT',
    path: '/users/me/notifications/settings/networks/{network}',
  },
  NOTIFICATIONS_MANAGEMENT_PAIRED_DEVICES_DELETE: {
    getPath: pathValues =>
      `/users/me/notifications/settings/paired-devices/${pathValues?.paired_device_id}`,
    method: 'DELETE',
    path: '/users/me/notifications/settings/paired-devices/{paired_device_id}',
  },
  NOTIFICATIONS_MANAGEMENT_PAIRED_DEVICES_SET_NOTIFICATION: {
    getPath: pathValues =>
      `/users/me/notifications/settings/paired-devices/${pathValues?.paired_device_id}`,
    method: 'PUT',
    path: '/users/me/notifications/settings/paired-devices/{paired_device_id}',
  },
  NOTIFICATIONS_MANAGEMENT_SAVE: {
    method: 'PUT',
    path: '/users/me/notifications/settings/general',
  },
  NOTIFICATIONS_NETWORK: {
    path: '/users/me/notifications/networks',
  },
  NOTIFICATIONS_OVERVIEW: {
    method: 'GET',
    mock: false,
    path: '/users/me/notifications',
  },
  NOTIFICATIONS_TEST_EMAIL: {
    method: 'POST',
    path: '/users/me/notifications/test-email',
  },
  NOTIFICATIONS_TEST_PUSH: {
    method: 'POST',
    path: '/users/me/notifications/test-push',
  },
  NOTIFICATIONS_TEST_WEBHOOK: {
    method: 'POST',
    mock: false,
    path: '/users/me/notifications/test-webhook',
  },
  PRODUCT_SUMMARY: {
    mock: false,
    path: '/product-summary',
  },
  REGISTER: {
    method: 'POST',
    mock: true,
    path: '/users',
  },
  SAVE_VALIDATOR_DASHBOARDS_SETTINGS: {
    getPath: values =>
      `/users/me/notifications/settings/validator-dashboards/${values?.dashboard_id}/groups/${values?.group_id}`,
    method: 'POST',
    path: '/users/me/notifications/settings/validator-dashboards/{dashboard_id}/groups/{group_id}',
  },
  SEARCH: {
    method: 'POST',
    path: '/search',
  },
  STRIPE_CHECKOUT_SESSION: {
    method: 'POST',
    mock: false,
    path: '/user/stripe/create-checkout-session',
  },
  STRIPE_CUSTOMER_PORTAL: {
    method: 'POST',
    mock: false,
    path: '/user/stripe/customer-portal',
  },
  USER: {
    mock: false,
    path: '/users/me',
  },
  USER_CHANGE_EMAIL: {
    method: 'PUT',
    mock: true,
    path: '/users/me/email',
  },
  USER_CHANGE_PASSWORD: {
    method: 'PUT',
    mock: true,
    path: '/users/me/password',
  },
  USER_DASHBOARDS: {
    mock: false,
    path: '/users/me/dashboards',
  },
  USER_DELETE: {
    method: 'DELETE',
    mock: true,
    path: '/users/me',
  },
}
