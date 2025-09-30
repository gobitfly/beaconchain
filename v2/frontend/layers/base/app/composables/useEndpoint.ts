export const useApiUrl = () => {
  const config = useRuntimeConfig()
  const {
    apiClientHoodi,
    apiClientMainnet,
  } = config.public

  const hoodiApiUrl = new URL(apiClientHoodi as string)
  const mainnetApiUrl = new URL(apiClientMainnet as string)

  return {
    hoodi: hoodiApiUrl.origin,
    mainnet: mainnetApiUrl.origin,
  }
}
