import type { RouteLocationNormalizedLoaded } from 'vue-router'

export default function ({
  name,
  params,
  path,
}: RouteLocationNormalizedLoaded) {
  const { has } = useFeatureFlag()
  const config = useRuntimeConfig()
  const v1Domain = config.public.v1Domain || 'https://beaconcha.in'
  if (name === 'slug') {
    name = params.slug?.[0]
  }
  function redirectToV1(path: `/${string}`) {
    return navigateTo(`${v1Domain}${path}`, { external: true })
  }
  function redirect(url: string) {
    return navigateTo(`${url}`, { external: true })
  }
  switch (name) {
    case 'address':
      return redirectToV1(`/address/${params.id || params.slug?.[1]}`)
    case 'beaconchain-dashboard-mobile-app-on-github':
      return redirect(externalLink.beaconchain.github.mobileApp)
    case 'beaconchain-on-discord':
      return redirect(externalLink.beaconchain.discord)
    case 'beaconchain-on-github':
      return redirect(externalLink.beaconchain.github.beaconchain)
    case 'beaconchain-on-x':
      return redirect(externalLink.beaconchain.x)
    case 'block':
      return redirectToV1(`/block/${params.id || params.slug?.[1]}`)
    case 'epoch':
      return redirectToV1(`/epoch/${params.id || params.slug?.[1]}`)
    case 'imprint':
      return redirectToV1('/imprint')
    case 'index':
      return redirectToV1('/')
    case 'mobile':
      return redirectToV1('/mobile')
    case 'p':
    case 'product':
    case 'products':
      if (!has('feature-product-landing')) {
        return abortNavigation()
      }
      break
    case 'privacy':
      return redirect(externalLink.beaconchain.privacyPolicy)
    case 'requestReset':
      return redirectToV1('/requestReset')
    case 'shop':
      return redirect(externalLink.beaconchain.shop)
    case 'slot':
      return redirectToV1(`/slot/${params.id || params.slug?.[1]}`)
    case 'status':
      return redirect(externalLink.beaconchain.status)
    case 'terms':
      return redirect(externalLink.beaconchain.termsOfService)
    case 'tx':
      return redirectToV1(`/tx/${params.id || params.slug?.[1]}`)
    case 'validator':
    case 'validator-id':
      return redirectToV1(`/validator/${params.id || params.slug?.[1]}`)
  }
  const currentEnvironment = config.public.deploymentType
  if (currentEnvironment === 'production' && path.startsWith('/playground')) {
    return abortNavigation()
  }
}
