export const usePremiumPerks = () => {
  const user = useFetchedData('user')
  const hasShareCustomDashboard = computed(() => {
    return user.value?.premium_perks?.share_custom_dashboards ?? false
  })
  const maxValidatorDashboards = computed(() => {
    return user.value?.premium_perks?.validator_dashboards ?? 0
  })
  return {
    hasShareCustomDashboard,
    maxValidatorDashboards,
  }
}
