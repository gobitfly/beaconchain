export function useUserStore() {
  const { data } = useUserSession()
  const { fetch } = useCustomFetch()
  const router = useRouter()

  const doLogout = async () => {
    await fetch('LOGOUT')
    data.value = undefined
    router.replace('/')
  }

  const user = computed(() => {
    return data.value
  })

  const hasV1Notifications = computed(() => {
    return !!user.value?.has_v1_notifications
  })

  const isLoggedIn = computed(() => {
    return !!user.value
  })

  const premium_perks = computed(() => user.value?.premium_perks)

  return {
    doLogout,
    hasV1Notifications,
    isLoggedIn,
    premium_perks,
    user,
  }
}
