// Code generated by tygo. DO NOT EDIT.
/* eslint-disable */
import type { ChartHistorySeconds, ApiDataResponse } from './common'

//////////
// source: user.go

export const UserGroupAdmin = "ADMIN";
export interface UserInfo {
  id: number /* uint64 */;
  email: string;
  api_keys: string[];
  api_perks: ApiPerks;
  premium_perks: PremiumPerks;
  subscriptions: UserSubscription[];
}
export interface UserSubscription {
  product_id: string;
  product_name: string;
  product_category: ProductCategory;
  product_store: ProductStore;
  start: number /* int64 */;
  end: number /* int64 */;
}
export type InternalGetUserInfoResponse = ApiDataResponse<UserInfo>;
export interface EmailUpdate {
  id: number /* uint64 */;
  current_email: string;
  pending_email: string;
}
export type InternalPostUserEmailResponse = ApiDataResponse<EmailUpdate>;
export interface AdConfigurationUpdateData {
  jquery_selector: string;
  insert_mode: string;
  refresh_interval: number /* uint64 */;
  for_all_users: boolean;
  banner_id: number /* uint64 */;
  html_content: string;
  enabled: boolean;
}
export interface AdConfigurationData {
  key: string;
  AdConfigurationUpdateData?: AdConfigurationUpdateData;
}
export type ProductCategory = string;
export const ProductCategoryApi: ProductCategory = "api";
export const ProductCategoryPremium: ProductCategory = "premium";
export const ProductCategoryPremiumAddon: ProductCategory = "premium_addon";
export type ProductStore = string;
export const ProductStoreStripe: ProductStore = "stripe";
export const ProductStoreIosAppstore: ProductStore = "ios-appstore";
export const ProductStoreAndroidPlaystore: ProductStore = "android-playstore";
export const ProductStoreEthpool: ProductStore = "ethpool";
export const ProductStoreCustom: ProductStore = "custom";
export interface ProductSummary {
  validators_per_dashboard_limit: number /* uint64 */;
  stripe_public_key: string;
  api_products: ApiProduct[];
  premium_products: PremiumProduct[];
  extra_dashboard_validators_premium_addons: ExtraDashboardValidatorsPremiumAddon[];
}
export type InternalGetProductSummaryResponse = ApiDataResponse<ProductSummary>;
export interface ApiProduct {
  product_id: string;
  product_name: string;
  api_perks: ApiPerks;
  price_per_year_eur: number /* float64 */;
  price_per_month_eur: number /* float64 */;
  is_popular: boolean;
  stripe_price_id_monthly: string;
  stripe_price_id_yearly: string;
}
export interface ApiPerks {
  units_per_second: number /* uint64 */;
  units_per_month: number /* uint64 */;
  api_keys: number /* uint64 */;
  consensus_layer_api: boolean;
  execution_layer_api: boolean;
  layer2_api: boolean;
  no_ads: boolean; // note that this is somhow redunant, since there is already PremiumPerks.AdFree
  discord_support: boolean;
}
export interface PremiumProduct {
  product_name: string;
  premium_perks: PremiumPerks;
  price_per_year_eur: number /* float64 */;
  price_per_month_eur: number /* float64 */;
  is_popular: boolean;
  product_id_monthly: string;
  product_id_yearly: string;
  stripe_price_id_monthly: string;
  stripe_price_id_yearly: string;
}
export interface ExtraDashboardValidatorsPremiumAddon {
  product_name: string;
  extra_dashboard_validators: number /* uint64 */;
  price_per_year_eur: number /* float64 */;
  price_per_month_eur: number /* float64 */;
  product_id_monthly: string;
  product_id_yearly: string;
  stripe_price_id_monthly: string;
  stripe_price_id_yearly: string;
}
export interface PremiumPerks {
  ad_free: boolean; // note that this is somhow redunant, since there is already ApiPerks.NoAds
  validator_dashboards: number /* uint64 */;
  validators_per_dashboard: number /* uint64 */;
  validator_groups_per_dashboard: number /* uint64 */;
  share_custom_dashboards: boolean;
  manage_dashboard_via_api: boolean;
  bulk_adding: boolean;
  chart_history_seconds: ChartHistorySeconds;
  email_notifications_per_day: number /* uint64 */;
  configure_notifications_via_api: boolean;
  validator_group_notifications: number /* uint64 */;
  webhook_endpoints: number /* uint64 */;
  mobile_app_custom_themes: boolean;
  mobile_app_widget: boolean;
  monitor_machines: number /* uint64 */;
  machine_monitoring_history_seconds: number /* uint64 */;
  custom_machine_alerts: boolean;
}
export interface StripeCreateCheckoutSession {
  sessionId?: string;
  error?: string;
}
export interface StripeCustomerPortal {
  url: string;
}
export interface OAuthAppData {
  ID: number /* uint64 */;
  Owner: number /* uint64 */;
  AppName: string;
  RedirectURI: string;
  Active: boolean;
}
