import type { NitroFetchOptions } from 'nitropack'
// import { defu } from 'defu'

const mockedPaths:Record<string, boolean> = {
  '/latestState': true
}

export function useCustomFetch<T> (pathName: string, options: NitroFetchOptions<string & {}> = {}): Promise<T> {
  /* const userAuth = useCookie('token')
  const config = useRuntimeConfig()

  const defaults: NitroFetchOptions<string & {}> = {
    baseURL: config.baseUrl ?? 'https://api.nuxt.com',
    // this overrides the default key generation, which includes a hash of
    // url, method, headers, etc. - this should be used with care as the key
    // is how Nuxt decides how responses should be deduplicated between
    // client and server
    key: url,

    // set user token if connected
    headers: userAuth.value
      ? { Authorization: `Bearer ${userAuth.value}` }
      : {},

    onResponse (_ctx) {
      // _ctx.response._data = new myBusinessResponse(_ctx.response._data)
    },

    onResponseError (_ctx) {
      // throw new myBusinessError()
    }
  }

  // for nice deep defaults, please use unjs/defu
  const params = defu(options, defaults) */

  const url = useRequestURL()
  const { public: { apiClientV1 } } = useRuntimeConfig()
  const path = mockedPaths[pathName] ? `${pathName}.json` : pathName
  const baseURL = mockedPaths[pathName] ? process.server ? `${url.protocol}${url.host}/mock` : './mock' : apiClientV1
  return $fetch<T>(path, { ...options, baseURL })
}
