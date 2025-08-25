export const usePrivateDashboards = () => {
  const dashboards = useFetchedData('privateDashboards')
  const validatorDashboards = computed(() => dashboards.value?.validator_dashboards ?? [])
  const totalValidatorDashboards = computed(() => validatorDashboards.value.length)
  const refresh = async () => {
    await refreshNuxtData('privateDashboards')
  }
  const clear = () => {
    clearNuxtData('privateDashboards')
  }
  return {
    clear,
    refresh,
    totalValidatorDashboards,
    validatorDashboards,
  }
}
