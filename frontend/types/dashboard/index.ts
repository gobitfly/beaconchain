// can be either a dashboard id or a list of hashed validators
export type DashboardKey = string

export type DashboardType = 'account' | 'notifications' | 'validator'

export enum COOKIE_DASHBOARD_ID {
  ACCOUNT = -3,
  VALIDATOR = -2,
}

export const DAHSHBOARDS_ALL_GROUPS_ID = -1
export const DAHSHBOARDS_NEXT_EPOCH_ID = -2

export type DashboardKeyData = {
  addEntities: (list: string[]) => void
  dashboardKey: globalThis.Ref<string>
  dashboardType: globalThis.Ref<DashboardType>
  isPublic: globalThis.Ref<boolean>
  isShared: globalThis.Ref<boolean>
  publicEntities: globalThis.Ref<string[]>
  removeEntities: (list: string[]) => void
  setDashboardKey: (key: string) => void
}

// smallest similarites of AccountDashboard and ValidatorDashboard
export interface Dashboard {
  id: number
  name: string
}

// For not logged in Users we store the Dashboard in Cookies
export interface CookieDashboard extends Dashboard {
  hash?: string
}
