export const usePrivateDashboards = () => {
  const dashboards = useFetchedData('privateDashboards')
  const validatorDashboards = computed(() => dashboards.value?.validator_dashboards ?? [])
  const refresh = async () => {
    await refreshNuxtData('privateDashboards')
  }
  const clear = () => {
    clearNuxtData('privateDashboards')
  }
  return {
    clear,
    refresh,
    validatorDashboards,
  }
}
