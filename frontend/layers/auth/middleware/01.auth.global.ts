export default defineNuxtRouteMiddleware(async () => {
  const {
    getUser,
    hasSession,
  } = useUserSession()
  if (!hasSession.value) {
    await callOnce(async () => {
      await getUser()
    })
  }
})
