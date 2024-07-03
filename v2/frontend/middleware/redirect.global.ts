import type { RouteLocationNormalizedLoaded } from 'vue-router'

export default function ({ name, params }: RouteLocationNormalizedLoaded) {
  const v1Domain = useRuntimeConfig().public.v1Domain || 'https://beaconcha.in'
  const showInDevelopment = Boolean(useRuntimeConfig().public.showInDevelopment)
  if (name === 'slug') {
    name = params.slug?.[0]
  }
  switch (name) {
    case 'block':
      return navigateTo(`${v1Domain}/block/${params.id || params.slug?.[1]}`, {
        external: true
      })
    case 'slot':
      return navigateTo(`${v1Domain}/slot/${params.id || params.slug?.[1]}`, {
        external: true
      })
    case 'epoch':
      return navigateTo(`${v1Domain}/epoch/${params.id || params.slug?.[1]}`, {
        external: true
      })
    case 'tx':
      return navigateTo(`${v1Domain}/tx/${params.id || params.slug?.[1]}`, {
        external: true
      })
    case 'address':
      return navigateTo(
        `${v1Domain}/address/${params.id || params.slug?.[1]}`,
        { external: true }
      )
    case 'validator-id':
    case 'validator':
      return navigateTo(
        `${v1Domain}/validator/${params.id || params.slug?.[1]}`,
        { external: true }
      )
    case 'mobile':
      return navigateTo(`${v1Domain}/mobile`, { external: true })
    case 'user-settings':
      // TODO: Remove once backend for this page is ready
      if (!showInDevelopment) {
        return navigateTo(`${v1Domain}/user/settings`, { external: true })
      }
  }
}
