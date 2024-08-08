export function useBcSeo(
  pageTitle?:
    | string
    | Ref<string | number | undefined>
    | ComputedRef<string | number | undefined>,
  removeDynamicUrlValue = false,
) {
  const { t: $t } = useTranslation()
  const route = useRoute()
  const { networkInfo } = useNetworkStore()

  const year = new Date().getFullYear()

  const url = 'https://beaconcha.in'
  const logo = `${url}/img/logo.png`
  const ogUrl = () => {
    const value
      = removeDynamicUrlValue
      && Object.values(route.params).find(
        v => !!v && typeof v === 'string' && route.fullPath.endsWith(v),
      )
    const path = value
      ? route.fullPath.substring(
        0,
        route.fullPath.lastIndexOf(value as string) - 1,
      )
      : route.fullPath
    return `${url}${path}`
  }
  // Maybe we want to have page specific description and keywords in the future, but for now we keep it simple
  const description = () => $t('seo.description')
  const keywords = () => $t('seo.keywords')
  const imageAlt = () => $t('seo.image_alt')

  const dynamicTitle = () => {
    const parts: string[] = [$t('seo.title'), 'beaconcha.in', year.toString()]
    if (typeof pageTitle === 'string') {
      parts.splice(0, 0, $t(pageTitle))
    }
    else if (pageTitle?.value) {
      parts.splice(0, 0, `${pageTitle.value}`)
    }
    return (
      networkInfo.value.description
      + ' '
      + networkInfo.value.name
      + ' '
      + parts.join(' - ')
    )
  }

  useSeoMeta({
    title: dynamicTitle,
    description,
    keywords,
    ogTitle: dynamicTitle,
    ogType: 'website',
    ogImage: logo,
    ogImageAlt: imageAlt,
    ogDescription: description,
    ogUrl,
    ogSiteName: 'beaconcha.in',
    twitterCard: 'summary',
    twitterSite: '@etherchain_org',
    twitterTitle: dynamicTitle,
    twitterDescription: description,
    twitterImage: logo,
    twitterImageAlt: imageAlt,
  })
}
