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
  const isGuest = computed(() => !!data.isGuest.value)
  const isShared = computed(() => !!data.isShared.value)
  const publicEntities = computed(() => data.publicEntities.value ?? [])
  const isPrivate = computed(() => isLoggedIn.value && !isGuest.value)
  const setDashboardKey = (key: string) => data.setDashboardKey(key)
  const dashboardType = computed(() => data.dashboardType.value)

  return {
    ...data,
    dashboardKey,
    dashboardType,
    isGuest,
    isPrivate,
    isShared,
    publicEntities,
    setDashboardKey,
  }
}
