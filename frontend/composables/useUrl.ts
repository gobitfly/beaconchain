export const useUrl = () => {
  const requestUrl = useRequestURL()
  const {
    href: currentUrl,
    origin,
  } = requestUrl
  return {
    currentUrl,
    origin,
  }
}
