import type { Dashboard } from '../api/dashboard'

// can be ether a dashboard id or a list of hashed validators
export type DashboardKey = string

export type DashboardType = 'validator' | 'account'

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

export interface ExtendedDashboard extends Dashboard{
  hash?: string;
}

/*
export type ExtendedDashboard = {
  id: number ;
  name: string;
  hash?: string
}

*/
