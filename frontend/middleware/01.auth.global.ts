export default defineNuxtRouteMiddleware(async () => {
  // we only call user on server side, since on the client, we do not know if we have a `session_id cookie`
  // if (isClientSide) return
  await useApi('/api/bff/users/me', {
    getCachedData: (key, nuxtApp) => nuxtApp.payload[key] ?? nuxtApp.payload.data[key],
    key: 'user',
  })
})
