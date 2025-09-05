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
      return redirect('https://github.com/gobitfly/eth2-beaconchain-explorer-app')
    case 'beaconchain-on-discord':
      return redirect('https://dsc.gg/beaconchain')
    case 'beaconchain-on-github':
      return redirect('https://github.com/gobitfly/beaconchain')
    case 'beaconchain-on-x':
      return redirect('https://x.com/beaconcha_in')
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
    case 'privacy':
      return redirect('https://storage.googleapis.com/legal.beaconcha.in/privacy.pdf')
    case 'register':
      return redirectToV1('/register')
    case 'requestReset':
      return redirectToV1('/requestReset')
    case 'shop':
      return redirect('https://shop.beaconcha.in')
    case 'slot':
      return redirectToV1(`/slot/${params.id || params.slug?.[1]}`)
    case 'status':
      return redirect('https://status.beaconcha.in/')
    case 'terms':
      return redirect('https://storage.googleapis.com/legal.beaconcha.in/tos.pdf')
    case 'tx':
      return redirectToV1(`/tx/${params.id || params.slug?.[1]}`)
    case 'user-settings':
      if (!has('feature-user_settings')) {
        return redirectToV1('/user/settings')
      }
      break
    case 'validator':
    case 'validator-id':
      return redirectToV1(`/validator/${params.id || params.slug?.[1]}`)
  }
  const currentEnvironment = config.public.deploymentType
  if (currentEnvironment === 'production' && path.startsWith('/playground')) {
    return abortNavigation()
  }
}
