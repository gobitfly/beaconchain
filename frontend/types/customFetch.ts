import { simulateAPIresponseForTheSearchBar, simulateAPIresponseAboutNetworkList } from '~/utils/mock'

export enum API_PATH {
  AD_CONFIGURATIONs = '/adConfigurations',
  USER = '/user/me',
  USER_DASHBOARDS = '/user/dashboards',
  DASHBOARD_CREATE_ACCOUNT = '/dashboard/createAccount',
  DASHBOARD_CREATE_VALIDATOR = '/dashboard/createValidator',
  DASHBOARD_DELETE_ACCOUNT = '/dashboard/deleteAccountDashbaoard',
  DASHBOARD_DELETE_VALIDATOR = '/dashboard/deleteValidatorDashboard',
  DASHBOARD_RENAME_ACCOUNT = '/dashboard/renameAccountDashbaoard',
  DASHBOARD_RENAME_VALIDATOR = '/dashboard/renameValidatorDashboard',
  DASHBOARD_VALIDATOR_MANAGEMENT = '/validator-dashboards/validators',
  DASHBOARD_VALIDATOR_GROUPS = '/validator-dashboards/groups',
  DASHBOARD_VALIDATOR_CREATE_PUBLIC_ID = '/validator-dashboards/publicIds',
  DASHBOARD_VALIDATOR_EDIT_PUBLIC_ID = '/validator-dashboards/editPublicIds',
  DASHBOARD_VALIDATOR_GROUP_MODIFY = '/validator-dashboards/group-modify',
  DASHBOARD_VALIDATOR_REWARDS_CHART = '/dashboard/validatorRewardsChart',
  DASHBOARD_VALIDATOR_BLOCKS = '/validator-dashboards/blocks',
  DASHBOARD_VALIDATOR_WITHDRAWALS = '/validator-dashboards/withdrawals',
  DASHBOARD_VALIDATOR_TOTAL_WITHDRAWALS = '/validator-dashboards/total-withdrawals',
  DASHBOARD_VALIDATOR_EPOCH_DUTY = '/validator-dashboards/epoch_duty',
  DASHBOARD_VALIDATOR_INDICES = '/validator-dashboards/indices',
  DASHBOARD_SUMMARY = '/dashboard/validatorSummary',
  DASHBOARD_SUMMARY_DETAILS = '/dashboard/validatorSummaryDetails',
  DASHBOARD_VALIDATOR_REWARDS = '/dashboard/validatorRewards',
  DASHBOARD_VALIDATOR_REWARDS_DETAILS = '/dashboard/validatorRewardsDetails',
  DASHBOARD_SUMMARY_CHART = '/dashboard/validatorSummaryChart',
  DASHBOARD_EL_DEPOSITS = '/dashboard/elDeposits',
  DASHBOARD_EL_DEPOSITS_TOTAL = '/dashboard/elDepositsTotal',
  DASHBOARD_CL_DEPOSITS = '/dashboard/clDeposits',
  DASHBOARD_CL_DEPOSITS_TOTAL = '/dashboard/clDepositsTotal',
  DASHBOARD_OVERVIEW = '/dashboard/overview',
  DASHBOARD_SLOTVIZ = '/dashboard/slotViz',
  LATEST_STATE = '/latestState',
  LOGIN = '/login',
  LOGOUT = '/logout',
  SEARCH = '/search',
  AVAILABLE_NETWORKS = '/availableNetworks',
  PRODUCT_SUMMARY = '/productSummary',
  STRIPE_CUSTOMER_PORTAL = '/stripe/customer-portal',
  STRIPE_CHECKOUT_SESSION = '/stripe/checkout-session'
}

export type PathValues = Record<string, string | number>

interface MockFunction {
  (body?: any, param?: PathValues, query?: PathValues) : any
}

type MappingData = {
  path: string,
  getPath?: (values?: PathValues) => string,
  mock?: boolean,
  mockFunction?: MockFunction,
  legacy?: boolean
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE' // 'GET' will be used as default
}

