// Code generated by tygo. DO NOT EDIT.
/* eslint-disable */
import type { ApiDataResponse, ValidatorStateCounts } from './common'

//////////
// source: mobile.go

export interface MobileBundleData {
  bundle_url?: string;
  has_native_update_available: boolean;
}
export type GetMobileLatestBundleResponse = ApiDataResponse<MobileBundleData>;
export interface MobileWidgetData {
  validator_state_counts: ValidatorStateCounts;
  last_24h_income: string /* decimal.Decimal */;
  last_7d_income: string /* decimal.Decimal */;
  last_30d_apr: number /* float64 */;
  last_30d_efficiency: number /* float64 */;
  network_efficiency: number /* float64 */;
  rpl_price: string /* decimal.Decimal */;
  rpl_apr: number /* float64 */;
}
export type InternalGetValidatorDashboardMobileWidgetResponse = ApiDataResponse<MobileWidgetData>;
