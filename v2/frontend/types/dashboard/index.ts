// can be either a dashboard id or a list of hashed validators
export type DashboardKey = string

export type DashboardType = 'validator' | 'account'

export enum COOKIE_DASHBOARD_ID{
  VALIDATOR = -2,
  ACCOUNT = -3,
}

// TODO: once the search PR is finished check if we can get these from somewhere else
export type ValidatorDashboardNetwork = 'ethereum' | 'gnosis'

export const DAHSHBOARDS_ALL_GROUPS_ID = -1
export const DAHSHBOARDS_NEXT_EPOCH_ID = -2

export type DashboardKeyData = {
  dashboardType:globalThis.Ref<DashboardType>,
  dashboardKey:globalThis.Ref<string>,
  isPublic:globalThis.Ref<boolean>,
  publicEntities:globalThis.Ref<string[]>,
  addEntities:(list:string[]) =>void,
  removeEntities:(list:string[]) =>void,
  setDashboardKey:(key:string) =>void,
}

// smallest similarites of AccountDashboard and ValidatorDashboard
export interface Dashboard {
  id: number;
  name: string;
}

// For not logged in Users we store the Dashboard in Cookies
export interface CookieDashboard extends Dashboard{
  hash?: string;
}
