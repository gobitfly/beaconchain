import type { DashboardKey } from '~/types/dashboard'

export const useDataFetcher = <T extends object>(
  fetchFunc: (
    dashboardKey: DashboardKey
  ) => Promise<T | undefined>,
  errFunc: (e: any) => void,
) => {
  const data = ref<T>()
  const isLoading = ref(false)
  const refresh = (dashboardKey: DashboardKey) => {
    isLoading.value = true
    fetchFunc(dashboardKey)
      .then((fetchedData) => {
        data.value = fetchedData
      })
      .catch((e) => {
        errFunc(e)
      })
      .finally(() => {
        isLoading.value = false
      })
  }
  return {
    data,
    isLoading,
    refresh,
  }
}
