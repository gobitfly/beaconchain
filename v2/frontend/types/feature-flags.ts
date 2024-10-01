export type FeatureFlag = (typeof FEATURE_FLAGS)[number]
const FEATURE_FLAGS = [
  'feature-account_dashboards',
  'feature-notifications',
  'feature-user_settings',
] as const
