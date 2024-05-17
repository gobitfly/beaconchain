import type { RouteLocationNormalizedLoaded } from 'vue-router'

export default function ({ name, params }: RouteLocationNormalizedLoaded) {
  const v1Domain = useRuntimeConfig().public.v1Domain || 'https://beaconcha.in'
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
      return navigateTo(`${v1Domain}/slot/${params.id || params.slug?.[1]}`, {
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
  }
}
