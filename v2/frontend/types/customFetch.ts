import {
  mockManageNotificationsGeneral,
  simulateAPIresponseAboutNetworkList,
  simulateAPIresponseForTheSearchBar,
} from '~/utils/mock'

export enum API_PATH {
  AD_CONFIGURATIONs = '/adConfigurations',
  AVAILABLE_NETWORKS = '/availableNetworks',
  DASHBOARD_CL_DEPOSITS = '/dashboard/clDeposits',
  DASHBOARD_CL_DEPOSITS_TOTAL = '/dashboard/clDepositsTotal',
  DASHBOARD_CREATE_ACCOUNT = '/dashboard/createAccount',
  DASHBOARD_CREATE_VALIDATOR = '/dashboard/createValidator',
  DASHBOARD_DELETE_ACCOUNT = '/dashboard/deleteAccountDashbaoard',
  DASHBOARD_DELETE_VALIDATOR = '/dashboard/deleteValidatorDashboard',
  DASHBOARD_EL_DEPOSITS = '/dashboard/elDeposits',
  DASHBOARD_EL_DEPOSITS_TOTAL = '/dashboard/elDepositsTotal',
  DASHBOARD_OVERVIEW = '/dashboard/overview',
  DASHBOARD_RENAME_ACCOUNT = '/dashboard/renameAccountDashbaoard',
  DASHBOARD_RENAME_VALIDATOR = '/dashboard/renameValidatorDashboard',
  DASHBOARD_SLOTVIZ = '/dashboard/slotViz',
  DASHBOARD_SUMMARY = '/dashboard/validatorSummary',
  DASHBOARD_SUMMARY_CHART = '/dashboard/validatorSummaryChart',
  DASHBOARD_SUMMARY_DETAILS = '/dashboard/validatorSummaryDetails',
  DASHBOARD_VALIDATOR_BLOCKS = '/validator-dashboards/blocks',
  DASHBOARD_VALIDATOR_CREATE_PUBLIC_ID = '/validator-dashboards/publicIds',
  DASHBOARD_VALIDATOR_EDIT_PUBLIC_ID = '/validator-dashboards/editPublicIds',
  DASHBOARD_VALIDATOR_EPOCH_DUTY = '/validator-dashboards/epoch_duty',
  DASHBOARD_VALIDATOR_GROUP_MODIFY = '/validator-dashboards/group-modify',
  DASHBOARD_VALIDATOR_GROUPS = '/validator-dashboards/groups',
  DASHBOARD_VALIDATOR_INDICES = '/validator-dashboards/indices',
  DASHBOARD_VALIDATOR_MANAGEMENT = '/validator-dashboards/validators',
  DASHBOARD_VALIDATOR_REWARDS = '/dashboard/validatorRewards',
  DASHBOARD_VALIDATOR_REWARDS_CHART = '/dashboard/validatorRewardsChart',
  DASHBOARD_VALIDATOR_REWARDS_DETAILS = '/dashboard/validatorRewardsDetails',
  DASHBOARD_VALIDATOR_TOTAL_WITHDRAWALS = '/validator-dashboards/total-withdrawals',
  DASHBOARD_VALIDATOR_WITHDRAWALS = '/validator-dashboards/withdrawals',
  GET_NOTIFICATIONS_SETTINGS_DASHBOARD = '/notifications/managementDashboard',
  LATEST_STATE = '/latestState',
  LOGIN = '/login',
  LOGOUT = '/logout',
  NOTIFICATIONS_DASHBOARDS = '/notifications/dashboards',
  NOTIFICATIONS_MACHINE = '/notifications/machines',
  NOTIFICATIONS_MANAGEMENT_GENERAL = '/notifications/managementGeneral',
  NOTIFICATIONS_NETWORK = '/notifications/networks',
  NOTIFICATIONS_TEST_EMAIL = '/notifications/test_email',
  NOTIFICATIONS_TEST_PUSH = '/notifications/test_push',
  NOTIFICATIONS_TEST_WEBHOOK = '/users/me/notifications/test-webhook',
  PRODUCT_SUMMARY = '/productSummary',
  REGISTER = '/register',
  SAVE_DASHBOARDS_SETTINGS = '/settings-dashboards',
  SEARCH = '/search',
  STRIPE_CHECKOUT_SESSION = '/stripe/checkout-session',
  STRIPE_CUSTOMER_PORTAL = '/stripe/customer-portal',
  USER = '/user/me',
  USER_CHANGE_EMAIL = '/user/changeEmail',
  USER_CHANGE_PASSWORD = '/user/changePassword',
  USER_DASHBOARDS = '/user/dashboards',
  USER_DELETE = '/user/delete',
}

export type PathValues = Record<string, number | string>

interface MockFunction {
  (body?: any, param?: PathValues, query?: PathValues): any,
}

type MappingData = {
  getPath?: (values?: PathValues) => string,
  legacy?: boolean,
  method?: 'DELETE' | 'GET' | 'POST' | 'PUT', // 'GET' will be used as default
  mock?: boolean,
  mockFunction?: MockFunction,
  path: string,
}

