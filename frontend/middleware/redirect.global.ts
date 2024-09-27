import type { RouteLocationNormalizedLoaded } from 'vue-router'

export default function ({
  name,
  params,
  query,
}: RouteLocationNormalizedLoaded) {
  const { has } = useFeatureFlag()
  const config = useRuntimeConfig()
  const v1Domain = config.public.v1Domain || 'https://beaconcha.in'
  if (name === 'slug') {
    name = params.slug?.[0]
  }
  switch (name) {
    case 'address':
      return navigateTo(
        `${v1Domain}/address/${params.id || params.slug?.[1]}`,
        { external: true },
      )
    case 'block':
      return navigateTo(`${v1Domain}/block/${params.id || params.slug?.[1]}`, { external: true })
    case 'dashboard':
    case 'dashboard-id':
      if (query.validators && typeof query.validators === 'string') {
        const list = query.validators
          .split(',')
          .filter((v) => {
            return isInt(v) || isPublicKey(v)
          })
          .slice(0, 20)
          .join(',')
        if (list.length) {
          const hash = toBase64Url(list)
          return navigateTo(`/dashboard/${hash}`)
        }
      }
      break
    case 'epoch':
      return navigateTo(`${v1Domain}/epoch/${params.id || params.slug?.[1]}`, { external: true })
    case 'mobile':
      return navigateTo(`${v1Domain}/mobile`, { external: true })
    case 'notifications':
      if (!has('feature-notifications')) {
        return navigateTo(`${v1Domain}/user/notifications`)
      }
      break
    case 'requestReset':
      return navigateTo(`${v1Domain}/requestReset`, { external: true })
    case 'slot':
      return navigateTo(`${v1Domain}/slot/${params.id || params.slug?.[1]}`, { external: true })
    case 'tx':
      return navigateTo(`${v1Domain}/tx/${params.id || params.slug?.[1]}`, { external: true })
    case 'user-settings':
      if (!has('feature-user_settings')) {
        return navigateTo(`${v1Domain}/user/settings`)
      }
      break
    case 'validator':
    case 'validator-id':
      return navigateTo(
        `${v1Domain}/validator/${params.id || params.slug?.[1]}`,
        { external: true },
      )
  }
}
