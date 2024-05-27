export function useBcSeo (pageTitle?: string | Ref<string | number | undefined> | ComputedRef<string | number | undefined>) {
  const { t: $t } = useI18n()

  const year = new Date().getFullYear()

  const ogUrl = 'https://beaconcha.in/'
  const logo = `${ogUrl}/img/logo.png`
  const description = () => $t('seo.description')
  const keywords = () => $t('seo.description')
  const imageAlt = () => $t('seo.image_alt')

  const dynamicTitle = () => {
    const parts: string[] = [$t('seo.title'), 'beaconcha.in', year.toString()]
    if (typeof pageTitle === 'string') {
      parts.splice(0, 0, $t(pageTitle))
    } else if (pageTitle?.value) {
      parts.splice(0, 0, `${pageTitle.value}`)
    }
    return parts.join(' - ')
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
    twitterImageAlt: imageAlt
  })
}
