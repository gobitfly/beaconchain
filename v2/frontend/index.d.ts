declare module 'nuxt/schema' {
  interface PublicRuntimeConfig {
    deploymentType?: 'development' | 'production' | 'staging',
  }
}
export {}
