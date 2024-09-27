/**
 * This implements a custom $fetch, to globally add things like
 *   - baseURL
 *   - add headers, e.g. x-csrf-token, x-ssr-secret
 */

export default defineNuxtPlugin((
  // nuxtApp
) => {
  const {
    apiClient,
  } = useRuntimeConfig().public
  // Todo: replace useCsrfStore() with csrfToken, when useCustomFetch is removed
  // const csrfToken = ref('')
  const APIcallTimeout = 30 * 1000 // 30 seconds
  const {
    setTokenCsrf,
    tokenCsrf,
  } = useCsrfStore()
  const fetchApi = $fetch.create({
    baseURL: !isDevelopmentEnvironment ? apiClient : '',
    credentials: 'include', // include cookies on client side
    method: 'GET',
    onRequest({
      options,
    }) {
      options.headers = new Headers({
        ...options.headers,
        ...useRequestHeaders([ 'cookie' ]),
      })
      if (isServerSide) {
        const { ssrSecret } = useRuntimeConfig().private
        options.headers.append('x-ssr-secret', ssrSecret)
      }
      if (tokenCsrf.value && options.method !== 'GET') {
        options.headers.append('x-csrf-token', tokenCsrf.value)
      }
    },
    onResponse({ response }) {
      // current `csrf token` is retrieved from GET requests
      const newToken = response.headers.get('x-csrf-token')
      if (newToken) {
        // csrfToken.value = newToken
        setTokenCsrf(newToken)
      }
    },
    async onResponseError({ response }) {
      if (response.status === 401) {
        // await nuxtApp.runWithContext(() => navigateTo('/login'))
      }
    },
    signal: AbortSignal.timeout(APIcallTimeout),
  })

  return {
    provide: {
      fetchApi,
    },
  }
})