export const mapping: Record<string, MappingData> = {
  [API_PATH.AD_CONFIGURATIONs]: {
    getPath: values => `/ad-configurations?keys=${values?.keys}`,
    mock: true,
    path: '/ad-configurations?={keys}',
  },
  [API_PATH.AVAILABLE_NETWORKS]: {
    method: 'GET',
    mock: true,
    mockFunction: simulateAPIresponseAboutNetworkList,
    path: '/available-networks',
  },
  [API_PATH.DASHBOARD_CL_DEPOSITS]: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/consensus-layer-deposits`,
    mock: false,
    path: '/validator-dashboards/{dashboard_id}/consensus-layer-deposits',
  },
  [API_PATH.DASHBOARD_CL_DEPOSITS_TOTAL]: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/total-consensus-layer-deposits`,
    mock: false,
    path: '/validator-dashboards/{dashboard_id}/total-consensus-layer-deposits',
  },
  [API_PATH.DASHBOARD_CREATE_ACCOUNT]: {
    method: 'POST',
    mock: true,
    path: '/account-dashboards',
  },
  [API_PATH.DASHBOARD_CREATE_VALIDATOR]: {
    method: 'POST',
    mock: false,
    path: '/validator-dashboards',
  },
  [API_PATH.DASHBOARD_DELETE_ACCOUNT]: {
    getPath: values => `/account-dashboards/${values?.dashboardKey}`,
    method: 'DELETE',
    mock: true,
    path: '/account-dashboards/{dashboardKey}',
  },
  [API_PATH.DASHBOARD_DELETE_VALIDATOR]: {
    getPath: values => `/validator-dashboards/${values?.dashboardKey}`,
    method: 'POST',
    mock: false,
    path: '/validator-dashboards/{dashboard_id}/validators/bulk-deletions',
  },
  [API_PATH.DASHBOARD_EL_DEPOSITS]: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/execution-layer-deposits`,
    mock: false,
    path: '/validator-dashboards/{dashboard_id}/execution-layer-deposits',
  },
  [API_PATH.DASHBOARD_EL_DEPOSITS_TOTAL]: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/total-execution-layer-deposits`,
    mock: false,
    path: '/validator-dashboards/{dashboard_id}/total-execution-layer-deposits',
  },
  [API_PATH.DASHBOARD_OVERVIEW]: {
    getPath: values => `/validator-dashboards/${values?.dashboardKey}`,
    mock: false,
    path: '/validator-dashboards/{dashboardKey}',
  },
  [API_PATH.DASHBOARD_RENAME_ACCOUNT]: {
    getPath: values => `/account-dashboards/${values?.dashboardKey}/name`,
    method: 'PUT',
    mock: true,
    path: '/account-dashboards/{dashboardKey}/name',
  },
  [API_PATH.DASHBOARD_RENAME_VALIDATOR]: {
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/name`,
    method: 'PUT',
    mock: false,
    path: '/validator-dashboards/{dashboardKey}/name',
  },
  [API_PATH.DASHBOARD_SLOTVIZ]: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/slot-viz`,
    mock: false,
    path: '/validator-dashboards/{dashboardKey}/slot-viz',
  },
  [API_PATH.DASHBOARD_SUMMARY]: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/summary`,
    mock: false,
    path: '/validator-dashboards/{dashboardKey}/summary',
  },
  [API_PATH.DASHBOARD_SUMMARY_CHART]: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/summary-chart`,
    mock: false,
    path: '/validator-dashboards/{dashboardKey}/summary-chart?',
  },
  [API_PATH.DASHBOARD_SUMMARY_DETAILS]: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/groups/${values?.groupId}/summary`,
    mock: false,
    path: '/validator-dashboards/{dashboardKey}/groups/{group_id}/summary',
  },
  [API_PATH.DASHBOARD_VALIDATOR_BLOCKS]: {
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/blocks`,
    mock: false,
    path: 'validator-dashboards/{dashboard_id}/blocks',
  },
  [API_PATH.DASHBOARD_VALIDATOR_CREATE_PUBLIC_ID]: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/public-ids`,
    method: 'POST',
    mock: false,
    path: '/validator-dashboards/{dashboardKey}/public-ids',
  },
  [API_PATH.DASHBOARD_VALIDATOR_EDIT_PUBLIC_ID]: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/public-ids/${values?.publicId}`,
    method: 'PUT',
    mock: false,
    path: '/validator-dashboards/{dashboardKey}/public-ids/{publicId}',
  },
  [API_PATH.DASHBOARD_VALIDATOR_EPOCH_DUTY]: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/duties/${values?.epoch}`,
    mock: false,
    path: '/validator-dashboards/{dashboard_id}/duties/{epoch}:',
  },
  [API_PATH.DASHBOARD_VALIDATOR_GROUP_MODIFY]: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/groups/${values?.groupId}`,
    method: 'PUT', // can be 'DELETE' = delete group or 'PUT' = modify group
    mock: false,
    path: 'validator-dashboards/{dashboard_id}/groups/{group_id}',
  },
  [API_PATH.DASHBOARD_VALIDATOR_GROUPS]: {
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/groups`,
    method: 'POST',
    mock: false,
    path: 'validator-dashboards/{dashboard_id}/groups',
  },
  [API_PATH.DASHBOARD_VALIDATOR_INDICES]: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/summary/validators`,
    mock: false,
    path: '/validator-dashboards/{dashboard_id}/summary/validators',
  },
  [API_PATH.DASHBOARD_VALIDATOR_MANAGEMENT]: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/validators`,
    mock: false,
    path: 'validator-dashboards/{dashboard_id}/validators/bulk-deletions',
  },
  [API_PATH.DASHBOARD_VALIDATOR_REWARDS]: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/rewards`,
    mock: false,
    path: '/validator-dashboards/{dashboardKey}/rewards',
  },
  [API_PATH.DASHBOARD_VALIDATOR_REWARDS_CHART]: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/rewards-chart`,
    mock: false,
    path: '/validator-dashboards/{dashboardKey}/rewards-chart',
  },
  [API_PATH.DASHBOARD_VALIDATOR_REWARDS_DETAILS]: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/groups/${values?.groupId}/rewards/${values?.epoch}`,
    mock: false,
    path: '/validator-dashboards/{dashboardKey}/groups/{group_id}/rewards',
  },
  [API_PATH.DASHBOARD_VALIDATOR_TOTAL_WITHDRAWALS]: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/total-withdrawals`,
    mock: false,
    path: 'validator-dashboards/{dashboard_id}/total-withdrawals',
  },
  [API_PATH.DASHBOARD_VALIDATOR_WITHDRAWALS]: {
    getPath: values =>
      `/validator-dashboards/${values?.dashboardKey}/withdrawals`,
    mock: false,
    path: 'validator-dashboards/{dashboard_id}/withdrawals',
  },
  [API_PATH.GET_NOTIFICATIONS_SETTINGS_DASHBOARD]: {
    mock: false,
    path: '/users/me/notifications/settings/dashboards',
  },
  [API_PATH.LATEST_STATE]: {
    mock: false,
    mockFunction: mockLatestState,
    path: '/latest-state',
  },
  [API_PATH.LOGIN]: {
    method: 'POST',
    mock: false,
    path: '/login',
  },
  [API_PATH.LOGOUT]: {
    method: 'POST',
    mock: false,
    path: '/logout',
  },
  [API_PATH.NOTIFICATIONS_DASHBOARDS]: {
    path: '/users/me/notifications/dashboards',
  },
  [API_PATH.NOTIFICATIONS_MACHINE]: {
    path: '/users/me/notifications/machines',
  },
  [API_PATH.NOTIFICATIONS_MANAGEMENT_GENERAL]: {
    mock: true,
    mockFunction: mockManageNotificationsGeneral,
    path: '/users/me/notifications/settings/general',
  },
  [API_PATH.NOTIFICATIONS_NETWORK]: {
    path: '/users/me/notifications/networks',
  },
  [API_PATH.NOTIFICATIONS_TEST_EMAIL]: {
    method: 'POST',
    mock: true,
    path: '/users/me/notifications/test-emails',
  },
  [API_PATH.NOTIFICATIONS_TEST_PUSH]: {
    method: 'POST',
    mock: true,
    path: '/users/me/notifications/test-push',
  },
  [API_PATH.NOTIFICATIONS_TEST_WEBHOOK]: {
    method: 'POST',
    mock: false,
    path: '/users/me/notifications/test-webhook',
  },
  [API_PATH.PRODUCT_SUMMARY]: {
    mock: false,
    path: '/product-summary',
  },
  [API_PATH.REGISTER]: {
    method: 'POST',
    mock: true,
    path: '/users',
  },
  [API_PATH.SAVE_DASHBOARDS_SETTINGS]: {
    getPath: values =>
      `/users/me/notifications/settings/${values?.for}-dashboards/${values?.dashboardKey}/groups/${values?.groupId}`,
    method: 'POST',
    mock: false,
    path: '/users/me/notifications/settings/{for}-dashboards/{dashboard_key}/groups/{group_id}',
  },
  [API_PATH.SEARCH]: {
    method: 'POST',
    mock: false,
    mockFunction: simulateAPIresponseForTheSearchBar,
    path: '/search',
  },
  [API_PATH.STRIPE_CHECKOUT_SESSION]: {
    method: 'POST',
    mock: false,
    path: '/user/stripe/create-checkout-session',
  },
  [API_PATH.STRIPE_CUSTOMER_PORTAL]: {
    method: 'POST',
    mock: false,
    path: '/user/stripe/customer-portal',
  },
  [API_PATH.USER]: {
    mock: false,
    path: '/users/me',
  },
  [API_PATH.USER_CHANGE_EMAIL]: {
    method: 'PUT',
    mock: true,
    path: '/users/me/email',
  },
  [API_PATH.USER_CHANGE_PASSWORD]: {
    method: 'PUT',
    mock: true,
    path: '/users/me/password',
  },
  [API_PATH.USER_DASHBOARDS]: {
    mock: false,
    path: '/users/me/dashboards',
  },
  [API_PATH.USER_DELETE]: {
    method: 'DELETE',
    mock: true,
    path: '/users/me',
  },
}
