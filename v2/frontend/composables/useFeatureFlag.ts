export type FeatureFlag
  = | 'feature-account_dashboards'
    | 'feature-product-landing'

export const useFeatureFlag = () => {
  const currentEnvironment = useRuntimeConfig().public.deploymentType
  if (!currentEnvironment) {
    throw createError('Environment variable `deploymentType` is not provided.')
  }

  const staging: FeatureFlag[] = [ 'feature-product-landing' ]
  const development: FeatureFlag[] = [
    ...staging,
    'feature-account_dashboards',
  ]
  const featureCatalog: Record<typeof currentEnvironment, FeatureFlag[]> = {
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
