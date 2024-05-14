// Code generated by tygo. DO NOT EDIT.
/* eslint-disable */
import type { VDBPublicId, ApiDataResponse } from './common'

//////////
// source: dashboard.go

export interface AccountDashboard {
  id: number /* uint64 */;
  name: string;
}
export interface ValidatorDashboard {
  id: number /* uint64 */;
  name: string;
  public_ids: VDBPublicId[];
}
export interface UserDashboardsData {
  validator_dashboards: ValidatorDashboard[];
  account_dashboards: AccountDashboard[];
}
export type GetUserDashboardsResponse = ApiDataResponse<UserDashboardsData>;
