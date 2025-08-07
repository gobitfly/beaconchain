export const useLoginUrl = () => {
  const v1Domain = useV1Domain()
  const route = useRoute()
  const { origin } = useRequestURL()
  const currentUrl = computed(() => `${origin}${route.fullPath}`)
  return `${v1Domain}/login?redirect=${encodeURIComponent(currentUrl.value)}`
}
