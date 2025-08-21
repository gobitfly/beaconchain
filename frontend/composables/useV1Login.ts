export const useV1Login = () => {
  const v1Domain = useV1Domain()
  const route = useRoute()
  const { origin } = useRequestURL()
  const currentUrl = computed(() => `${origin}${route.fullPath}`)

  const url = `${v1Domain}/login?redirect=${encodeURIComponent(currentUrl.value)}`

  const navigateToV1Login = async () => {
    await navigateTo(url, {
      external: true,
    })
  }

  return {
    navigateToV1Login,
    url,
  }
}
