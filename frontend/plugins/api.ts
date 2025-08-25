export default defineNuxtPlugin(() => {
  const api = $fetch.create({
    headers: {
      cookie: useRequestHeader('cookie') ?? '',
    },
    // onRequest({
    //   options,
    // }) {
    //   if (isServerSide) {
    //     const cookiesForServerSideRequests = useRequestHeader('cookie')
    //     if (cookiesForServerSideRequests) {
    //       options.headers.append('cookie', cookiesForServerSideRequests)
    //     }
    //   }
    //   // SSR-aware example: add cookies or auth from context
    //     cookie: nuxtApp.ssrContext?.event?.req?.headers.cookie,
    // },
    // async onResponseError({
    //   response,
    // }) {
    //   const redirectTo = encodeURIComponent(useRoute().fullPath)
    //   if (response.statusText === 'Unauthorized') {
    //     // TODO prevent infinite loop
    //     await nuxtApp.runWithContext(() => navigateTo({
    //       name: 'login',
    //       query: {
    //         redirectTo,
    //       },
    //     }))
    //   }
    // },
  })

  return {
    provide: {
      api,
    },
  }
})
