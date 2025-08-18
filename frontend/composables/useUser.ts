export const useUser = () => {
  const user = useFetchedData('user')
  const { clear: clearPrivateDashboards } = usePrivateDashboards()
  const isLoggedIn = computed(() => !!user.value)
  const { $api } = useNuxtApp()
  const clear = () => {
    clearNuxtData('user')
  }
  const logout = async () => {
    await $api('/api/bff/auth/logout', {
      method: 'POST',
    }).then(async () => {
      clear()
      clearPrivateDashboards()
      // await navigateTo({ name: 'dashboard' })
    })
  }

  return {
    // dashboards,
    isLoggedIn,
    logout,
    user,
  }
}
