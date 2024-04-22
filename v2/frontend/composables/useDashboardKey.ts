import { inject } from 'vue'
import type { DashboardKeyData } from '~/types/dashboard'

export function useDashboardKey () {
  const data = inject<DashboardKeyData>('dashboard-key')
  const { isLoggedIn } = useUserStore()

  if (!data) {
    throw new Error('useDashboardKey must be in a child of useDashboardKeyProvider')
  }

  const dashboardKey = computed(() => data.dashboardKey.value ?? '')
  const isPublic = computed(() => !!data.isPublic.value)
  const publicEntities = computed(() => data.publicEntities.value ?? [])
  const isPrivate = computed(() => isLoggedIn.value && !isPublic.value)
  const setDashboardKey = (key: string) => data.setDashboardKey(key)

  return { ...data, dashboardKey, isPublic, isPrivate, publicEntities, setDashboardKey }
}
