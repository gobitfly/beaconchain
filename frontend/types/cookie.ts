export enum COOKIE_KEY {
  COOKIES_PREFERENCE = 'cookies-preference',
  CURRENCY= 'currency',
  REFRESH_TOKEN= 'refresh-token',
  ACCESS_TOKEN= 'access-token',
  VALIDATOR_DASHOBARD_KEY= 'validator-dashboard-key',
  ACCOUNT_DASHOBARD_KEY= 'account-dashboard-key',
  USER_DASHBOARDS= 'user-dashboards',
  SLOT_VIZ_SELECTED_CATEGORIES = 'slot-viz-selected-categories'
}

export type CookiesPreference = 'all' | 'functional' | undefined
