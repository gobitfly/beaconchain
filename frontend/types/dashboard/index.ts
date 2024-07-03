// can be either a dashboard id or a list of hashed validators
export type DashboardKey = string

export type DashboardType = 'validator' | 'account' | 'notifications'

export enum COOKIE_DASHBOARD_ID{
  VALIDATOR = -2,
  ACCOUNT = -3,
}

export const DAHSHBOARDS_ALL_GROUPS_ID = -1
export const DAHSHBOARDS_NEXT_EPOCH_ID = -2

export type DashboardKeyData = {
  dashboardType:globalThis.Ref<DashboardType>,
  dashboardKey:globalThis.Ref<string>,
  isPublic:globalThis.Ref<boolean>,
  isShared:globalThis.Ref<boolean>,
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
