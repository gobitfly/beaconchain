export const usePremiumPerks = () => {
  const user = useFetchedData('user')
  const hasShareCustomDashboard = computed(() => {
    return user.value?.premium_perks?.share_custom_dashboards ?? false
  })
  return {
    hasShareCustomDashboard,
  }
}
