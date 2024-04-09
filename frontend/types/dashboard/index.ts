import type { Dashboard } from '../api/dashboard'

// can be ether a dashboard id or a list of hashed validators
export type DashboardKey = string

export type DashboardType = 'validator' | 'account'

export enum COOKIE_DASHBOARD_ID{
  VALIDATOR = -2,
  ACCOUNT = -3,
}

// TODO: once the search PR is finished check if we can get these from somewhere else
export type ValidatorDashboardNetwork = 'ethereum' | 'gnosis'

export const DAHSHBOARDS_ALL_GROUPS_ID = -1

export type DashboardKeyData = {
  dashboardKey:globalThis.Ref<string>,
  isPublic:globalThis.Ref<boolean>,
  publicEntities:globalThis.Ref<string[]>,
  addEntities:(list:string[]) =>void,
  removeEntities:(list:string[]) =>void,
}

// For not logged in Users we store the Dashboard in Cookies
export interface CookieDashboard extends Dashboard{
  hash?: string;
}