export const mapping: Record<string, MappingData> = {
  [API_PATH.DASHBOARD_VALIDATOR_MANAGEMENT]: {
    path: 'validator-dashboards/{dashboard_id}/validators',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/validators`,
    mock: false
  },
  [API_PATH.DASHBOARD_VALIDATOR_BLOCKS]: {
    path: 'validator-dashboards/{dashboard_id}/blocks',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/blocks`,
    mock: false
  },
  [API_PATH.DASHBOARD_VALIDATOR_WITHDRAWALS]: {
    path: 'validator-dashboards/{dashboard_id}/withdrawals',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/withdrawals`,
    mock: false
  },
  [API_PATH.DASHBOARD_VALIDATOR_TOTAL_WITHDRAWALS]: {
    path: 'validator-dashboards/{dashboard_id}/total-withdrawals',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/total-withdrawals`,
    mock: false
  },
  [API_PATH.DASHBOARD_VALIDATOR_GROUPS]: {
    path: 'validator-dashboards/{dashboard_id}/groups',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/groups`,
    mock: false,
    method: 'POST'
  },
  [API_PATH.DASHBOARD_VALIDATOR_GROUP_MODIFY]: {
    path: 'validator-dashboards/{dashboard_id}/groups/{group_id}',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/groups/${values?.groupId}`,
    mock: false,
    method: 'PUT' // can be 'DELETE' = delete group or 'PUT' = modify group
  },
  [API_PATH.AD_CONFIGURATIONs]: {
    path: '/ad-configurations?={keys}',
    getPath: values => `/ad-configurations?keys=${values?.keys}`,
    mock: true
  },
  [API_PATH.USER]: {
    path: '/users/me',
    mock: false
  },
  [API_PATH.USER_DASHBOARDS]: {
    path: '/users/me/dashboards',
    mock: false
  },
  [API_PATH.DASHBOARD_CREATE_ACCOUNT]: {
    path: '/account-dashboards',
    mock: true,
    method: 'POST'
  },
  [API_PATH.DASHBOARD_CREATE_VALIDATOR]: {
    path: '/validator-dashboards',
    mock: false,
    method: 'POST'
  },
  [API_PATH.DASHBOARD_DELETE_ACCOUNT]: {
    path: '/account-dashboards/{dashboardKey}',
    getPath: values => `/account-dashboards/${values?.dashboardKey}`,
    mock: true,
    method: 'DELETE'
  },
  [API_PATH.DASHBOARD_DELETE_VALIDATOR]: {
    path: '/validator-dashboards/{dashboardKey}',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}`,
    mock: false,
    method: 'DELETE'
  },
  [API_PATH.DASHBOARD_RENAME_ACCOUNT]: {
    path: '/account-dashboards/{dashboardKey}/name',
    getPath: values => `/account-dashboards/${values?.dashboardKey}/name`,
    mock: true,
    method: 'PUT'
  },
  [API_PATH.DASHBOARD_RENAME_VALIDATOR]: {
    path: '/validator-dashboards/{dashboardKey}/name',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/name`,
    mock: false,
    method: 'PUT'
  },
  [API_PATH.DASHBOARD_VALIDATOR_CREATE_PUBLIC_ID]: {
    path: '/validator-dashboards/{dashboardKey}/public-ids',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/public-ids`,
    mock: false,
    method: 'POST'
  },
  [API_PATH.DASHBOARD_VALIDATOR_EDIT_PUBLIC_ID]: {
    path: '/validator-dashboards/{dashboardKey}/public-ids/{publicId}',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/public-ids/${values?.publicId}`,
    mock: false,
    method: 'PUT'
  },
  [API_PATH.DASHBOARD_SUMMARY_DETAILS]: {
    path: '/validator-dashboards/{dashboardKey}/groups/{group_id}/summary',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/groups/${values?.groupId}/summary`,
    mock: false
  },
  [API_PATH.DASHBOARD_SUMMARY]: {
    path: '/validator-dashboards/{dashboardKey}/summary',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/summary`,
    mock: false
  },
  [API_PATH.DASHBOARD_VALIDATOR_REWARDS_CHART]: {
    path: '/validator-dashboards/{dashboardKey}/rewards-chart',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/rewards-chart`,
    mock: false
  },
  [API_PATH.DASHBOARD_VALIDATOR_REWARDS_DETAILS]: {
    path: '/validator-dashboards/{dashboardKey}/groups/{group_id}/rewards',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/groups/${values?.groupId}/rewards/${values?.epoch}`,
    mock: false
  },
  [API_PATH.DASHBOARD_VALIDATOR_REWARDS]: {
    path: '/validator-dashboards/{dashboardKey}/rewards',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/rewards`,
    mock: false
  },
  [API_PATH.DASHBOARD_EL_DEPOSITS]: {
    path: '/validator-dashboards/{dashboard_id}/execution-layer-deposits',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/execution-layer-deposits`,
    mock: false
  },
  [API_PATH.DASHBOARD_EL_DEPOSITS_TOTAL]: {
    path: '/validator-dashboards/{dashboard_id}/total-execution-layer-deposits',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/total-execution-layer-deposits`,
    mock: false
  },
  [API_PATH.DASHBOARD_CL_DEPOSITS]: {
    path: '/validator-dashboards/{dashboard_id}/consensus-layer-deposits',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/consensus-layer-deposits`,
    mock: false
  },
  [API_PATH.DASHBOARD_CL_DEPOSITS_TOTAL]: {
    path: '/validator-dashboards/{dashboard_id}/total-consensus-layer-deposits',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/total-consensus-layer-deposits`,
    mock: false
  },
  [API_PATH.DASHBOARD_SUMMARY_CHART]: {
    path: '/validator-dashboards/{dashboardKey}/summary-chart?',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/summary-chart`,
    mock: false
  },
  [API_PATH.DASHBOARD_OVERVIEW]: {
    path: '/validator-dashboards/{dashboardKey}',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}`,
    mock: false
  },
  [API_PATH.DASHBOARD_SLOTVIZ]: {
    path: '/validator-dashboards/{dashboardKey}/slot-viz',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/slot-viz`,
    mock: false
  },
  [API_PATH.DASHBOARD_VALIDATOR_EPOCH_DUTY]: {
    path: '/validator-dashboards/{dashboard_id}/duties/{epoch}:',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/duties/${values?.epoch}`,
    mock: false
  },
  [API_PATH.DASHBOARD_VALIDATOR_INDICES]: {
    path: '/validator-dashboards/{dashboard_id}/validator-indices',
    getPath: values => `/validator-dashboards/${values?.dashboardKey}/validator-indices`,
    mock: false
  },
  [API_PATH.LATEST_STATE]: {
    path: '/latest-state',
    mockFunction: mockLatestState,
    mock: false
  },
  [API_PATH.LOGIN]: {
    path: '/login',
    method: 'POST',
    mock: false
  },
  [API_PATH.LOGOUT]: {
    path: '/logout',
    method: 'POST',
    mock: false
  },
  [API_PATH.SEARCH]: {
    path: '/search',
    method: 'POST',
    mockFunction: simulateAPIresponseForTheSearchBar,
    mock: true
  },
  [API_PATH.AVAILABLE_NETWORKS]: {
    path: '/available-networks',
    method: 'GET',
    mockFunction: simulateAPIresponseAboutNetworkList,
    mock: true
  },
  [API_PATH.PRODUCT_SUMMARY]: {
    path: '/product-summary',
    mock: false
  },
  [API_PATH.STRIPE_CUSTOMER_PORTAL]: {
    path: '/user/stripe/customer-portal',
    method: 'POST',
    mock: false
  },
  [API_PATH.STRIPE_CHECKOUT_SESSION]: {
    path: '/user/stripe/create-checkout-session',
    method: 'POST',
    mock: false
  }
}
