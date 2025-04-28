export const useDashboard = () => {
  type DashboardMode = 'guest' | 'private' | 'shared'
  const dashboardMode = ref<DashboardMode>('guest')
  return {
    dashboardMode,
  }
}
