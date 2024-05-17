import type { RouteLocationNormalizedLoaded } from 'vue-router'

export default function ({ name, params }:RouteLocationNormalizedLoaded) {
  if (name === 'slug') {
    name = params.slug?.[0]
  }
  switch (name) {
    case 'block':
      return navigateTo(`https://beaconcha.in/block/${params.id || params.slug?.[1]}`, { external: true })
    case 'slot':
      return navigateTo(`https://beaconcha.in/slot/${params.id || params.slug?.[1]}`, { external: true })
    case 'epoch':
      return navigateTo(`https://beaconcha.in/slot/${params.id || params.slug?.[1]}`, { external: true })
    case 'tx':
      return navigateTo(`https://beaconcha.in/tx/${params.id || params.slug?.[1]}`, { external: true })
    case 'address':
      return navigateTo(`https://beaconcha.in/address/${params.id || params.slug?.[1]}`, { external: true })
    case 'validator-id':
    case 'validator':
      return navigateTo(`https://beaconcha.in/validator/${params.id || params.slug?.[1]}`, { external: true })
  }
}
