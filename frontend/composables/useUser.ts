export const useUser = () => {
  const user = useFetchedData('user')
  const isLoggedIn = computed(() => !!user.value)
  const { $api } = useNuxtApp()
  const logout = async () => {
    await $api('/api/auth/logout', {
      method: 'POST',
    }).then(() => {
      user.value = null
      navigateTo({ name: 'dashboard' })
    })
  }
  return {
    isLoggedIn,
    logout,
  }
}
