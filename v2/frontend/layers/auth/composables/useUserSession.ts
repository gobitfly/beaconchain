export const useUserSession = () => {
  const data = useState<Awaited<ReturnType<typeof getUser>>>('user-session')
  const hasSession = computed(() => !!data.value)
  const requestFetch = useRequestFetch()
  const getUser = async () => {
    const result = await requestFetch('/api/bff/users/me')
      .catch((error) => {
        // session_id invalid or expired
        if (error?.status === 401) return
        throw createError({
          statusCode: error?.status || 500,
          statusMessage: error?.statusMessage || 'Unknown error',
        })
      })
    data.value = result
    return result
  }
  return {
    data,
    getUser,
    hasSession,
  }
}
