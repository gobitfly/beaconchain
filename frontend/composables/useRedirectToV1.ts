const options = {
  external: true,
  redirectCode: 307,
  replace: true,
}
export const useRedirectToV1 = () => {
  const { v1Domain } = useRuntimeConfig().public ?? 'https://beaconcha.in'
  return (path: `/${string}` | `https://${string}`) => {
    const isAbsolutePath = path.startsWith('https://')
    if (isAbsolutePath) {
      return navigateTo(path, options)
    }
    return navigateTo(`${v1Domain}${path}`, options)
  }
}
