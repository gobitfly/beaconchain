export type API_PATH =
  | 'AD_CONFIGURATIONs'
  | 'DASHBOARD_CL_DEPOSITS'
  | 'DASHBOARD_CL_DEPOSITS_TOTAL'
  | 'DASHBOARD_CREATE_ACCOUNT'
  | 'DASHBOARD_CREATE_VALIDATOR'
  | 'DASHBOARD_DELETE_ACCOUNT'
  | 'DASHBOARD_DELETE_VALIDATOR'
  | 'DASHBOARD_EL_DEPOSITS'
  | 'DASHBOARD_EL_DEPOSITS_TOTAL'
  | 'DASHBOARD_OVERVIEW'
  | 'DASHBOARD_RENAME_ACCOUNT'
  | 'DASHBOARD_RENAME_VALIDATOR'
  | 'DASHBOARD_SLOTVIZ'
  | 'DASHBOARD_SUMMARY'
  | 'DASHBOARD_SUMMARY_CHART'
  | 'DASHBOARD_SUMMARY_DETAILS'
  | 'DASHBOARD_VALIDATOR_BLOCKS'
  | 'DASHBOARD_VALIDATOR_CREATE_PUBLIC_ID'
  | 'DASHBOARD_VALIDATOR_EDIT_PUBLIC_ID'
  | 'DASHBOARD_VALIDATOR_EPOCH_DUTY'
  | 'DASHBOARD_VALIDATOR_GROUP_MODIFY'
  | 'DASHBOARD_VALIDATOR_GROUPS'
  | 'DASHBOARD_VALIDATOR_INDICES'
  | 'DASHBOARD_VALIDATOR_MANAGEMENT'
  | 'DASHBOARD_VALIDATOR_MANAGEMENT_DELETE'
  | 'DASHBOARD_VALIDATOR_REWARDS'
  | 'DASHBOARD_VALIDATOR_REWARDS_CHART'
  | 'DASHBOARD_VALIDATOR_REWARDS_DETAILS'
  | 'DASHBOARD_VALIDATOR_TOTAL_WITHDRAWALS'
  | 'DASHBOARD_VALIDATOR_WITHDRAWALS'
  | 'GET_NOTIFICATIONS_SETTINGS_DASHBOARD'
  | 'LATEST_STATE'
  | 'LOGIN'
  | 'LOGOUT'
  | 'NOTIFICATIONS_CLIENTS'
  | 'NOTIFICATIONS_DASHBOARDS'
  | 'NOTIFICATIONS_DASHBOARDS_DETAILS_ACCOUNT'
  | 'NOTIFICATIONS_DASHBOARDS_DETAILS_VALIDATOR'
  | 'NOTIFICATIONS_MACHINE'
  | 'NOTIFICATIONS_MANAGEMENT_CLIENTS_SET_NOTIFICATION'
  | 'NOTIFICATIONS_MANAGEMENT_DASHBOARD_ACCOUNT_SET_NOTIFICATION'
  | 'NOTIFICATIONS_MANAGEMENT_DASHBOARD_VALIDATOR_SET_NOTIFICATION'
  | 'NOTIFICATIONS_MANAGEMENT_GENERAL'
  | 'NOTIFICATIONS_MANAGEMENT_NETWORK_SET_NOTIFICATION'
  | 'NOTIFICATIONS_MANAGEMENT_PAIRED_DEVICES_DELETE'
  | 'NOTIFICATIONS_MANAGEMENT_PAIRED_DEVICES_SET_NOTIFICATION'
  | 'NOTIFICATIONS_MANAGEMENT_SAVE'
  | 'NOTIFICATIONS_NETWORK'
  | 'NOTIFICATIONS_OVERVIEW'
  | 'NOTIFICATIONS_TEST_EMAIL'
  | 'NOTIFICATIONS_TEST_PUSH'
  | 'NOTIFICATIONS_TEST_WEBHOOK'
  | 'PRODUCT_SUMMARY'
  | 'REGISTER'
  | 'SAVE_VALIDATOR_DASHBOARDS_SETTINGS'
  | 'SEARCH'
  | 'STRIPE_CHECKOUT_SESSION'
  | 'STRIPE_CUSTOMER_PORTAL'
  | 'USER'
  | 'USER_CHANGE_EMAIL'
  | 'USER_CHANGE_PASSWORD'
  | 'USER_DASHBOARDS'
  | 'USER_DELETE'

export type PathValues = Record<string, boolean | number | string>

type MappingData = {
  getPath?: (values?: PathValues) => string,
  legacy?: boolean,
  method?: 'DELETE' | 'GET' | 'POST' | 'PUT', // 'GET' will be used as default
  mock?: boolean,
  mockFunction?: MockFunction,
  path: string,
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
    path: 'validator-dashboards/{dashboard_id}/groups',
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
    mock: false,
    mockFunction: mockLatestState,
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
  NOTIFICATIONS_DASHBOARDS_DETAILS_ACCOUNT: {
    getPath: pathValues =>
      `/users/me/notifications/account-dashboards/${pathValues?.dashboard_id}`
      + `/groups/${pathValues?.group_id}/epochs/${pathValues?.epoch}`,
    path: '/users/me/notifications/account-dashboards/{dashboard_id}/groups/{group_id}/epochs/{epoch}',
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
