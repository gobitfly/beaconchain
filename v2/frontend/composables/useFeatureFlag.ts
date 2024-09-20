import { warn } from 'vue'
import type { FeatureFlag } from '~/types/feature-flags'

export const useFeatureFlag = () => {
  type Environment = 'development' | 'production' | 'staging'

  const currentEnvironment = useRuntimeConfig().public.deploymentType as Environment
  if (!currentEnvironment) {
    warn('Environment variable `deploymentType` is not set.')
  }

  const staging: FeatureFlag[] = [ 'feature-notifications' ]
  const development: FeatureFlag[]
  = [
    ...staging,
    'feature-account_dashboards',
    'feature-user_settings',
  ]
  const featureCatalog: Record<Environment, FeatureFlag[]> = {
    development,
    production: [],
    staging,
  }

  const activeFeatures = featureCatalog[currentEnvironment]

  const has = (feature: FeatureFlag) => activeFeatures.includes(feature)

  return {
    activeFeatures,
    has,
  }
}
