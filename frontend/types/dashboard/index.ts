export enum GUEST_DASHBOARD_ID {
  ACCOUNT = -3,
  VALIDATOR = -2,
}

// can be either a dashboard id or a list of hashed validators
export type DashboardKey = string

export type DashboardType = 'account' | 'notifications' | 'validator'

export const DAHSHBOARDS_ALL_GROUPS_ID = -1
export const DAHSHBOARDS_NEXT_EPOCH_ID = -2

// smallest similarites of AccountDashboard and ValidatorDashboard
export interface Dashboard {
  id: number,
  name: string,
}

export type DashboardKeyData = {
  addEntities: (list: string[]) => void,
  dashboardKey: globalThis.Ref<string>,
  dashboardType: globalThis.Ref<DashboardType>,
  isGuestDashboard: globalThis.Ref<boolean>,
  isSharedDashboard: globalThis.Ref<boolean>,
  publicEntities: globalThis.Ref<string[]>,
  removeEntities: (list: string[]) => void,
  setDashboardKey: (key: string) => void,
}

// For not logged in Users we store the Dashboard in Cookies
export interface GuestDashboard extends Dashboard {
  key?: string,
}
