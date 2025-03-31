import { inject } from 'vue'
import type { DashboardKeyData } from '~/types/dashboard'

export function useDashboardKey() {
  const data = inject<DashboardKeyData>('dashboard-key')
  const { isLoggedIn } = useUserStore()

  if (!data) {
    throw new Error(
      'useDashboardKey must be in a child of useDashboardKeyProvider',
    )
  }

  const dashboardKey = computed(() => data.dashboardKey.value ?? '')
  const isGuestDashboard = computed(() => !!data.isGuestDashboard.value)
  const isSharedDashboard = computed(() => !!data.isSharedDashboard.value)
  const publicEntities = computed(() => data.publicEntities.value ?? [])
  const isPrivateDashboard = computed(() => isLoggedIn.value && !isGuestDashboard.value)
  const setDashboardKey = (key: string) => data.setDashboardKey(key)
  const dashboardType = computed(() => data.dashboardType.value)

  return {
    ...data,
    dashboardKey,
    dashboardType,
    isGuestDashboard,
    isPrivateDashboard,
    isSharedDashboard,
    publicEntities,
    setDashboardKey,
  }
}
